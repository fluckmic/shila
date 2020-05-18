package networkEndpoint

import (
	"fmt"
	"io"
	"net"
	"shila/config"
	"shila/core/model"
	"shila/layer"
	"shila/log"
	"strconv"
	"sync"
)

var _ model.ServerNetworkEndpoint = (*Server)(nil)

type Server struct{
	Base
	connections map[model.NetworkAddressAndPathKey]  net.Conn
	lock        sync.Mutex
	header 		model.NetworkConnectionTriple
	listenTo 	Address
	holdingArea []*model.Packet
	isValid     bool
	isSetup     bool // TODO: merge to "state" object
	isRunning   bool
}

func newServer(listenTo model.NetworkAddress, label model.EndpointLabel, config config.NetworkEndpoint) model.ServerNetworkEndpoint {
	return &Server{
		Base: 			Base{
								label: 	label,
								config: config,
						},
		connections: 	make(map[model.NetworkAddressAndPathKey]  net.Conn),
		lock: 			sync.Mutex{},
		listenTo: 		listenTo.(Address),
		holdingArea:    make([]*model.Packet, 0, config.SizeHoldingArea),
		isValid: 		true,
	}
}

func (s *Server) SetupAndRun() error {

	if s.IsRunning() {
		return Error(fmt.Sprint("Unable to setup and run server {", s.Label(), " ", s.Key(), "}. - Server is already running."))
	}

	if !s.IsValid() {
		return Error(fmt.Sprint("Unable to setup and run server {", s.Label(), " ", s.Key(), "}. - Server no longer valid."))
	}

	if s.IsSetup() {
		return nil
	}


	// Set up the listener
	listener, err := net.ListenTCP(s.listenTo.Addr.Network(), &s.listenTo.Addr)
	if err != nil {
		return Error(fmt.Sprint("Unable to setup and run server {", s.Label(), " ", s.Key(), "}. - ", err.Error()))
	}

	// Create the channels
	s.ingress = make(chan *model.Packet, s.config.SizeIngressBuff)
	s.egress  = make(chan *model.Packet, s.config.SizeEgressBuff)

	// Start listening for incoming connections.
	go s.serveIncomingConnections(listener)

	log.Verbose.Print("Server {",s.Label()," ",s.Key(),"} started to listen for incoming backbone connections on {", s.Key(), "}.")

	// Start to handle incoming packets
	go s.serveEgress()

	s.isSetup   = true
	s.isRunning = true
	return nil
}

func (s *Server) TearDown() error {
	// TODO!
	return nil
}

func (s *Server) TrafficChannels() model.PacketChannels {
	return model.PacketChannels{Ingress: s.ingress, Egress: s.egress}
}

func (s *Server) Label() model.EndpointLabel {
	return s.label
}

func (s *Server) Key() model.EndpointKey {
	return model.EndpointKey(Generator{}.GetAddressKey(s.listenTo))
}

func (s *Server) IsRunning() bool {
	return s.isRunning
}

func (s *Server) serveIncomingConnections(listener net.Listener){
	for {
		if connection, err := listener.Accept(); err != nil {
			log.Info.Print("Error serving incoming backbone connection in server {" ,s.Label(), " ", s.Key(), "}. - ",err.Error())
		} else {
			go s.handleConnection(connection)
		}
	}
}

func (s *Server) handleConnection(connection net.Conn) {

	// Get the address from the client side
	srcAddr := Generator{}.NewAddress(connection.RemoteAddr().String())
	// Get the path taken from client to this server
	path 	:= Generator{}.NewPath("")

	// Generate the key
	key := model.KeyGenerator{}.NetworkAddressAndPathKey(srcAddr, path)

	// Add the new connection to the mapping, such that it can be found by the egress handler.
	s.lock.Lock()
	if _, ok := s.connections[key]; ok {
		s.lock.Unlock()
		panic(fmt.Sprint("Trying to add backbone connection with key {", key, "} in " +
			"server {", s.Label(), " ", s.Key(), "}. There already exists a backbone connection with that key.")) // TODO: Handle panic!
	} else {
		s.connections[key] = connection
		s.lock.Unlock()
	}

	// A new connection was established. This is good news for
	// all packets waiting in the holding area.
	go s.flushHoldingArea()

	// Start the ingress handler for the connection.
	log.Verbose.Print("Server {", s.Label(), " ", s.Key(), "} started handling a new backbone connection {", key, "}.")
	s.serveIngress(connection)

	// No longer necessary or possible to serve the ingress, remove the connection from the mapping.
	s.lock.Lock()
	delete(s.connections, key)
	s.lock.Unlock()

	log.Verbose.Print("Server {", s.Label(), " ", s.Key(), "} removed backbone connection {", key, "}.")

	return
}

