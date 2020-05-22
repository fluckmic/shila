package networkEndpoint

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"shila/config"
	"shila/core/shila"
	"shila/layer/tcpip"
	"strconv"
	"strings"
	"sync"
)

var _ shila.ServerNetworkEndpoint = (*Server)(nil)

type Server struct{
	Base
	backboneConnections map[shila.NetworkAddressAndPathKey]  net.Conn
	networkConnectionId shila.NetFlow
	listener            net.Listener
	lock                sync.Mutex
	holdingArea         []*shila.Packet
	isValid             bool
	isSetup             bool // TODO: merge to "state" object
	isRunning           bool
}

func newServer(netConnId shila.NetFlow, label shila.EndpointLabel, config config.NetworkEndpoint) shila.ServerNetworkEndpoint {
	return &Server{
		Base: 			Base{
								label: 	label,
								config: config,
						},
		backboneConnections: make(map[shila.NetworkAddressAndPathKey]  net.Conn),
		networkConnectionId: netConnId,
		lock:                sync.Mutex{},
		holdingArea:         make([]*shila.Packet, 0, config.SizeHoldingArea),
		isValid:             true,
	}
}

func (s *Server) SetupAndRun() error {

	if s.IsRunning() {
		return Error(fmt.Sprint("Unable to setup and run server {", s.Label(), ",", s.Key(), "}. - Server is already running."))
	}

	if !s.IsValid() {
		return Error(fmt.Sprint("Unable to setup and run server {", s.Label(), ",", s.Key(), "}. - Server no longer valid."))
	}

	if s.IsSetup() {
		return nil
	}

	// set up the listener
	src := s.networkConnectionId.Src.(Address)
	listener, err := net.ListenTCP(src.Addr.Network(), &src.Addr)
	if err != nil {
		return Error(fmt.Sprint("Unable to setup and run server {", s.Label(), ",", s.Key(), "}. - ", err.Error()))
	}

	// Create the channels
	s.ingress = make(chan *shila.Packet, s.config.SizeIngressBuff)
	s.egress  = make(chan *shila.Packet, s.config.SizeEgressBuff)

	// Start listening for incoming backboneConnections.
	s.listener = listener
	go s.serveIncomingConnections()

	// log.Verbose.Print("Server {", s.Label(), "," , s.Key(), "} started to listen for incoming backbone connections on {", s.Key(), "}.")

	// Start to handle incoming packets
	go s.serveEgress()

	s.isSetup   = true
	s.isRunning = true

	return nil
}

func (s *Server) TearDown() error {

	// nt("Tear down server {", s.Label(), ",", s.Key(), "}.")

	s.isSetup 	= false
	s.isValid 	= false
	s.isRunning = false

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
	return shila.EndpointKey(shila.GetNetworkAddressKey(s.networkConnectionId.Src))
}

