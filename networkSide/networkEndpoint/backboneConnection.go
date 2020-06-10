package networkEndpoint

import (
	"encoding/gob"
	"github.com/scionproto/scion/go/lib/snet"
	"io"
	"shila/core/shila"
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

	// Connection already exists, just write ingress and return.
	if conn := conns.retrieve(shila.GetNetworkAddressKey(rAddress)); conn != nil {
		conn.writeIngress(buff)
		return
	}

	// Connection not yet exists, we first have to create a new one and add it to the mapping.
	if conn, err := newBackboneConnection(rAddress, conns); err == nil {
		return	// Just silently return if we fail, for whatever reason, to create a new backbone connection.
	} else {
		conns.add(conn.keys[0], conn)
		conn.writeIngress(buff)
		return
	}
}

func (conns *BackboneConnections) WriteEgress(to shila.NetworkAddress, buff []byte)  {

	conns.lock.Lock()
	defer conns.lock.Unlock()

	if conn := conns.retrieve(shila.GetNetworkAddressKey(to)); conn != nil {
		conn.writeEgress(buff)
		return
	}

	/*
		if con, ok := server.backboneConnections[dstKey]; ok {
			writer := io.Writer(con.Backbone)
			_, err := writer.Write(p.Payload)
			if err != nil && server.state.Not(shila.Running) {
				return										// Server turned down anyway.
			}
			// If just the connection was closed, then we ignore the error and drop the packet. The ingress handler
			// has observed the issue as well and will sooner or later remove the closed (or faulty) connection.
		} else {
			// Currently there is no backbone connection available to send the packet, put packet into holding area.
			server.lock.Lock()
			server.holdingArea = append(server.holdingArea, p)
			server.lock.Unlock()
		}
		}
	*/
}

type BackboneConnection struct {
	keys        [] shila.NetworkAddressKey
	netFlows    NetFlows
	serverRole  shila.EndpointRole
	ipFlow      shila.IPFlow
	inReader    *io.PipeReader
	inWriter    *io.PipeWriter
	connections *BackboneConnections
	lock        sync.Mutex
}

type NetFlows struct {
	effective   shila.NetFlow
	represented shila.NetFlow
}

func newBackboneConnection(rAddress shila.NetworkAddress, conns *BackboneConnections) (*BackboneConnection, error) {

	path, _ := network.PathGenerator{}.New("")		// FIXME: path!
	netFlow := shila.NetFlow{							// Net flow which is represented by this connection.
		Src:  conns.server.lAddress.(*snet.UDPAddr),
		Path: path,
		Dst:  rAddress.(*snet.UDPAddr),
	}

	inReader, inWriter := io.Pipe()

	conn := &BackboneConnection{
		keys:		 make([] shila.NetworkAddressKey, 2) ,
		netFlows:	 NetFlows{effective: netFlow, represented: netFlow},
		serverRole:  conns.server.Role(),
		inReader:    inReader,
		inWriter:    inWriter,
		connections: conns,
	}

	conn.keys = append(conn.keys, shila.GetNetworkAddressKey(rAddress))

	go conn.decodeIngress()		// Start the decoder.

	return conn, nil
}

