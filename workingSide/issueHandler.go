package workingSide

import (
	"shila/core/shila"
	"shila/networkSide/networkEndpoint"
	"shila/shutdown"
)

// This issue handler handles all issues related to a connection.
func (m *Manager) issueHandler() {
	for issue := range m.endpointIssues {
		// Handle issues sent from the kernel endpoint
		var ep interface{} = issue.Issuer
		if server, ok := ep.(*networkEndpoint.Server); ok {
			m.handleServerNetworkEndpointIssues(server, issue)
			return
		} else if client, ok := ep.(*networkEndpoint.Client); ok {
			m.handleNetworkClientIssue(client, issue)
			return
		}
		shutdown.Fatal(shila.CriticalError("Should not happen."))
	}
}

func (m *Manager) handleServerNetworkEndpointIssues(server shila.NetworkServerEndpoint, issue shila.EndpointIssuePub) {

	if server.Role() == shila.TrafficNetworkEndpoint {
		var err interface{} = issue.Error
		if _, ok := err.(*networkEndpoint.ConnectionError); ok {
			m.connections.Close(issue.Key, issue.Error)
			return
		}
	}
	shutdown.Fatal(shila.CriticalError("Should not happen."))
}

func (m *Manager) handleNetworkClientIssue(client shila.NetworkClientEndpoint, issue shila.EndpointIssuePub) {

	if  client.Role() == shila.TrafficNetworkEndpoint ||
		client.Role() == shila.ContactNetworkEndpoint {

		var err interface{} = issue.Error
		if _, ok := err.(*networkEndpoint.ConnectionError); ok {
			m.connections.Close(issue.Key, issue.Error)
			return
		}
	}
	shutdown.Fatal(shila.CriticalError("Should not happen."))
}
