//
package kernelEndpoint

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {
	SizeIngressBuffer 		int		// Size (shila packets) of the ingress buffer.
	SizeEgressBuffer  		int		// Size (shila packets) of the egress buffer.
	SizeRawIngressBuffer 	int		// Size (bytes) of the raw ingress buffer.
	SizeRawIngressStorage	int		// Size (bytes) of the storage holding raw ingress data.
	ReadSizeRawIngress    	int		// Minimal number of bytes read from the raw ingress channel at once.
}

func hardCodedConfig() config {
	return config{
		SizeIngressBuffer:     10,
		SizeEgressBuffer:      10,
		SizeRawIngressBuffer:  500,
		SizeRawIngressStorage: 1500,
		ReadSizeRawIngress:    30,
	}
}