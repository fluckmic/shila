package networkEndpoint

import (
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"shila/core/shila"
	"shila/layer/tcpip"
	"shila/log"
	"shila/networkSide/network"
	"strconv"
	"sync"
	"time"
)

var _ shila.NetworkServerEndpoint = (*Server)(nil)

type Server struct{
	Base
	backboneConnections map[shila.NetworkAddressAndPathKey]  networkConnection
	flow	            shila.Flow
	listener            net.Listener
	lock                sync.Mutex
	holdingArea         []*shila.Packet
}

func NewServer(flow shila.Flow, label shila.EndpointLabel, endpointIssues shila.EndpointIssuePubChannel) shila.NetworkServerEndpoint {
	return &Server{
		Base: 			Base{
								label: 			label,
								state: 			shila.NewEntityState(),
								endpointIssues: endpointIssues,
						},
		backboneConnections: make(map[shila.NetworkAddressAndPathKey]  networkConnection),
		flow:             	 flow,
		lock:                sync.Mutex{},
		holdingArea:         make([]*shila.Packet, 0, Config.SizeHoldingArea),
	}
}

func (s *Server) SetupAndRun() error {

	if s.state.Not(shila.Uninitialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", s.state, "}."))
	}

	// set up the listener
	src := s.flow.NetFlow.Src.(*net.TCPAddr)
	listener, err := net.ListenTCP(src.Network(), src)
	if err != nil {
		return shila.ThirdPartyError(fmt.Sprint("Unable to setup and run server {", s.Label(), "} listening on {", s.Key(), "}. - ", err.Error()))
	}

	// Create the channels
	s.ingress = make(chan *shila.Packet, Config.SizeIngressBuffer)
	s.egress  = make(chan *shila.Packet, Config.SizeEgressBuffer)

	// Start listening for incoming backbone connections.
	s.listener = listener
	go s.serveIncomingConnections()

	log.Verbose.Print(s.msg("Started listening."))

	// Start to handle incoming packets
	go s.serveEgress()

	// Start the resending functionality
	go s.resending()

	s.state.Set(shila.Running)
	return nil
}

func (s *Server) TearDown() error {

	s.state.Set(shila.TornDown)

	// Close the listener
	// Server no longer listens for incoming connections
	err := s.listener.Close()

	// Close all incoming connections
	// Terminates all workers processing incoming an connection and the corresponding packetizer
	for _, conn := range s.backboneConnections {
		err = conn.Backbone.Close()
	}

	// Close the ingress channel
	// Working side no longer processes this endpoint
	close(s.ingress)

	return err
}

func (s *Server) TrafficChannels() shila.PacketChannels {
	return shila.PacketChannels{Ingress: s.ingress, Egress: s.egress}
}

func (s *Server) Label() shila.EndpointLabel {
	return s.label
}

func (s *Server) Key() shila.EndpointKey {
	return shila.EndpointKey(shila.GetNetworkAddressKey(s.flow.NetFlow.Src))
}

func (s *Server) serveIncomingConnections(){
	for {
		if connection, err := s.listener.Accept(); err != nil {
			return
		} else {
			go s.handleBackboneConnection(connection.(*net.TCPConn))
		}
	}
}

func (s *Server) handleBackboneConnection(backboneConnection *net.TCPConn) {

	// Fetch the network address from the client side as well as the path taken.
	srcAddr 	:= backboneConnection.RemoteAddr().(*net.TCPAddr)
	trueDstAddr := backboneConnection.LocalAddr().(*net.TCPAddr)
	path, err	:= network.PathGenerator{}.New("")
	if err != nil {
		s.closeBackboneConnection(backboneConnection, err); return
	}

	trueNetFlow := shila.NetFlow{
		Src:  srcAddr,
		Path: path,
		Dst:  trueDstAddr,
	}

	log.Verbose.Print(s.msgFlowRelated(trueNetFlow, "Accepted."))

	var srcAddrs [] shila.NetworkAddress
	srcAddrs = append(srcAddrs, srcAddr)

	// The client sends as very first message its corresponding IP flow.
	reader := io.Reader(backboneConnection)
	decoder := gob.NewDecoder(reader)
	var receivedFlow shila.IPFlow
	if err := decoder.Decode(&receivedFlow); err != nil {
		s.closeBackboneConnection(backboneConnection, err); return
	}

	// Determine the network address of this network endpoint depending on the functionality
	var dstAddr shila.NetworkAddress
	if s.Label() == shila.ContactingNetworkEndpoint {
		// It is the responsibility of the contacting server endpoint to determine the correct network source address
		dstAddr, err  = network.AddressGenerator{}.New(net.JoinHostPort(trueDstAddr.IP.String(), strconv.Itoa(receivedFlow.Dst.Port)))
		if err != nil {
			s.closeBackboneConnection(backboneConnection, err); return
		}

	} else if s.Label() == shila.TrafficNetworkEndpoint {

		// For the traffic server endpoint, the client sends the address of the corresponding contacting endpoint.
		reader := io.Reader(backboneConnection)
		decoder := gob.NewDecoder(reader)
		var dstAddrContacting net.TCPAddr
		if err := decoder.Decode(&dstAddrContacting); err != nil {
			s.closeBackboneConnection(backboneConnection, err); return
		}
		srcAddrs = append(srcAddrs, &dstAddrContacting)

	} else {
		s.closeBackboneConnection(backboneConnection, shila.CriticalError(fmt.Sprint("Wrong server label."))); return
	}

	connection := networkConnection{
		Identifier: shila.Flow{IPFlow: receivedFlow, NetFlow: shila.NetFlow{
			Src:  dstAddr,
			Path: path,
			Dst:  srcAddrs[0],
		}},
		Backbone:   backboneConnection,
	}

	// Generate the keys
	var keys []shila.NetworkAddressAndPathKey
	for _, dstAddr := range srcAddrs {
		keys = append(keys, shila.GetNetworkAddressAndPathKey(dstAddr, path))
	}

	// Add the new backboneConnection to the mapping, such that it can be found by the egress handler.
	s.lock.Lock()
	for _, key := range keys {
		if _, ok := s.backboneConnections[key]; ok {
			s.lock.Unlock()
			s.closeBackboneConnection(backboneConnection, err); return
		} else {
			s.backboneConnections[key] = connection
		}
	}
	s.lock.Unlock()

	// Start the ingress handler for the backboneConnection.
	s.serveIngress(connection)

	// No longer necessary or possible to serve the ingress, remove the backboneConnection from the mapping.
	s.lock.Lock()
	for _, key := range keys {
		delete(s.backboneConnections, key)
	}
	s.lock.Unlock()

	return
}

