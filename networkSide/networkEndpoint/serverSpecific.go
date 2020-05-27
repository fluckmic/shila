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
	netFlow             shila.NetFlow
	listener            net.Listener
	lock                sync.Mutex
	holdingArea         []*shila.Packet
	isValid             bool
	isSetup             bool // TODO: merge to "state" object
	isRunning           bool
}

func NewServer(flow shila.NetFlow, label shila.EndpointLabel) shila.NetworkServerEndpoint {
	return &Server{
		Base: 			Base{
								label: 	label,
						},
		backboneConnections: make(map[shila.NetworkAddressAndPathKey]  net.Conn),
		netFlow:             flow,
		lock:                sync.Mutex{},
		holdingArea:         make([]*shila.Packet, 0, Config.SizeHoldingArea),
		isValid:             true,
	}
}

func (s *Server) SetupAndRun() error {

	if s.IsRunning() {
		return  shila.CriticalError(fmt.Sprint("Unable to setup and run server {", s.Label(), ",", s.Key(), "}. - Server is already running."))
	}

	if !s.IsValid() {
		return  shila.CriticalError(fmt.Sprint("Unable to setup and run server {", s.Label(), ",", s.Key(), "}. - Server no longer valid."))
	}

	if s.IsSetup() {
		return nil
	}

	// set up the listener
	src := s.netFlow.Src.(*net.TCPAddr)
	listener, err := net.ListenTCP(src.Network(), src)
	if err != nil {
		return shila.ThirdPartyError(fmt.Sprint("Unable to setup and run server {", s.Label(), "} listening on {", s.Key(), "}. - ", err.Error()))
	}

	// Create the channels
	s.ingress = make(chan *shila.Packet, Config.SizeIngressBuff)
	s.egress  = make(chan *shila.Packet, Config.SizeEgressBuff)

	// Start listening for incoming backbone connections.
	s.listener = listener
	go s.serveIncomingConnections()

	log.Verbose.Print("Server {", s.Label(), "} started to listen for incoming backbone connections on {", s.Key(), "}.")

	// Start to handle incoming packets
	go s.serveEgress()

	s.isSetup   = true
	s.isRunning = true

	return nil
}

func (s *Server) TearDown() error {

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
	return shila.EndpointKey(shila.GetNetworkAddressKey(s.netFlow.Src))
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
	srcAddr := network.AddressGenerator{}.New(connection.RemoteAddr().String())
	// Not the path taken from client to this server
	path 	:= network.PathGenerator{}.New("")

	// Generate the keys
	var keys []shila.NetworkAddressAndPathKey
	keys = append(keys, shila.GetNetworkAddressAndPathKey(srcAddr, path))

	// The client traffic endpoint sends as a very first message
	// the src address of its corresponding contacting endpoint.
	if s.Label() == shila.TrafficNetworkEndpoint {
		if srcAddrReceived, err := bufio.NewReader(connection).ReadString('\n'); err != nil {

		} else {
			contactSrcAddr := network.AddressGenerator{}.New(strings.TrimSuffix(srcAddrReceived,"\n"))
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

func (s *Server) flushHoldingArea() {
	log.Verbose.Print("Server {", s.Label(), "} listening on {", s.Key(), "} flushes the holding area.")
	for _, p := range s.holdingArea {
		s.egress <- p
	}
}

func (s *Server) serveIngress(connection net.Conn) {

	// Prepare everything for the packetizer
	ingressRaw := make(chan byte, Config.SizeReadBuffer)

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
		panic(fmt.Sprint("Wrong server label {", s.Label(), "} in serving ingress functionality of " +
			"server {", s.Key(), "}.")) // TODO: Handle panic!
	}

	reader := io.Reader(connection)
	storage := make([]byte, Config.SizeReadBuffer)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, Config.BatchSizeRead)
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
			}
		} else {
		// Currently there is no connection available to send the packet, the packet has therefore to wait
		// in the holding area. Whenever a new connection is established all the packets in the holding area
		// are again processed; hopefully they can be send out this time.
			log.Verbose.Print("Server {", s.Label(), "} listening on {", s.Key(), "} directs packet for backbone connection key {", key, "} into holding area.")
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

	// Create the packet network netFlow
	dstAddr := s.netFlow.Src
	//srcAddr := network.Address{Addr: *connection.RemoteAddr().(*net.TCPAddr)}
	srcAddr := network.AddressGenerator{}.New(connection.RemoteAddr().String())
	path 	:= network.PathGenerator{}.New("")
	header  := shila.NetFlow{Src: dstAddr, Path: path, Dst: srcAddr }

	for {
		if rawData, _ := tcpip.PacketizeRawData(ingressRaw, Config.SizeReadBuffer); rawData != nil { // TODO: Handle error
			if iPHeader, err := shila.GetIPFlow(rawData); err != nil {
				panic(fmt.Sprint("Unable to get IP netFlow in packetizer of server {", s.Key(),
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

	// Fetch the parts for the packet network netFlow which are fixed.
	path 		:= network.PathGenerator{}.New("")
	srcAddr 	:= network.AddressGenerator{}.New(connection.RemoteAddr().String())
	localAddr 	:= connection.LocalAddr().(*net.TCPAddr)

	for {
		if rawData, _ := tcpip.PacketizeRawData(ingressRaw, Config.SizeReadBuffer); rawData != nil { // TODO: Handle error
			if iPHeader, err := shila.GetIPFlow(rawData); err != nil {
				panic(fmt.Sprint("Unable to get IP netFlow in packetizer of server {", s.Key(),
					"}. - ", err.Error())) // TODO: Handle panic!
			} else {
				dstAddr := network.AddressGenerator{}.New(net.JoinHostPort(localAddr.IP.String(), strconv.Itoa(iPHeader.Dst.Port)))
				header  := shila.NetFlow{Src: dstAddr, Path: path, Dst: srcAddr}
				s.ingress <- shila.NewPacketWithNetFlow(s, iPHeader, header, rawData)
			}
		} else {
			return // Raw ingress channel closed.
		}
	}
}
