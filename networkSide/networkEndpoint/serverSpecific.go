package networkEndpoint

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"shila/core/shila"
	"shila/layer/tcpip"
	"shila/log"
	"shila/networkSide/network"
	"strconv"
	"strings"
	"sync"
)

var _ shila.NetworkServerEndpoint = (*Server)(nil)

type Server struct{
	Base
	backboneConnections map[shila.NetworkAddressAndPathKey]  net.Conn
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
		backboneConnections: make(map[shila.NetworkAddressAndPathKey]  net.Conn),
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

	log.Verbose.Print("Server {", s.Label(), "} started to listen for incoming backbone connections on {", s.Key(), "}.")

	// Start to handle incoming packets
	go s.serveEgress()

	s.state.Set(shila.Running)
	return nil
}

func (s *Server) TearDown() error {

	s.state.Set(shila.TornDown)

	// Close the egress channel
	// Server stops sending out packets
	close(s.egress)

	// Close the listener
	// Server no longer listens for incoming connections
	err := s.listener.Close()

	// Close all incoming connections
	// Terminates all workers processing incoming an connection and the corresponding packetizer
	for _, conn := range s.backboneConnections {
		err = conn.Close()
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
			go s.handleConnection(connection)
		}
	}
}

func (s *Server) handleConnection(connection net.Conn) {

	// Get the address from the client side
	srcAddr, err := network.AddressGenerator{}.New(connection.RemoteAddr().String())
	if err != nil {
		s.closeConnection(connection, err); return
	}

	// Get the path taken from client to this server
	path, err	:= network.PathGenerator{}.New("")
	if err != nil {
		s.closeConnection(connection, err); return
	}

	// Generate the keys
	var keys []shila.NetworkAddressAndPathKey
	keys = append(keys, shila.GetNetworkAddressAndPathKey(srcAddr, path))

	// The client traffic endpoint sends as a very first message
	// the src address of its corresponding contacting endpoint.
	if s.Label() == shila.TrafficNetworkEndpoint {
		if srcAddrReceived, err := bufio.NewReader(connection).ReadString('\n'); err != nil {
			s.closeConnection(connection, err); return
		} else {
			contactSrcAddr, _ := network.AddressGenerator{}.New(strings.TrimSuffix(srcAddrReceived,"\n"))
			keys = append(keys, shila.GetNetworkAddressAndPathKey(contactSrcAddr, path))
		}
	}

	// Add the new connection to the mapping, such that it can be found by the egress handler.
	s.lock.Lock()
	for _, key := range keys {
		if _, ok := s.backboneConnections[key]; ok {
			s.lock.Unlock()
			s.closeConnection(connection, err); return
		} else {
			s.backboneConnections[key] = connection
			log.Verbose.Print("Server {", s.Label(), "} listening on {", s.Key(), "} started handling a new backbone connection {", key, "}.")
		}
	}
	s.lock.Unlock()

	// A new connection was established. This is good news for
	// all packets waiting in the holding area.
	go s.flushHoldingArea()

	// Start the ingress handler for the connection.
	s.serveIngress(connection)

	// No longer necessary or possible to serve the ingress, remove the connection from the mapping.
	s.lock.Lock()
	for _, key := range keys {
		log.Verbose.Print("Server {", s.Label(), "} listening on {", s.Key(), "} removed backbone connection {", key, "}.")
		delete(s.backboneConnections, key)
	}
	s.lock.Unlock()

	return
}

func (s *Server) closeConnection(connection net.Conn, err error) {
	connection.Close()
	log.Error.Print("Closed connection in Server {", s.Label(), ",", s.Key(), "}. - ", err.Error())
}

func (s *Server) flushHoldingArea() {
	log.Verbose.Print("Server {", s.Label(), "} listening on {", s.Key(), "} flushes the holding area.")
	for _, p := range s.holdingArea {
		s.egress <- p
	}
}

