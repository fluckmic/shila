package networkEndpoint

import (
	"shila/core/shila"
)

type BackboneConnections struct {}

func (conns *BackboneConnections) retrieve(key shila.NetworkAddressKey) *BackboneConnection {
	return &BackboneConnection{}
}

func (conns *BackboneConnections) tearDown() error {
	return nil
}

func NewBackboneConnections() BackboneConnections {
	return BackboneConnections{}
}



type BackboneConnection struct {}

func (conn *BackboneConnection) writeIngress(buff []byte) {}

func (conn *BackboneConnection) writeEgress(buff []byte)  {
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

		/*

func (server * Server) insertBackboneConnection(keys []shila.NetworkAddressKey, conn *networkConnection) error {


	server.lock.Lock()
	defer server.lock.Unlock()

	for _, key := range keys {
		if _, ok := server.backboneConnections[key]; ok {
			return shila.TolerableError(fmt.Sprint("Duplicate key {", key, "}."))
		}
	}
	for _, key := range keys {
		server.backboneConnections[key] = conn
	}
	return nil

return nil
}

func (server * Server) calculateSrcAddrOfTrafficServerNetworkEndpoint(representingFlow shila.Flow) (shila.NetworkAddress, error) {
	// It is the responsibility of the contact server endpoint to calculate the
	// source address of the corresponding traffic server endpoint.
	ip   := representingFlow.NetFlow.Src.(*net.TCPAddr).IP.String()
	port := strconv.Itoa(representingFlow.IPFlow.Src.Port)
	return network.AddressGenerator{}.New(net.JoinHostPort(ip, port))
}

func (server * Server) closeBackboneConnectionWithErrorMsg(conn *net.TCPConn, flow shila.NetFlow, err error, Says string) {
	log.Error.Print(server.msgFlowRelated(flow, shila.PrependError(err, Says).Error()))
	conn.Close()
	log.Error.Print(server.msgFlowRelated(flow, "Closed backbone connection."))
}

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

			// Create the true net lAddress (Server: src <- dst)
			path, _ := network.PathGenerator{}.New("")
			trueNetFlow := shila.NetFlow{
				Src:  backConn.LocalAddr().(*net.TCPAddr),
				Path: path,
				Dst:  backConn.RemoteAddr().(*net.TCPAddr),
			}

			log.Verbose.Print(server.msgFlowRelated(trueNetFlow, "Accepted backbone connection."))

			// Fetch the control message
			var ctrlMsg controlMessage
			if err := gob.NewDecoder(io.Reader(backConn)).Decode(&ctrlMsg); err != nil {
				server.closeBackboneConnectionWithErrorMsg(backConn, trueNetFlow, err, "Cannot fetch control message.")
				return
			}

			// Create the representing lAddress (Content of control message is correct relative to sender)
			representingFlow := shila.Flow{ IPFlow: ctrlMsg.IPFlow.Swap(), NetFlow: trueNetFlow }

			// If endpoint is a contacting endpoint, then the representing lAddress is different from the true net lAddress:
			if server.Role() == shila.ContactingNetworkEndpoint {
				srcAddrTrafficEndpoint, err := server.calculateSrcAddrOfTrafficServerNetworkEndpoint(representingFlow)
				if err != nil {
					server.closeBackboneConnectionWithErrorMsg(backConn, trueNetFlow, err, "Cannot generate src address of traffic server network endpoint.")
					return
				}
				representingFlow.NetFlow.Src = srcAddrTrafficEndpoint
			}

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


