//
package workingSide

import (
	"fmt"
	"shila/core/connection"
	"shila/core/shila"
	"shila/kernelSide/kernelEndpoint"
	"shila/log"
	"shila/networkSide/networkEndpoint"
	"shila/shutdown"
)

type Manager struct {
	label 				Label
	connections     	connection.Mapping
	trafficChannelPubs 	shila.PacketChannelPubChannel
	endpointIssues 	   	shila.EndpointIssuePubChannel
}

func New(connections connection.Mapping, trafficChannelPubs shila.PacketChannelPubChannel,
	endpointIssues shila.EndpointIssuePubChannel, label Label) *Manager {
	return &Manager{
		label: 				label,
		connections:     	connections,
		trafficChannelPubs: trafficChannelPubs,
		endpointIssues:    	endpointIssues,
	}
}

func (m *Manager) Setup() error {
	return nil
}

func (m *Manager) Start() error {

	shutdown.Check()

	go m.trafficWorker()
	go m.issueWorker()

	return nil
}

func (m *Manager) CleanUp() { }

func (m *Manager) trafficWorker() {
	for trafficChannelPub := range m.trafficChannelPubs {
		//log.Verbose.Print("Working side {", m.label, "} received announcement for new traffic channel {",
		//	trafficChannelPub.Publisher.Key(), ",", trafficChannelPub.Publisher.Label(), "}.")
		go m.serveTrafficChannel(trafficChannelPub.Channel, Config.NumberOfWorkerPerChannel)
	}
}

func (m *Manager) serveTrafficChannel(buffer shila.PacketChannel, numberOfWorker int) {
	for id := 0; id < numberOfWorker; id++ {
		go m.handleTrafficChannel(buffer)
	}
}

func (m *Manager) handleTrafficChannel(buffer shila.PacketChannel) {
	for p := range buffer {
		m.processTrafficChannel(p)
	}
}

func (m *Manager) processTrafficChannel(p *shila.Packet) {

	// Get the corresponding connection and processes the packet..
	con := m.connections.Retrieve(p.Flow)

	err := con.ProcessPacket(p)

	// Any error leads inevitably to the closing of the connection.
	// All later packet that are processed by the same connection are silently dropped.
	// The closed connection is removed after a while; after its removal a packet is might
	// abel to use the packet without any error.

	switch err := err.(type) {
	case shila.ThirdPartyError: 	log.Info.Print(err.Error())		// Really not our fault.
	case shila.TolerableError:  	log.Info.Panic(err.Error())		// Probably our fault.
	case shila.CriticalError:		log.Error.Panic(err.Error()) 	// Most likely our fault.
	default:						return
	}
}

func (m *Manager) issueWorker() {
	for issue := range m.endpointIssues {

		// Handle issues from the kernel endpoint
		var ep interface{} = issue.Issuer
		if _, ok := ep.(*kernelEndpoint.Device); ok {
			m.handleKernelEndpointIssue(issue)
			return
		} else if server, ok := ep.(*networkEndpoint.Server); ok {
			m.handleServerNetworkEndpointIssues(server, issue)
			return
		} else if client, ok := ep.(*networkEndpoint.Client); ok {
			m.handleNetworkClientIssue(client, issue)
			return
		}

		// Should really not happen..
		log.Error.Panic(fmt.Sprint("Unknown endpoint point label {",
			issue.Issuer.Label(), "} in issue worker."))
	}
}

func (m *Manager) handleKernelEndpointIssue(issue shila.EndpointIssuePub) {
	log.Error.Print("Unhandled kernel endpoint issue in {", issue.Issuer.Key(), "}.")
	shutdown.Fatal(issue.Error)
}

func (m *Manager) handleServerNetworkEndpointIssues(server shila.NetworkServerEndpoint, issue shila.EndpointIssuePub) {

	if server.Label() == shila.ContactingNetworkEndpoint {
		var err interface{} = issue.Error
		if _, ok := err.(*shila.ParsingError); ok {
			panic("Implement me.") // TODO.
			// Try to reestablish the contacting server network endpoint?
		}
		if _, ok := err.(*shila.NetworkEndpointTimeout); ok {
			panic("Implement me.") // TODO.
		}
	}

	if server.Label() == shila.TrafficNetworkEndpoint {
		var err interface{} = issue.Error
		if _, ok := err.(*shila.ParsingError); ok {
			con := m.connections.Retrieve(issue.Flow)
			con.Close(issue.Error)
			return
		}
		if _, ok := err.(*shila.NetworkEndpointTimeout); ok {
			con := m.connections.Retrieve(issue.Flow)
			con.Close(issue.Error)
			return
		}
	}

	log.Error.Print("Unhandled server network endpoint issue in {", server.Key(), "}.")
	shutdown.Fatal(issue.Error)
}

func (m *Manager) handleNetworkClientIssue(client shila.NetworkClientEndpoint, issue shila.EndpointIssuePub) {

	log.Error.Print("Client endpoint issue in {", client.Label(), "} - ", issue.Error.Error())

	// If there is an error in a network client endpoint we just close the associated connection.
	// Since client endpoints are just created through connections, there should always be an associated one.
	con := m.connections.Retrieve(issue.Flow)
	con.Close(issue.Error)
}

