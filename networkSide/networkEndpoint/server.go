package networkEndpoint

import (
	"fmt"
	"io"
	"net"
	"shila/config"
	"shila/core/model"
	"shila/log"
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
			log.Info.Print("Error serving incoming connection in Server {", s.Label(), " ", s.Key(), "}. - ", err.Error())
		} else {
			go s.handleConnection(connection)
		}
	}
}

func (s *Server) handleConnection(connection net.Conn) {}

func (s *Server) serveEgress() {
	for p := range s.trafficChannels.Egress {
		// Retrieve key to get the correct connection
		if key, err := p.NetworkHeaderDstAndPathKey(); err == nil {
			if con, ok := s.connections[key]; ok {
				writer := io.Writer(con)
				_, err := writer.Write(p.RawPayload())
				if err != nil && !s.IsValid() {
					// Error doesn't matter, kernel endpoint is no longer valid anyway.
					return
				} else if err != nil {
					packetKey, _ := p.IPHeaderKey()
					panic(fmt.Sprint("Unable to send data for packet {", packetKey,"} in the " +
						"server {", s.Key(), "} for connection key {", key,"}. - ", err.Error())) // TODO: Handle panic!
				}
			} else {
				packetKey, _ := p.IPHeaderKey()
				panic(fmt.Sprint("Can't find an egress connection for packet {", packetKey,"} in the " +
					"server {", s.Key(), "} for connection key {", key,"}.")) // TODO: Handle panic!
			}
		} else {
			panic(fmt.Sprint("Unable to fetch egress connection key in server {", s.Key(), "}.")) // TODO: Handle panic!
		}
	}
}

func (s *Server) IsSetup() bool {
	return s.isSetup
}

func (s *Server) IsValid() bool {
	return s.isValid
}
