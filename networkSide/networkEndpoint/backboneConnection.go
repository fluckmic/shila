package networkEndpoint

import (
	"encoding/gob"
	"fmt"
	"github.com/scionproto/scion/go/lib/snet"
	"io"
	"shila/core/shila"
	"shila/log"
	"shila/networkSide/network"
	"sync"
)

type BackboneConnectionsMapping map[shila.NetworkAddressKey] *BackboneConnection

type BackboneConnections struct {
	connections 	BackboneConnectionsMapping
	server			*Server
	lock			sync.Mutex
}

func NewBackboneConnections(server *Server) BackboneConnections {
	return BackboneConnections{
		connections: 	make(BackboneConnectionsMapping),
		server:			server,
	}
}

func (conns *BackboneConnections) retrieve(key shila.NetworkAddressKey) *BackboneConnection {
	if conn, ok := conns.connections[key]; ok {
		return conn
	} else {
		return nil
	}
}

func (conns *BackboneConnections) remove(key shila.NetworkAddressKey) {
	conns.lock.Lock()
	defer conns.lock.Unlock()
	delete(conns.connections, key)
}

func (conns *BackboneConnections) add(key shila.NetworkAddressKey, conn *BackboneConnection) {
	conns.lock.Lock()
	defer conns.lock.Unlock()
	conns.connections[key] = conn
	return
}

func (conns *BackboneConnections) TearDown() error {
	return nil
}

func (conns *BackboneConnections) WriteIngress(rAddress shila.NetworkAddress, buff []byte) {

	conns.lock.Lock()
	defer conns.lock.Unlock()

	conn := conns.retrieve(shila.GetNetworkAddressKey(rAddress))
	if conn == nil {
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

func (conns *BackboneConnections) WriteEgress(packet *shila.Packet) error {

	conns.lock.Lock()
	defer conns.lock.Unlock()

	// We try to send out the data if there exists a backbone connection.
	if conn := conns.retrieve(shila.GetNetworkAddressKey(packet.Flow.NetFlow.Dst)); conn != nil {
		return conn.writeEgress(packet.Payload)		// If writing fails, then because of an issue with the connection.
	}

	// If there is no connection, meaning that there was no incoming traffic, then we put the packet into the
	// waiting area. It may take some time until the client on the other side is ready.
	conns.server.addToHoldingArea(packet)

	return nil
}

type BackboneConnection struct {
	keys        	[] shila.NetworkAddressKey
	netFlows    	NetFlows
	server			*Server
	ipFlow      	shila.IPFlow
	inReader    	*io.PipeReader
	inWriter    	*io.PipeWriter
	connections 	*BackboneConnections
	lock        	sync.Mutex
}

type NetFlows struct {
	effective   shila.NetFlow
	represented shila.NetFlow
}

func newBackboneConnection(rAddress shila.NetworkAddress, conns *BackboneConnections) *BackboneConnection {

	path, _ := network.PathGenerator{}.New("")		// FIXME: path!
	netFlow := shila.NetFlow{							// Net flow which is represented by this connection.
		Src:  conns.server.lAddress.(*snet.UDPAddr),
		Path: path,
		Dst:  rAddress.(*snet.UDPAddr),
	}

	inReader, inWriter := io.Pipe()

	conn := &BackboneConnection{
		keys:		 	make([] shila.NetworkAddressKey, 2) ,
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

	return conn
}

func (conn *BackboneConnection) decodeIngress() {

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

func (conn *BackboneConnection) retrieveControlMessage() (ctrlMsg controlMessage, err error) {
	if err = gob.NewDecoder(conn.inReader).Decode(&ctrlMsg); err != nil {
		err = shila.PrependError(shila.ParsingError("Failed to decode control message."), err.Error())
	}
	return
}

func (conn *BackboneConnection) processControlMessage(ctrlMsg controlMessage) error {

	// We catch any possible escalating error in the processing and just return a normal
	// error. This error leads to a removal of the connection without terminating the whole shila.
	defer func() error {
		if err := recover(); err != nil {
			parsingError := shila.ParsingError("Cannot process control message.")
			if err, ok := err.(error); ok {
				return shila.PrependError(parsingError, err.Error())
			} else {
				return parsingError
			}
		}
		return nil
	}()

	// Set the ip flow
	conn.ipFlow = ctrlMsg.IPFlow.Swap()

	// If the backbone connection is part of a contact server network endpoint, then the connection
	// has to calculate the lAddress (w.r.t. the host) of the corresponding traffic endpoint.
	// The representing flow is then updated accordingly such that the payload received through
	// this connection is perceived as received through the traffic connection.
	if conn.server.Role() == shila.ContactingNetworkEndpoint {
		conn.netFlows.represented.Src.(*snet.UDPAddr).Host.Port = conn.ipFlow.Src.Port
	}

	// If the backbone connect is part of a traffic serer network endpoint, then the connection
	// has to be found by messages send along the corresponding contact backbone connection as well.
	if conn.server.Role() == shila.TrafficNetworkEndpoint {
		conn.keys = append(conn.keys, shila.GetNetworkAddressKey(&ctrlMsg.SrcAddrContactEndpoint))
		conn.connections.add(conn.keys[1], conn)
	}

	return nil
}

func (conn *BackboneConnection) processPayloadMessage() error {

	// Fetch the next payload message
	var pyldMsg payloadMessage
	if err := gob.NewDecoder(conn.inReader).Decode(&pyldMsg); err != nil {
		err = shila.PrependError(shila.ParsingError("Failed to decode payload message."), err.Error())
	}

	conn.server.ingress <- shila.NewPacketWithNetFlow(conn.server,
													  conn.ipFlow.Swap(),
													  conn.netFlows.represented.Swap(),
													  pyldMsg.Payload)

	return nil
}

func (conn *BackboneConnection) removeConnection() {
	for _, key := range conn.keys {
		conn.connections.remove(key)
	}
}

func (conn *BackboneConnection) writeIngress(buff []byte) (err error) {
	_, err = conn.inWriter.Write(buff)
	return
}

func (conn *BackboneConnection) writeEgress(buff []byte) (err error){
	_, err = conn.server.lConnection.WriteTo(buff, conn.netFlows.effective.Dst.(*snet.UDPAddr))
	return
}

func (conn *BackboneConnection) Identifier() string {
	return fmt.Sprint("{ Backbone connection in ", conn.server.Role(), " Server - ", conn.server.lAddress, " <- ",
		conn.netFlows.effective.Dst.String(), "}")
}

func (conn *BackboneConnection) Says(str string) string {
	return  fmt.Sprint(conn.Identifier(), ": ", str)
}