package networkEndpoint

import (
	"fmt"
	"net"
	"shila/config"
	"shila/core/model"
	"shila/log"
)

var _ model.ServerNetworkEndpoint = (*Server)(nil)

type Server struct{
	Base
	listenTo 	Address
	isRunning 	bool
}

func newServer(listenTo model.NetworkAddress, label model.EndpointLabel, config config.NetworkEndpoint) model.ServerNetworkEndpoint {
	return &Server{Base{label, model.TrafficChannels{}, config},listenTo.(Address), false}
}

func (s *Server) SetupAndRun() error {

	if s.IsRunning() {
		return Error(fmt.Sprint("Unable to setup and run server {", s.Label(), " ", s.Key(), "}. - Server is already running."))
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

func (s *Server) handleConnection(connection net.Conn) {

}
