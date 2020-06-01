package networkEndpoint

import "time"

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {
	SizeIngressBuffer               int           // Size (shila packets) of the ingress buffer.
	SizeEgressBuffer                int           // Size (shila packets) of the egress buffer.
	SizeRawIngressBuffer            int           // Size (bytes) of the raw ingress buffer.
	SizeRawIngressStorage           int           // Size (bytes) of the storage holding raw ingress data.
	ReadSizeRawIngress              int           // Minimal number of bytes read from the raw ingress channel at once.
	WaitingTimeUntilConnectionRetry time.Duration // Time to wait until a client endpoint tries to reconnect after a established connection has failed.
	ServerResendInterval            time.Duration // Time to wait until a server endpoint tries to resend a packet.
	SizeHoldingArea                 int           // Initial size (shila packets) of the holding area.
}

func hardCodedConfig() config {
	return config{
		SizeIngressBuffer:               10,
		SizeEgressBuffer:                10,
		SizeRawIngressBuffer:            500,
		SizeRawIngressStorage:           1500,
		ReadSizeRawIngress:              30,
		WaitingTimeUntilConnectionRetry: time.Second * 2,
		ServerResendInterval:            time.Second * 2,
		SizeHoldingArea:                 100,
	}
}

