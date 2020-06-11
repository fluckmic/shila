package networkEndpoint

import "time"

var Config config

func init() {
	Config 		 = hardCodedConfig()
}

type config struct {
	SizeIngressBuffer               int           // Size (shila packets) of the Ingress buffer.
	SizeEgressBuffer                int           // Size (shila packets) of the Egress buffer.
	SizeRawIngressStorage           int           // Size (bytes) of the storage holding raw Ingress data.
	WaitingTimeAfterConnectionIssue time.Duration // Time to wait after a connection issue has occured.s
	ServerResendInterval            time.Duration // Time to wait until a server endpoint tries to resend a packet.
	SizeHoldingArea                 int           // Initial size (shila packets) of the holding area.
}

func hardCodedConfig() config {
	return config{
		SizeIngressBuffer:               10,
		SizeEgressBuffer:                10,
		SizeRawIngressStorage:           1500,
		WaitingTimeAfterConnectionIssue: time.Second * 2,
		ServerResendInterval:            time.Second * 2,
		SizeHoldingArea:                 100,
	}
}

