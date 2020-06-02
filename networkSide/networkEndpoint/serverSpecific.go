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
	backboneConnections map[shila.NetworkAddressAndPathKey]  *networkConnection
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
		backboneConnections: make(map[shila.NetworkAddressAndPathKey]  *networkConnection),
		flow:             	 flow,
		lock:                sync.Mutex{},
		holdingArea:         make([]*shila.Packet, 0, Config.SizeHoldingArea),
	}
}

func (s *Server) SetupAndRun() error {

	if s.state.Not(shila.Uninitialized) {
		return shila.CriticalError(s.msg(fmt.Sprint("In wrong state {", s.state, "}.")))
	}

	// Set up the listener
	src := s.flow.NetFlow.Src.(*net.TCPAddr)
	listener, err := net.ListenTCP(src.Network(), src)
	if err != nil {
		return shila.ThirdPartyError(shila.PrependError(err, s.msg(fmt.Sprint("Unable to setup listener."))).Error())
	}
	s.listener = listener

	// Create the channels
	s.ingress = make(chan *shila.Packet, Config.SizeIngressBuffer)
	s.egress  = make(chan *shila.Packet, Config.SizeEgressBuffer)

	// Start listening for incoming backbone connections.
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

	err := s.listener.Close()		// Close the listener, server no longer listens for incoming connections

	// Close all incoming connections
	// Terminates all workers processing incoming an connection and the corresponding packetizer
	for _, conn := range s.backboneConnections {
		err = conn.Backbone.Close()
	}

	close(s.ingress)	// Close the ingress channel (Working side no longer processes this endpoint)

	log.Verbose.Print(s.msg("Got torn down."))
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

func (s *Server) handleBackboneConnection(backConn *net.TCPConn) {

	// Create the true net flow (Server: src <- dst)
	path, _ := network.PathGenerator{}.New("")
	trueNetFlow := shila.NetFlow{
		Src:  backConn.LocalAddr().(*net.TCPAddr),
		Path: path,
		Dst:  backConn.RemoteAddr().(*net.TCPAddr),
	}

	log.Verbose.Print(s.msgFlowRelated(trueNetFlow, "Accepted backbone connection."))

	// Fetch the control message
	var ctrlMsg controlMessage
	if err := gob.NewDecoder(io.Reader(backConn)).Decode(&ctrlMsg); err != nil {
		s.closeBackboneConnectionWithErrorMsg(backConn, trueNetFlow, err, "Cannot fetch control message.")
		return
	}

	// Create the representing flow (Content of control message is correct relative to sender)
	representingFlow := shila.Flow{ IPFlow: ctrlMsg.IPFlow.Swap(), NetFlow: trueNetFlow }

	// If endpoint is a contacting endpoint, then the representing flow is different from the true net flow:
	if s.Label() == shila.ContactingNetworkEndpoint {
		srcAddrTrafficEndpoint, err := s.calculateSrcAddrOfTrafficServerNetworkEndpoint(representingFlow)
		if err != nil {
			s.closeBackboneConnectionWithErrorMsg(backConn, trueNetFlow, err, "Cannot generate src address of traffic server network endpoint.")
			return
		}
		representingFlow.NetFlow.Src = srcAddrTrafficEndpoint
	}

	// Generate the keys
	var keys []shila.NetworkAddressAndPathKey
	keys = append(keys, shila.GetNetworkAddressAndPathKey(representingFlow.NetFlow.Dst, representingFlow.NetFlow.Path))
	// We need also to be able to send messages to the contact client network endpoint.
	if s.Label() == shila.TrafficNetworkEndpoint {
		// For the moment we can use the same path for this key as for the representing flow.
		keys = append(keys, shila.GetNetworkAddressAndPathKey(&ctrlMsg.SrcAddrContactEndpoint, representingFlow.NetFlow.Path))
	}

	// Create the connection wrapper
	connection := networkConnection{
		Label:            s.label,
		TrueNetFlow:      trueNetFlow,
		RepresentingFlow: representingFlow,
		Backbone:         backConn,
	}

	// Add the new backbone connection to the mapping, so that it can be found by the egress handler.
	if err := s.insertBackboneConnection(keys, &connection); err != nil {
		s.closeBackboneConnectionWithErrorMsg(backConn, trueNetFlow, err, "Cannot insert backbone conn.")
		return
	}

	// Start the ingress handler for the backbone connection.
	s.serveIngress(&connection)

	// No longer necessary or possible to serve the ingress, remove the backbone connection from the mapping.
	s.lock.Lock()
	for _, key := range keys {
		delete(s.backboneConnections, key)
	}
	s.lock.Unlock()

	return
}

func (s* Server) insertBackboneConnection(keys []shila.NetworkAddressAndPathKey, conn *networkConnection) error {

	s.lock.Lock()
	defer s.lock.Unlock()

	for _, key := range keys {
		if _, ok := s.backboneConnections[key]; ok {
			return shila.TolerableError(fmt.Sprint("Duplicate key {", key, "}."))
		}
	}
	for _, key := range keys {
		s.backboneConnections[key] = conn
	}
	return nil
}

func (s* Server) calculateSrcAddrOfTrafficServerNetworkEndpoint(representingFlow shila.Flow) (shila.NetworkAddress, error) {
	// It is the responsibility of the contact server endpoint to calculate the
	// source address of the corresponding traffic server endpoint.
	ip   := representingFlow.NetFlow.Src.(*net.TCPAddr).IP.String()
	port := strconv.Itoa(representingFlow.IPFlow.Src.Port)
	return network.AddressGenerator{}.New(net.JoinHostPort(ip, port))
}

func (s* Server) closeBackboneConnectionWithErrorMsg(conn *net.TCPConn, flow shila.NetFlow, err error, msg string) {
	log.Error.Print(s.msgFlowRelated(flow, shila.PrependError(err, msg).Error()))
	conn.Close()
	log.Error.Print(s.msgFlowRelated(flow, "Closed backbone connection."))
}

func (s *Server) serveIngress(connection *networkConnection) {

	ingressRaw := make(chan byte, Config.SizeRawIngressBuffer)
	go s.packetize(connection.RepresentingFlow, ingressRaw)

	reader := io.Reader(connection.Backbone)
	storage := make([]byte, Config.SizeRawIngressStorage)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, Config.ReadSizeRawIngress)
		// If the incoming connection suffers from an error, we close it and return. The server instance is still able
		// to receive backboneConnections as long as it is not shut down by the manager of the network side.
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
			err := shila.PrependError(shila.ParsingError(err.Error()), "Issue in raw data packetizer.")
			s.endpointIssues <- shila.EndpointIssuePub{	Issuer: s, Flow: flow, Error: err }
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
				// Server network endpoint is not able to send out the given packet.
				err := shila.NetworkEndpointTimeout("Unable to send packet.")
				s.endpointIssues <- shila.EndpointIssuePub { Issuer: s,	Flow: p.Flow, Error: err }
			}
		}
	}
}

func (s *Server) serveEgress() {
	for p := range s.egress {
		key := p.Flow.NetFlow.DstAndPathKey()				// Retrieve key to get the correct connection
		if con, ok := s.backboneConnections[key]; ok {
			writer := io.Writer(con.Backbone)
			_, err := writer.Write(p.Payload)
			if err != nil && s.state.Not(shila.Running) {
				return										// Server turned down anyway.
			}
			// If just the connection was closed, then we ignore the error and drop the packet. The ingress handler
			// has observed the issue as well and will sooner or later remove the closed (or faulty) connection.
		} else {
			// Currently there is no backbone connection available to send the packet, put packet into holding area.
			s.holdingArea = append(s.holdingArea, p)
		}
	}
}

func (s *Server) msg(str string) string {
	return fmt.Sprint("Server {", s.Label(), " - ", s.flow.NetFlow.Src, " <- *}: ", str)
}

func (s *Server) msgFlowRelated(flow shila.NetFlow, str string) string {
	return fmt.Sprint("Server {", s.Label(), " - ", flow.Src.String(), " <- ", flow.Dst.String(),"}: ", str)
}