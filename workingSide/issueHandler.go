package workingSide

import (
	"shila/core/shila"
	"shila/networkSide/networkEndpoint"
	"shila/shutdown"
)

// This issue handler handles all issues related to a connection.
func (manager *Manager) issueHandler() {
	for issue := range manager.endpointIssues {
		// Handle issues sent from the kernel endpoint
		var ep interface{} = issue.Issuer
		if server, ok := ep.(*networkEndpoint.Server); ok {
			manager.handleServerNetworkEndpointIssues(server, issue)
			return
		} else if client, ok := ep.(*networkEndpoint.Client); ok {
			manager.handleNetworkClientIssue(client, issue)
			return
		}
		shutdown.Fatal(shila.CriticalError("Received issue from unhandled endpoint type. Should not happen."))
	}
}

func (manager *Manager) handleServerNetworkEndpointIssues(server shila.NetworkServerEndpoint, issue shila.EndpointIssuePub) {

	if server.Role() == shila.TrafficNetworkEndpoint {
		var err interface{} = issue.Error
		if _, ok := err.(*networkEndpoint.ConnectionError); ok {
			manager.connections.Close(issue.Key, issue.Error)
			return
		}
	}
	shutdown.Fatal(shila.CriticalError("Received issue from endpoint with unhandled role. Should not happen."))
}

func (manager *Manager) handleNetworkClientIssue(client shila.NetworkClientEndpoint, issue shila.EndpointIssuePub) {

	if  client.Role() == shila.TrafficNetworkEndpoint ||
		client.Role() == shila.ContactNetworkEndpoint {

		var err interface{} = issue.Error
		if _, ok := err.(*networkEndpoint.ConnectionError); ok {
			manager.connections.Close(issue.Key, issue.Error)
			return
		}
	}
	shutdown.Fatal(shila.CriticalError("Received issue from endpoint with unhandled role. Should not happen."))
}