func (s *Server) IsRunning() bool {
	return s.isRunning
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

	// Not the address from the client side
	srcAddr := Generator{}.NewAddress(connection.RemoteAddr().String())
	// Not the path taken from client to this server
	path 	:= Generator{}.NewPath("")

	// Generate the keys
	var keys []shila.NetworkAddressAndPathKey
	keys = append(keys, shila.GetNetworkAddressAndPathKey(srcAddr, path))

	// The client traffic endpoint sends as a very first message
	// the src address of its corresponding contacting endpoint.
	if s.Label() == shila.TrafficNetworkEndpoint {
		if srcAddrReceived, err := bufio.NewReader(connection).ReadString('\n'); err != nil {

		} else {
			contactSrcAddr := Generator{}.NewAddress(strings.TrimSuffix(srcAddrReceived,"\n"))
			keys = append(keys, shila.GetNetworkAddressAndPathKey(contactSrcAddr, path))
		}
	}

	// Add the new connection to the mapping, such that it can be found by the egress handler.
	s.lock.Lock()
	for _, key := range keys {
		if _, ok := s.backboneConnections[key]; ok {
			s.lock.Unlock()
			panic(fmt.Sprint("Trying to add backbone connection with key {", key, "} in " + "server {", s.Label(),
			",", s.Key(), "}. There already exists a backbone connection with that key.")) // TODO: Handle panic!
		} else {
			s.backboneConnections[key] = connection
			// log.Verbose.Print("Server {", s.Label(), ",", s.Key(), "} started handling a new backbone connection {", key, "}.")
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
		// log.Verbose.Print("Server {", s.Label(), ",", s.Key(), "} removed backbone connection {", key, "}.")
		delete(s.backboneConnections, key)
	}
	s.lock.Unlock()

	return
}

func (s *Server) flushHoldingArea() {
	// log.Verbose.Print("Server {", s.Label(), " ", s.Key(), "} flushes the holding area.")
	for _, p := range s.holdingArea {
		s.egress <- p
	}
}

func (s *Server) serveIngress(connection net.Conn) {

	// Prepare everything for the packetizer
	ingressRaw := make(chan byte, s.config.SizeReadBuffer)

	if s.Label() == shila.ContactingNetworkEndpoint {
		// Server is the contacting server, it is his responsibility
		// to extract the necessary data from the ip packet to be able
		// to set the correct network networkConnectionId.
		go s.packetizeContacting(ingressRaw, connection)

	} else if s.Label() == shila.TrafficNetworkEndpoint {
		// Server receives normal traffic, the connection over which the
		// packet was received contains enough information to set
		// the correct network networkConnectionId.
		go s.packetizeTraffic(ingressRaw, connection)
	} else {
		panic(fmt.Sprint("Wrong server label {", s.Label(), "} in serving ingress functionality of " +
			"server {", s.Key(), "}.")) // TODO: Handle panic!
	}

	reader := io.Reader(connection)
	storage := make([]byte, s.config.SizeReadBuffer)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, s.config.BatchSizeRead)
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
			nBytesWritten, err := writer.Write(p.Payload)
			if err != nil && !s.IsValid() {
				// Error doesn't matter, kernel endpoint is no longer valid anyway.
				return
			} else if err != nil {
				panic(fmt.Sprint("Unable to send data for packet {", p.Flow.IPFlow.Key(), "} in the " +
					"server {", s.Label(), ",", s.Key(), "} for backbone connection key {", key ,"}. - ", err.Error())) // TODO: Handle panic!
			} else {
				_ = nBytesWritten
				// log.Verbose.Print("Server {", s.Label()," ", s.Key(), "} wrote {", nBytesWritten, "}.")
			}
		} else {
		// Currently there is no connection available to send the packet, the packet has therefore to wait
		// in the holding area. Whenever a new connection is established all the packets in the holding area
		// are again processed; hopefully they can be send out this time.
			// log.Verbose.Print("Server {", s.Label(), ",", s.Key(), "} directs packet with backbone connection key {", key, "} in holding area.")
			s.holdingArea = append(s.holdingArea, p)
		}
	}
}

func (s *Server) IsSetup() bool {
	return s.isSetup
}

func (s *Server) IsValid() bool {
	return s.isValid
}

func (s *Server) packetizeTraffic(ingressRaw chan byte, connection net.Conn) {

	// Create the packet network networkConnectionId
	dstAddr := s.networkConnectionId.Src
	srcAddr := Address{Addr: *connection.RemoteAddr().(*net.TCPAddr)}
	path 	:= Generator{}.NewPath("")
	header  := shila.NetFlow{Src: dstAddr, Path: path, Dst: srcAddr }

	for {
		if rawData  := tcpip.PacketizeRawData(ingressRaw, s.config.SizeReadBuffer); rawData != nil {
			if iPHeader, err := shila.GetIPFlow(rawData); err != nil {
				panic(fmt.Sprint("Unable to get IP networkConnectionId in packetizer of server {", s.Key(),
					"}. - ", err.Error())) // TODO: Handle panic!
			} else {
				s.ingress <- shila.NewPacketWithNetFlow(s, iPHeader, header, rawData)
			}
		} else {
			return // Raw ingress channel closed.
		}
	}
}

func (s *Server) packetizeContacting(ingressRaw chan byte, connection net.Conn) {

	// Fetch the parts for the packet network networkConnectionId which are fixed.
	path 		:= Generator{}.NewPath("")
	srcAddr 	:= Generator{}.NewAddress(connection.RemoteAddr().String())
	localAddr 	:= connection.LocalAddr().(*net.TCPAddr)

	for {
		if rawData  := tcpip.PacketizeRawData(ingressRaw, s.config.SizeReadBuffer); rawData != nil {
			if iPHeader, err := shila.GetIPFlow(rawData); err != nil {
				panic(fmt.Sprint("Unable to get IP networkConnectionId in packetizer of server {", s.Key(),
					"}. - ", err.Error())) // TODO: Handle panic!
			} else {
				dstAddr := Generator{}.NewAddress(net.JoinHostPort(localAddr.IP.String(), strconv.Itoa(iPHeader.Dst.Port)))
				header  := shila.NetFlow{Src: dstAddr, Path: path, Dst: srcAddr}
				s.ingress <- shila.NewPacketWithNetFlow(s, iPHeader, header, rawData)
			}
		} else {
			return // Raw ingress channel closed.
		}
	}
}