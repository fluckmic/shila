package networkEndpoint

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/scionproto/scion/go/lib/snet"
	"io"
	"shila/core/shila"
	"shila/log"
	"sync"
)

type ServerBackboneConnectionsMapping map[shila.NetworkAddressKey] *ServerBackboneConnection

type ServerBackboneConnections struct {
	connections ServerBackboneConnectionsMapping
	server      *Server
	lock        sync.Mutex
}

func NewBackboneConnections(server *Server) ServerBackboneConnections {
	return ServerBackboneConnections{
		connections: 	make(ServerBackboneConnectionsMapping),
		server:			server,
	}
}

func (conns *ServerBackboneConnections) retrieve(key shila.NetworkAddressKey) *ServerBackboneConnection {
	conns.lock.Lock()
	defer conns.lock.Unlock()
	if conn, ok := conns.connections[key]; ok {
		return conn
	} else {
		return nil
	}
}

func (conns *ServerBackboneConnections) remove(key shila.NetworkAddressKey) {
	conns.lock.Lock()
	defer conns.lock.Unlock()
	delete(conns.connections, key)
}

func (conns *ServerBackboneConnections) add(key shila.NetworkAddressKey, conn *ServerBackboneConnection) {
	conns.lock.Lock()
	defer conns.lock.Unlock()
	conns.connections[key] = conn
	return
}

func (conns *ServerBackboneConnections) TearDown() error {
	return nil
}

func (conns *ServerBackboneConnections) WriteIngress(rAddress shila.NetworkAddress, buff []byte) {

	conn := conns.retrieve(shila.GetNetworkAddressKey(rAddress))
	if conn == nil {
		log.Verbose.Print("About to create a new Backbone connection for ", rAddress)
		// Connection not yet exists, we first have to create a new one and add it to the mapping.
		if conn = newBackboneConnection(rAddress, conns); conn == nil {
			log.Error.Println(conn.server.Says("Failed to create a new backbone connection."))
			return
		}
		conns.add(conn.keys[0], conn)
	}

	if err := conn.writeIngress(buff); err != nil {
		log.Error.Println(conn.Says(err.Error()))
	}
	return
}

func (conns *ServerBackboneConnections) WriteEgress(packet *shila.Packet) error {

	// We try to send out the data if there exists a backbone connection.
	if conn := conns.retrieve(shila.GetNetworkAddressKey(packet.Flow.NetFlow.Dst)); conn != nil {
		return conn.writeEgress(packet.Payload)		// If writing fails, then because of an issue with the connection.
	}

	// If there is no connection, meaning that there was no incoming traffic, then we put the packet into the
	// waiting area. It may take some time until the client on the other side is ready.
	conns.server.addToHoldingArea(packet)

	return nil
}

type NetFlows struct {
	effective	shila.NetFlow
	represented	shila.NetFlow
}

type ServerBackboneConnection struct {
	keys        	[] shila.NetworkAddressKey
	netFlows    	NetFlows
	server			*Server
	ipFlow      	shila.IPFlow
	inReader    	*io.PipeReader
	inWriter    	*io.PipeWriter
	connections 	*ServerBackboneConnections
	lock        	sync.Mutex
}

func newBackboneConnection(rAddress shila.NetworkAddress, conns *ServerBackboneConnections) *ServerBackboneConnection {

	netFlow := shila.NetFlow{							// Net flow which is represented by this connection.
		Src:  conns.server.lAddress.(*snet.UDPAddr),
		Path: rAddress.(*snet.UDPAddr).Path,
		Dst:  rAddress.(*snet.UDPAddr),
	}

	inReader, inWriter := io.Pipe()

	conn := &ServerBackboneConnection{
		keys:		 	make([] shila.NetworkAddressKey, 0, 2) ,
		netFlows:	 	NetFlows{effective: netFlow, represented: netFlow},
		server:			conns.server,
		inReader:    	inReader,
		inWriter:    	inWriter,
		connections: 	conns,
	}

	conn.keys = append(conn.keys, shila.GetNetworkAddressKey(rAddress))

	go conn.decodeIngress()		// Start the decoder.
								// If there is an issue in the decoding process then the process removes
								// the connection from the mapping and terminates.

								// Packets are no longer send out once the connection is removed. This leads to a
								// timeout issue send back to the shila connection handler which closes the corresponding
								// connection.

	log.Verbose.Println(conn.Says("Created."))
	return conn
}

