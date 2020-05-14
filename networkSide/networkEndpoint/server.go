package networkEndpoint

import (
	"fmt"
	"io"
	"net"
	"shila/config"
	"shila/core/model"
	"shila/layer"
	"shila/log"
	"shila/shutdown"
)

var _ model.ServerNetworkEndpoint = (*Server)(nil)

type Server struct{
	Base
	connections map[model.NetworkAddressAndPathKey]  net.Conn
	listenTo 	Address
	isValid     bool
	isSetup     bool // TODO: merge to "state" object
	isRunning   bool
}

func newServer(listenTo model.NetworkAddress, label model.EndpointLabel, config config.NetworkEndpoint) model.ServerNetworkEndpoint {
	return &Server{
		Base: Base{label, model.TrafficChannels{},config},
		connections: make(map[model.NetworkAddressAndPathKey]  net.Conn),
		listenTo: listenTo.(Address),
		isValid: true,
	}
}

func (s *Server) SetupAndRun() error {

	if s.IsRunning() {
		return Error(fmt.Sprint("Unable to setup and run server {", s.Label(), " ", s.Key(), "}. - Server is already running."))
	}

	if !s.IsValid() {
		return Error(fmt.Sprint("Unable to setup and run server {", s.Label()," ",s.Key(), "}. - Server no longer valid."))
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
	s.trafficChannels.Ingress = make(chan *model.Packet, s.config.SizeIngressBuff)
	s.trafficChannels.Egress  = make(chan *model.Packet, s.config.SizeEgressBuff)
	s.trafficChannels.Label   = s.Label()
	s.trafficChannels.Key	  = s.Key()

	// Start listening for incoming connections.
	go s.serveIncomingConnections(listener)

	log.Verbose.Print("Server {", s.Label(), " ", s.Key(), "} started to listen for incoming connections.")

	s.isSetup   = true
	s.isRunning = true
	return nil
}

func (s *Server) TearDown() error {
	return nil
}

func (s *Server) TrafficChannels() model.TrafficChannels {
	return s.trafficChannels
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
			log.Info.Print("Error serving incoming connection in server {", s.Label(), " ", s.Key(), "}. - ", err.Error())
		} else {
			go s.handleConnection(connection)
		}
	}
}

func (s *Server) handleConnection(connection net.Conn) {

	// Get the address from the client side
	srcAddr := Generator{}.NewAddress(connection.RemoteAddr().String())
	// Get the path taken from client to this server
	path := Generator{}.NewPath("")

	// Generate the key
	key := Generator{}.GetAddressPathKey(srcAddr, path)

	// Add the new connection to the mapping, such that it can be found by the egress handler.
	if _, ok := s.connections[key]; ok {
		panic(fmt.Sprint("Trying to add connection with key {", key, "} in " +
			"server {", s.Key(), "}. There already exists a connection with that key.")) // TODO: Handle panic!
	} else {
		s.connections[key] = connection
	}

	// Start the ingress handler for the connection.
	go s.serveIngress(connection)
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
		go s.packetizeTraffic(ingressRaw, connection)
	} else {
		panic(fmt.Sprint("Wrong server label {", s.Label(), "} in serving ingress functionality of " +
			"server {", s.Key(), "}.")) // TODO: Handle panic!
	}

	reader := io.Reader(connection)
	storage := make([]byte, s.config.SizeReadBuffer)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, s.config.BatchSizeRead)
		if err != nil && !s.IsValid() {
			// Error doesn't matter, kernel endpoint is no longer valid anyway.
			return
		} else if err != nil {
			panic(fmt.Sprint("Error in reading data in server {", s.Key(), "}. - ", err.Error())) // TODO: Handle panic!
		}
		for _, b := range storage[:nBytesRead] {
			ingressRaw <- b
		}
	}
}

func (s *Server) serveEgress() {
	for p := range s.trafficChannels.Egress {
		// Retrieve key to get the correct connection
		key := p.NetworkHeaderDstAndPathKey()
		if con, ok := s.connections[key]; ok {
			writer := io.Writer(con)
			_, err := writer.Write(p.GetRawPayload())
			if err != nil && !s.IsValid() {
				// Error doesn't matter, kernel endpoint is no longer valid anyway.
				return
			} else if err != nil {
				panic(fmt.Sprint("Unable to send data for packet {", p.IPHeaderKey(), "} in the " +
					"server {", s.Key(), "} for connection key {", key,"}. - ", err.Error())) // TODO: Handle panic!
			}
		} else {
			panic(fmt.Sprint("Can't find an egress connection for packet {", p.IPHeaderKey(), "} in the " +
				"server {", s.Key(), "} for connection key {", key,"}.")) // TODO: Handle panic!
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
	srcAddr := Generator{}.NewAddress(connection.RemoteAddr().String())
	path 	:= Generator{}.NewPath("")
	header  := model.NetworkHeader{Src: srcAddr, Path: path, Dst: dstAddr }

	for {
		rawData  := layer.PacketizeRawData(ingressRaw, s.config.SizeReadBuffer)
		if iPHeader, err := layer.GetIPHeader(rawData); err != nil {
			panic(fmt.Sprint("Unable to get IP header in packetizer of server {", s.Key(),
				"}. - ", err.Error())) // TODO: Handle panic!
		} else {
			s.trafficChannels.Ingress <- model.NewPacketInclNetworkHeader(s, iPHeader, header, rawData)
		}
	}
}

func (s *Server) packetizeContacting(ingressRaw chan byte, connection net.Conn) {

	/*
		dstAddr 	:= s.listenTo
		path        := Generator{}.NewPath("")
		srcNetwAddr := connection.RemoteAddr()

		for {
			rawData := packetizeRawData(ingressRaw, s.config.SizeReadBuffer)
			if ip4v, tcp, err := layer.DecodeIPv4andTCPLayer(rawData); err != nil {

			} else {
				p.ipHeader.Src.IP 		= ip4v.SrcIP
				p.ipHeader.Src.Port 	= int(tcp.SrcPort)
				p.ipHeader.Dst.IP 		= ip4v.DstIP
				p.ipHeader.Dst.Port 	= int(tcp.DstPort)
			}


			s.trafficChannels.Ingress <- model.NewPacketFromRawIPAndNetworkHeader(s, &header, rawData)
		}
	*/
	shutdown.Check() // Fatal error could occur.. :o

}