func (s *Server) flushHoldingArea() {
	log.Verbose.Print("Server {", s.Label(), " ", s.Key(), "} flushes the holding area.")
	for _, p := range s.holdingArea {
		s.egress <- p
	}
}

func (s *Server) serveIngress(connection net.Conn) {

	// Prepare everything for the packetizer
	ingressRaw := make(chan byte, s.config.SizeReadBuffer)

	if s.Label() == model.ContactingNetworkEndpoint {
		// Server is the contacting server, it is his responsibility
		// to extract the necessary data from the ip packet to be able
		// to set the correct network header.
		go s.packetizeContacting(ingressRaw, connection)

	} else if s.Label() == model.TrafficNetworkEndpoint {
		// Server receives normal traffic, the connection over which the
		// packet was received contains enough information to set
		// the correct network header.

		// TODO: The corresponding traffic client informs the traffic server for whom this connection is.

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
		// The server instance is still able to receive connections as long as it is not
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
		key := p.NetworkHeaderDstAndPathKey()
		if con, ok := s.connections[key]; ok {
			writer := io.Writer(con)
			nBytesWritten, err := writer.Write(p.GetRawPayload())
			if err != nil && !s.IsValid() {
				// Error doesn't matter, kernel endpoint is no longer valid anyway.
				return
			} else if err != nil {
				panic(fmt.Sprint("Unable to send data for packet {", p.IPHeaderKey(), "} in the " +
					"server {", s.Label(), " ", s.Key(), "} for backbone connection key {", key ,"}. - ", err.Error())) // TODO: Handle panic!
			} else {
				log.Verbose.Print("Server {", s.Label()," ", s.Key(), "} wrote {", nBytesWritten, "}.")
			}
		} else {
		// Currently there is no connection available to send the packet, the packet has therefore to wait
		// in the holding area. Whenever a new connection is established all the packets in the holding area
		// are again processed; hopefully they can be send out this time.
			log.Verbose.Print("Server {", s.Label()," ", s.Key(), "} directs packet with backbone connection key {", key, "} in holding area.")
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

	// Create the packet network header
	dstAddr := s.listenTo
	srcAddr := Address{Addr: *connection.RemoteAddr().(*net.TCPAddr)}
	path 	:= Generator{}.NewPath("")
	header  := model.NetworkConnectionTriple{Src: dstAddr, Path: path, Dst: srcAddr }

	for {
		rawData  := layer.PacketizeRawData(ingressRaw, s.config.SizeReadBuffer)
		if iPHeader, err := layer.GetIPHeader(rawData); err != nil {
			panic(fmt.Sprint("Unable to get IP header in packetizer of server {", s.Key(),
				"}. - ", err.Error())) // TODO: Handle panic!
		} else {
			s.ingress <- model.NewPacketInclNetworkHeader(s, iPHeader, header, rawData)
		}
	}
}

func (s *Server) packetizeContacting(ingressRaw chan byte, connection net.Conn) {

	// Fetch the parts for the packet network header which are fixed.
	path 		:= Generator{}.NewPath("")
	srcAddr 	:= Generator{}.NewAddress(connection.RemoteAddr().String())
	localAddr 	:= connection.LocalAddr().(*net.TCPAddr)

	for {
		if rawData  := layer.PacketizeRawData(ingressRaw, s.config.SizeReadBuffer); rawData != nil {
			if iPHeader, err := layer.GetIPHeader(rawData); err != nil {
				panic(fmt.Sprint("Unable to get IP header in packetizer of server {", s.Key(),
					"}. - ", err.Error())) // TODO: Handle panic!
			} else {
				dstAddr := Generator{}.NewAddress(net.JoinHostPort(localAddr.IP.String(), strconv.Itoa(iPHeader.Dst.Port)))
				header  := model.NetworkConnectionTriple{Src: dstAddr, Path: path, Dst: srcAddr}
				s.ingress <- model.NewPacketInclNetworkHeader(s, iPHeader, header, rawData)
			}
		}
	}

}
