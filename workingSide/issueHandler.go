package workingSide

import (
	"fmt"
	"github.com/scionproto/scion/go/lib/snet"
	"shila/core/shila"
	"shila/log"
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
		if connErr, ok := err.(networkEndpoint.ConnectionError); ok {
			log.Error.Print(server.Identifier(), " triggered connection error: ", connErr.Error())
			manager.connections.Close(issue.Key, issue.Error)
			return
		} else if opErr, ok := err.(*snet.OpError); ok {
			log.Error.Print(server.Identifier(), " received op error: ", opErr.Error())
			fmt.Printf("SCMP header: %v\n", opErr.SCMP())
			manager.connections.Close(issue.Key, issue.Error)
			return
		}
		shutdown.Fatal(shila.CriticalError(fmt.Sprint("Received unknown issue from ", server.Identifier())))
	}
	shutdown.Fatal(shila.CriticalError(fmt.Sprint("Received issue from endpoint with unhandled role: ", server.Identifier())))
}

func (manager *Manager) handleNetworkClientIssue(client shila.NetworkClientEndpoint, issue shila.EndpointIssuePub) {

	if  client.Role() == shila.TrafficNetworkEndpoint ||
		client.Role() == shila.ContactNetworkEndpoint {

		var err interface{} = issue.Error
		if connErr, ok := err.(networkEndpoint.ConnectionError); ok {
			log.Error.Print(client.Identifier(), " triggered connection error: ", connErr.Error())
			manager.connections.Close(issue.Key, issue.Error)
			return
		} else if opErr, ok := err.(*snet.OpError); ok {
			log.Error.Print(client.Identifier(), " received op error: ", opErr.Error())
			fmt.Printf("SCMP header: %v\n", opErr.SCMP())
			manager.connections.Close(issue.Key, issue.Error)
			return
		}
		shutdown.Fatal(shila.CriticalError(fmt.Sprint("Received unknown issue from: ", client.Identifier())))
	}
	shutdown.Fatal(shila.CriticalError(fmt.Sprint("Received issue from endpoint with unhandled role: ", client.Identifier())))
}