func (conn *ServerBackboneConnection) decodeIngress() {

	// The first message should be a control message which contains
	// all the information necessary to setup the backbone connection.
	ctrlMsg, err := conn.retrieveControlMessage()
	if err != nil {
		log.Error.Println(conn.Says(err.Error()))
		conn.removeConnection()
		return
	}

	// Process the control message.
	if err := conn.processControlMessage(ctrlMsg); err != nil {
		log.Error.Println(conn.Says(err.Error()))
		conn.removeConnection()
		return
	}

	// Now we are ready to listen for and process payload.
	for {
		if err := conn.processPayloadMessage(); err != nil {
			log.Error.Println(conn.Says(err.Error()))
			conn.removeConnection()
			return
		}
	}
}

func (conn *ServerBackboneConnection) retrieveControlMessage() (ctrlMsg controlMessage, err error) {
	if err = gob.NewDecoder(conn.inReader).Decode(&ctrlMsg); err != nil {
		err = shila.PrependError(ParsingError("Failed to decode control message."), err.Error())
	}
	return
}

func (conn *ServerBackboneConnection) processControlMessage(ctrlMsg controlMessage) error {

	// We catch any possible escalating error in the processing and just return a normal
	// error. This error leads to a removal of the connection without terminating the whole shila.
	defer func() error {
		if err := recover(); err != nil {
			parsingError := ParsingError("Cannot process control message.")
			if err, ok := err.(error); ok {
				return shila.PrependError(parsingError, err.Error())
			} else {
				return parsingError
			}
		}
		return nil
	}()

	// Set the ip flow
	conn.ipFlow   = ctrlMsg.IPFlow.Swap()

	// If the backbone connection is part of a contact server network endpoint, then the connection
	// has to calculate the lAddress (w.r.t. the host) of the corresponding traffic endpoint.
	// The representing flow is then updated accordingly such that the payload received through
	// this connection is perceived as received through the traffic connection.
	if conn.server.Role() == shila.ContactNetworkEndpoint {
		conn.netFlows.represented.Src.(*snet.UDPAddr).Host.Port = conn.ipFlow.Src.Port
	}

	// If the backbone connect is part of a traffic serer network endpoint, then the connection
	// has to be found by messages send along the corresponding contact backbone connection as well.
	if conn.server.Role() == shila.TrafficNetworkEndpoint {

		lAddrContactEndFull 	:= conn.netFlows.effective.Dst.(*snet.UDPAddr).Copy()
		lAddrContactEndFull.Host = &ctrlMsg.LAddrContactEnd
		conn.keys = append(conn.keys, shila.GetNetworkAddressKey(lAddrContactEndFull))
		conn.connections.add(conn.keys[1], conn)
	}

	return nil
}

func (conn *ServerBackboneConnection) processPayloadMessage() error {

	// Fetch the next payload message
	var pyldMsg payloadMessage
	if err := gob.NewDecoder(conn.inReader).Decode(&pyldMsg); err != nil {
		err = shila.PrependError(ParsingError("Failed to decode payload message."), err.Error())
	}
	if len(pyldMsg.Payload) == 0 {
		// From time to to we get a zero payload packet...?
		return nil
	}

	conn.server.Ingress <- shila.NewPacketWithNetFlowAndKind(conn.server,
													  		 conn.ipFlow.Swap(),
													  		 conn.netFlows.represented.Swap(),
													  		 pyldMsg.Payload)

	return nil
}

func (conn *ServerBackboneConnection) removeConnection() {
	for _, key := range conn.keys {
		conn.connections.remove(key)
	}
}

func (conn *ServerBackboneConnection) writeIngress(buff []byte) (err error) {
	_, err = conn.inWriter.Write(buff)
	return
}

func (conn *ServerBackboneConnection) writeEgress(payload []byte) (err error){

	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(payloadMessage{ Payload: payload }); err != nil {
		return shila.PrependError(err, "Cannot encode payload message.")
	}

	_, err = conn.server.lConnection.WriteTo(buffer.Bytes(), conn.netFlows.effective.Dst.(*snet.UDPAddr))
	return
}

func (conn *ServerBackboneConnection) Identifier() string {
	return fmt.Sprint("Backbone connection in Server ", conn.server.Role(), " (", conn.server.lAddress, " <- ",
		conn.netFlows.effective.Dst.String(), ")")
}

func (conn *ServerBackboneConnection) Says(str string) string {
	return  fmt.Sprint(conn.Identifier(), ": ", str)
}