func (conn *BackboneConnection) decodeIngress() {

	// The first message should be a control message which contains
	// all the information necessary to setup the backbone connection.
	ctrlMsg, err := conn.retrieveControlMessage()
	if err != nil {
		conn.removeConnection()
		return
	}

	// Process the control message.
	if err := conn.processControlMessage(ctrlMsg); err != nil {
		conn.removeConnection()
		return
	}

	// Now we are ready to listen for and process payload.
	for {
		if err := conn.processPayloadMessage(); err != nil {
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
	if conn.serverRole == shila.ContactingNetworkEndpoint {
		conn.netFlows.represented.Src.(*snet.UDPAddr).Host.Port = conn.ipFlow.Src.Port
	}

	// If the backbone connect is part of a traffic serer network endpoint, then the connection
	// has to be found by messages send along the corresponding contact backbone connection as well.
	if conn.serverRole == shila.TrafficNetworkEndpoint {
		conn.keys = append(conn.keys, shila.GetNetworkAddressKey(&ctrlMsg.SrcAddrContactEndpoint))
		conn.connections.add(conn.keys[1], conn)
	}

	return nil
}

func (conn *BackboneConnection) processPayloadMessage() error {
	return nil
}

func (conn *BackboneConnection) removeConnection() {
	for _, key := range conn.keys {
		conn.connections.remove(key)
	}
}

func (conn *BackboneConnection) writeIngress(buff []byte) {
	conn.inWriter.Write(buff)
}

func (conn *BackboneConnection) writeEgress(buff []byte) {
}


/*
func (server *Server) serveIngress(connection *networkConnection) {

	ingressRaw := make(chan byte, Config.SizeRawIngressBuffer)
	go server.packetize(connection.RepresentingFlow, ingressRaw)

	reader := io.Reader(connection.Backbone)
	storage := make([]byte, Config.SizeRawIngressStorage)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, Config.ReadSizeRawIngress)
		// If the incoming connection suffers from an error, we close it and return. The server instance is still able
		// to receive BackboneConnections as long as it is not shut down by the manager of the network side.
		if err != nil {
			close(ingressRaw) // Stop the packetizing.
			return
		}
		for _, b := range storage[:nBytesRead] {
			ingressRaw <- b
		}
	}
}

func (server *Server) packetize(flow shila.Flow, ingressRaw chan byte) {
	for {
		if rawData, err := tcpip.PacketizeRawData(ingressRaw, Config.SizeRawIngressStorage); rawData != nil {
			server.ingress <- shila.NewPacketWithNetFlow(server, flow.IPFlow.Swap(), flow.NetFlow.Swap(), rawData)
		} else {
			if err == nil {
				// All good, ingress raw closed.
				return
			}
			err := shila.PrependError(shila.ParsingError(err.Error()), "Issue in raw data packetizer.")
			server.issues <- shila.EndpointIssuePub{	Issuer: server, Flow: flow, Error: err }
			return
		}
	}
}

func (server *Server) resendFunctionality() {
	/*
		for {
			time.Sleep(Config.ServerResendInterval)
			server.lock.Lock()
			for _, p := range server.holdingArea {
				if p.TTL > 0 {
					p.TTL--
					server.egress <- p
				} else {
					// Server network endpoint is not able to send out the given packet.
					err := shila.NetworkEndpointTimeout("Unable to send packet.")
					server.issues <- shila.EndpointIssuePub { Issuer: server,	Flow: p.Flow, Error: err }
				}
			}
			server.holdingArea = server.holdingArea[:0]
			server.lock.Unlock()
		}

}

	func (server *Server) handleBackboneConnection(backConn *net.TCPConn) {

			// Generate the keys
			var keys []shila.NetworkAddressKey
			keys = append(keys, shila.GetNetworkAddressKey(representingFlow.NetFlow.Dst))
			// We need also to be able to send messages to the contact client network endpoint.
			if server.Role() == shila.TrafficNetworkEndpoint {
				// For the moment we can use the same path for this key as for the representing lAddress.
				keys = append(keys, shila.GetNetworkAddressKey(&ctrlMsg.SrcAddrContactEndpoint))
			}

			// Create the connection wrapper
			connection := networkConnection{
				EndpointRole:     server.role,
				TrueNetFlow:      trueNetFlow,
				RepresentingFlow: representingFlow,
				Backbone:         backConn,
			}

			// Add the new backbone connection to the mapping, so that it can be found by the egress handler.
			if err := server.insertBackboneConnection(keys, &connection); err != nil {
				server.closeBackboneConnectionWithErrorMsg(backConn, trueNetFlow, err, "Cannot insert backbone conn.")
				return
			}

			// Start the ingress handler for the backbone connection.
			server.serveIngress(&connection)

			// No longer necessary or possible to serve the ingress, remove the backbone connection from the mapping.
			/*
			server.lock.Lock()
			for _, key := range keys {
				delete(server.backboneConnections, key)
			}
			server.lock.Unlock()

return
}
		*/


