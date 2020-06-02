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

	// Create the true net flow.
	// Server: src <- dst
	path, _ := network.PathGenerator{}.New("")
	trueNetFlow := shila.NetFlow{
		Src:  backboneConnection.LocalAddr().(*net.TCPAddr),
		Path: path,
		Dst:  backboneConnection.RemoteAddr().(*net.TCPAddr),
	}

	log.Verbose.Print(s.msgFlowRelated(trueNetFlow, "Accepted backbone connection."))

	// Receive the control message
	type controlMessage struct {
		IPFlow                 shila.IPFlow
		srcAddrContactEndpoint net.TCPAddr
	}
	var ctrlMsg controlMessage
	if err := gob.NewDecoder(io.Reader(backboneConnection)).Decode(&ctrlMsg); err != nil {
		log.Error.Print(s.msgFlowRelated(trueNetFlow, shila.PrependError(err, "Unable to fetch control message.").Error()))
		backboneConnection.Close()
		log.Error.Print(s.msgFlowRelated(trueNetFlow, "Closed backbone connection."))
		return
	}

	// Create the representing flow
	representingFlow := shila.Flow{ IPFlow: ctrlMsg.IPFlow.Swap(), NetFlow: trueNetFlow }

	// If the endpoint is a contacting endpoint, then the representing flow is different from the true one.
	if s.Label() == shila.ContactingNetworkEndpoint {
		// It is the responsibility of the contact server endpoint to calculate the
		// source address of the corresponding traffic server endpoint.
		ip   := trueNetFlow.Src.(*net.TCPAddr).IP.String()
		port := strconv.Itoa(representingFlow.IPFlow.Src.Port)
		srcAddrTrafficEndpoint, err := network.AddressGenerator{}.New(net.JoinHostPort(ip, port))
		if err != nil {
			log.Error.Print(s.msgFlowRelated(trueNetFlow, shila.PrependError(err, "Cannot generate source address of traffic server endpoint.").Error()))
			backboneConnection.Close()
			log.Error.Print(s.msgFlowRelated(trueNetFlow, "Closed backbone connection."))
			return
		}
		representingFlow.NetFlow.Src = srcAddrTrafficEndpoint
	}

	// Generate the keys
	var keys []shila.NetworkAddressAndPathKey
	keys = append(keys, shila.GetNetworkAddressAndPathKey(representingFlow.NetFlow.Dst, representingFlow.NetFlow.Path))
	// We need also to be able to send messages to the contact client network endpoints.
	if s.Label() == shila.TrafficNetworkEndpoint {
		// For the moment we can use the same path for this key as for the representing flow.
		keys = append(keys, shila.GetNetworkAddressAndPathKey(&ctrlMsg.srcAddrContactEndpoint, representingFlow.NetFlow.Path))
	}

	// Create the connection wrapper
	connection := networkConnection{
		Label:            s.label,
		TrueNetFlow:      trueNetFlow,
		RepresentingFlow: representingFlow,
		Backbone:         backboneConnection,
	}

	// Add the new backboneConnection to the mapping, such that it can be found by the egress handler.
	s.lock.Lock()
	for _, key := range keys {
		if _, ok := s.backboneConnections[key]; ok {
			log.Error.Panic("Implement me.") // TODO.
			return
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

func (s *Server) serveIngress(connection networkConnection) {

	// Prepare everything for the packetizer
	ingressRaw := make(chan byte, Config.SizeRawIngressBuffer)

	// Representing net flow of server: src <- dst

	// The destination of the incoming packets is the representing source of the server.
	// The source of the incoming packets is the representing destination of the server.
	go s.packetize(connection.RepresentingFlow, ingressRaw)

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
				s.ingress <- shila.NewPacketWithNetFlow(s, flow.IPFlow.Swap(), flow.NetFlow.Swap(), rawData)
		} else {
			if err == nil {
				// All good, ingress raw closed.
				return
			}
			s.endpointIssues <- shila.EndpointIssuePub{
				Publisher: 	s,
				Flow:		flow,		// Publisher flow!
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
	return fmt.Sprint("Server {", s.Label(), " - ", flow.Src.String(),
		" <- ", flow.Dst.String(),"}: ", str)
}