func (s *Server) serveIngress(connection net.Conn) {

	// Prepare everything for the packetizer
	ingressRaw := make(chan byte, Config.SizeRawIngressBuffer)

	if s.Label() == shila.ContactingNetworkEndpoint {
		// Server is the contacting server, it is his responsibility
		// to extract the necessary data from the ip packet to be able
		// to set the correct network netFlow.
		go s.packetizeContacting(ingressRaw, connection)

	} else if s.Label() == shila.TrafficNetworkEndpoint {
		// Server receives normal traffic, the connection over which the
		// packet was received contains enough information to set
		// the correct network netFlow.
		go s.packetizeTraffic(ingressRaw, connection)
	} else {
		s.closeConnection(connection, shila.CriticalError(fmt.Sprint("Wrong server label."))); return
	}

	reader := io.Reader(connection)
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

func (s *Server) serveEgress() {
	for p := range s.egress {
		// Retrieve key to get the correct connection
		key := p.Flow.NetFlow.DstAndPathKey()
		if con, ok := s.backboneConnections[key]; ok {
			writer := io.Writer(con)
			_, err := writer.Write(p.Payload)
			if err != nil && s.state.Not(shila.Running) {
				// Server turned down anyway.
				return
			}
			// If just the connection was closed, then we ignore the error. The ingress handler has observed the
			// issue as well and will sooner or later remove the closed (or faulty) connection.

			// As soon as the connection is removed, follow up packets will be directed to the holding area, but
			// never be send out, since a reconnection has a different key.
			// TODO: https://github.com/fluckmic/shila/issues/14
		} else {
		// Currently there is no connection available to send the packet, the packet has therefore to wait
		// in the holding area. Whenever a new connection is established all the packets in the holding area
		// are again processed; hopefully they can be send out this time.
			log.Verbose.Print("Server {", s.Label(), "} listening on {", s.Key(), "} directs packet for " +
				"backbone connection key {", key, "} into holding area.")
			s.holdingArea = append(s.holdingArea, p)
		}
	}
}

func (s *Server) packetizeTraffic(ingressRaw chan byte, connection net.Conn) {

	// Closing the connection inside of this go routine terminates the corresponding ingress handler!

	// Create the packet network netFlow
	dstAddr 	 := s.flow.NetFlow.Src

	srcAddr, err := network.AddressGenerator{}.New(connection.RemoteAddr().String())
	if err != nil {
		s.closeConnection(connection, shila.PrependError(err, "Unable to get source address.")); return
	}

	path, err	 := network.PathGenerator{}.New("")
	if err != nil {
		s.closeConnection(connection, shila.PrependError(err, "Unable to get path.")); return
	}

	header  	 := shila.NetFlow{Src: dstAddr, Path: path, Dst: srcAddr }

	for {
		if rawData, err := tcpip.PacketizeRawData(ingressRaw, Config.SizeRawIngressStorage); rawData != nil {
			if iPHeader, err := shila.GetIPFlow(rawData); err != nil {
				// We were not able to get the IP flow from the raw data, but there was no issue parsing
				// the raw data. We therefore just drop the packet and hope that the next one is better..
				log.Error.Print("Unable to get IP net flow in traffic packetizer of server {", s.Key(),	"}. - ", err.Error())
			} else {
				s.ingress <- shila.NewPacketWithNetFlow(s, iPHeader, header, rawData)
			}
		} else {
			if err == nil {
				// All good, ingress raw closed.
				return
			}
			s.closeConnection(connection, shila.PrependError(err, "Issue in raw packetizer."))
		}
	}
}

func (s *Server) packetizeContacting(ingressRaw chan byte, connection net.Conn) {

	// Closing the connection inside of this go routine terminates the corresponding ingress handler!

	// Fetch the parts for the packet network netFlow which are fixed.
	srcAddr, err := network.AddressGenerator{}.New(connection.RemoteAddr().String())
	if err != nil {
		s.closeConnection(connection, shila.PrependError(err, "Unable to get source address.")); return
	}

	path, err	 := network.PathGenerator{}.New("")
	if err != nil {
		s.closeConnection(connection, shila.PrependError(err, "Unable to get path.")); return
	}

	localAddr 	 := connection.LocalAddr().(*net.TCPAddr)

	for {
		if rawData, err := tcpip.PacketizeRawData(ingressRaw, Config.SizeRawIngressStorage); rawData != nil {
			if iPHeader, err := shila.GetIPFlow(rawData); err != nil {
				// We were not able to get the IP flow from the raw data, but there was no issue parsing
				// the raw data. We therefore just drop the packet and hope that the next one is better..
				log.Error.Print("Unable to get IP net flow in contact packetizer of server {", s.Key(),	"}. - ", err.Error())
			} else {
				dstAddr, _  := network.AddressGenerator{}.New(net.JoinHostPort(localAddr.IP.String(), strconv.Itoa(iPHeader.Dst.Port)))
				header  	:= shila.NetFlow{Src: dstAddr, Path: path, Dst: srcAddr}
				s.ingress <- shila.NewPacketWithNetFlow(s, iPHeader, header, rawData)
			}
		} else {
			if err == nil {
				// All good, ingress raw closed.
				return
			}
			s.closeConnection(connection, shila.PrependError(err, "Issue in raw packetizer."))
		}
	}
}

func (s *Server) Flow() shila.Flow {
	return s.flow
}