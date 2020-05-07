package networkEndpoint

import (
	"shila/config"
	"shila/core/model"
)

var _ model.ServerNetworkEndpoint = (*Server)(nil)

type Server struct{
	listenTo model.NetworkAddress
	Base
}

func (s *Server) Key() model.EndpointKey {
	return model.EndpointKey(Generator{}.GetAddressKey(s.listenTo))
}

func newServer(listenTo model.NetworkAddress, label model.EndpointLabel, config config.NetworkEndpoint) model.ServerNetworkEndpoint {
	return &Server{listenTo, Base{label, model.TrafficChannels{}}}
}

func (s *Server) SetupAndRun() error {
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
