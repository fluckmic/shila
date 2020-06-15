package kernelSide

import (
	"shila/core/shila"
	"shila/kernelSide/kernelEndpoint"
	"shila/layer"
	"shila/shutdown"
)

func (manager *Manager) errorHandler() {
	for issue := range manager.endpointIssues {
		var ep interface{} = issue.Issuer
		if device, ok := ep.(*kernelEndpoint.Device); ok {
			if  device.Role() == shila.EgressKernelEndpoint ||
				device.Role() == shila.IngressKernelEndpoint {

				var err interface{} = issue.Error
				if errParsed, ok := err.(*kernelEndpoint.ConnectionError); ok {
					shutdown.Fatal(shila.PrependError(errParsed, "Kernel side error."))
				} else if errParsed, ok := err.(*layer.ParsingError); ok {
					shutdown.Fatal(shila.PrependError(errParsed, "Kernel side error."))
				} else {
					shutdown.Fatal(shila.CriticalError("Unhandled kernel endpoint error. Should not happen."))
				}

			} else {
				shutdown.Fatal(shila.CriticalError("Kernel endpoint with unhandled role publishes issue. Should not happen."))
			}
		} else {
			shutdown.Fatal(shila.CriticalError("Endpoint with unhandled type publishes issue. Should not happen."))
		}
	}
}
