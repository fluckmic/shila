package networkSide

import (
	"shila/core/shila"
	"shila/networkSide/networkEndpoint"
	"shila/shutdown"
)

func (manager *Manager) errorHandler() {
	for issue := range manager.serverEndpointIssues {
		var ep interface{} = issue.Issuer
		if server, ok := ep.(*networkEndpoint.Server); ok {
			if server.Role() == shila.ContactNetworkEndpoint {
				// An issue in the server contact network endpoint. For the moment we just shot down the whole machinery..
				shutdown.Fatal(shila.CriticalError(shila.PrependError(issue.Error, "Error in server contact network endpoint.").Error()))
			} else if server.Role() == shila.TrafficNetworkEndpoint {
				if endpointWrapper, ok := manager.serverTrafficEndpoints[server.Key()]; ok {
					// Publish an issue for every registered connection.
					for tcpFlow, _ := range endpointWrapper.TCPFlowRegister {
						manager.endpointIssues.Ingress <- shila.EndpointIssuePub{ Issuer: server, Key: tcpFlow, Error: issue.Error}
					}
				} else {
					shutdown.Fatal(shila.CriticalError("Unregistered server endpoint publishes issue. Should not happen."))
				}
			} else {
				shutdown.Fatal(shila.CriticalError("Server endpoint unhandled role publishes issue. Should not happen."))
			}
		} else {
			shutdown.Fatal(shila.CriticalError("Endpoint with unhandled type publishes issue. Should not happen."))
		}
	}
}