func (s *Server) closeBackboneConnection(connection *net.TCPConn, err error) {
	connection.Close()
	log.Error.Print("Closed backbone connection in Server {", s.Label(), ",", s.Key(), "}. - ", err.Error())
}

func (s *Server) serveIngress(connection networkConnection) {

	// Prepare everything for the packetizer
	ingressRaw := make(chan byte, Config.SizeRawIngressBuffer)
	go s.packetize(connection.Identifier, ingressRaw)

	reader := io.Reader(connection.Backbone)
	storage := make([]byte, Config.SizeRawIngressStorage)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, Config.ReadSizeRawIngress)
		// If the incoming connection suffers from an error, we close it and return.
		// The server instance is still able to receive backboneConnections as long as it is not
		// shut down by the manager of the network side.
		if err != nil {
			close(ingressRaw) // Stop the packetizing.
			return
		}
		for _, b := range storage[:nBytesRead] {
			ingressRaw <- b
		}
	}
}

func (s *Server) packetize(flow shila.Flow, ingressRaw chan byte) {
	for {
		if rawData, err := tcpip.PacketizeRawData(ingressRaw, Config.SizeRawIngressStorage); rawData != nil {
				s.ingress <- shila.NewPacketWithNetFlow(s, flow.IPFlow, flow.NetFlow, rawData)
		} else {
			if err == nil {
				// All good, ingress raw closed.
				return
			}
			s.endpointIssues <- shila.EndpointIssuePub{
				Publisher: 	s,
				Flow:		flow,
				Error:     	shila.PrependError(err, "Error in raw data packetizer."),
			}
			return
		}
	}
}

func (s *Server) resending() {
	for {
		time.Sleep(Config.ServerResendInterval)
		for _, p := range s.holdingArea {
			if p.TTL > 0 {
				p.TTL--
				s.egress <- p
			} else {
				s.endpointIssues <- shila.EndpointIssuePub{
					Publisher: s,
					Flow:      p.Flow,
					Error:     shila.ThirdPartyError("Unable to write data."),
				}
			}
		}
	}
}

func (s *Server) serveEgress() {
	for p := range s.egress {
		// Retrieve key to get the correct connection
		key := p.Flow.NetFlow.DstAndPathKey()
		if con, ok := s.backboneConnections[key]; ok {
			writer := io.Writer(con.Backbone)
			_, err := writer.Write(p.Payload)
			if err != nil && s.state.Not(shila.Running) {
				// Server turned down anyway.
				return
			}
			// If just the connection was closed, then we ignore the error and drop the packet.
			// The ingress handler has observed the issue as well and will sooner or later remove
			// the closed (or faulty) connection.
		} else {
			// Currently there is no backbone connection available to send the packet.
			// It's TTL value is decreased and it is put into the holding area.
			s.holdingArea = append(s.holdingArea, p)
			// log.Verbose.Print("Server {", s.Label(), "} listening on {", s.Key(), "} directs packet for " +
			// "backbone connection key {", key, "} into holding area.")
		}
	}
}

func (s *Server) Flow() shila.Flow {
	return s.flow
}

func (s *Server) msg(str string) string {
	return fmt.Sprint("Server {", s.Label(), " - ", s.flow.NetFlow.Src, " <- *}: ", str)
}

func (s *Server) msgFlowRelated(flow shila.NetFlow, str string) string {
	return fmt.Sprint("Server {", s.Label(), " - ", flow.Dst.String(),
		" <- ", flow.Src.String(),"}: ", str)
}