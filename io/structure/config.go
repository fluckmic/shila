package structure

type ConfigJSON struct {
	WorkingSide			WorkingSideConfigJSON
	Connection			ConnectionConfigJSON
	NetFlow				NetFlowConfigJSON
	KernelSide			KernelSideConfigJSON
	KernelEndpoint		KernelEndpointConfigJSON
	Logging				LoggingConfigJSON
	NetworkSide			NetworkSideConfigJSON
	NetworkEndpoint		NetworkEndpointConfigJSON
	Router				RouterConfigJSON
	Config				ConfigConfigJSON
}

type WorkingSideConfigJSON struct {
	NumberOfWorkerPerChannel 			int					// Number of worker per packet channel.
}

type ConnectionConfigJSON struct {
	VacuumInterval 	 					int 				// Minimal amount of time between two vacuum processes.
	MaxTimeUntouched 					int					// Maximal amount of time a connection can stay untouched before it is closed.

	WaitingTimeTrafficConnEstablishment int					// Minimal waiting time before a connection establishment with
															// the traffic server endpoint is attempted.
}

type NetFlowConfigJSON struct {
	Path 								string				// Path from where to load the routing entries inserted at startup.
}

type KernelSideConfigJSON struct {
	NumberOfEgressInterfaces 			int					// Number of the egress virtual interfaces.
	EgressNamespace          			string				// The name of the egress namespace.
	IngressNamespace         			string				// The name of the ingress namespace.
	IngressIP                			string				// The IP of the ingress virtual interface.
}

type KernelEndpointConfigJSON struct {
	SizeIngressBuffer 					int					// Size (shila packets) of the ingress buffer.
	SizeEgressBuffer  					int					// Size (shila packets) of the egress buffer.
	SizeRawIngressBuffer 				int					// Size (bytes) of the raw ingress buffer.
	SizeRawIngressStorage				int					// Size (bytes) of the storage holding raw ingress data.
	ReadSizeRawIngress    				int					// Minimal number of bytes read from the raw ingress channel at once.
	WaitingTimeUntilEscalation			int					// Time to wait until a kernel endpoint escalates after a connection to a tun device has been lost.
}

type LoggingConfigJSON struct {
	PrintVerbose 						bool				// Print verbose messages.
	DoEgressTimestamping				bool				// Generates and logs a timestamp before a packet is sent out (Just if packet has specific load.)
	DoIngressTimestamping				bool 				// Generates and logs a timestamp right after a packet was received (Just if packet has specific load.)
	TimestampFlushInterval				int					// The timestamp logs are flushed to file at least in this interval (or when the buffer is full)
	EgressTimestampLogPath				string				// Where to dump the log files for the egress timestamps.
	IngressTimestampLogPath				string				// Where to dump the log files for the ingress timestamps.
	EgressTimestampLogAdditionalLine	string				// Additional line which is added to the egress timestamp log.
	IngressTimestampLogAdditionalLine	string				// Additional line which is added to the ingress timestamp log.
}

type ConfigConfigJSON struct {
	DumpConfig							bool				// Dumps the complete configuration of shila upon start up.
	ConfigDumpPath						string				// Where to dump the config dump.
}

type NetworkSideConfigJSON struct {
	ContactingServerPort           	 	int					// Default port on which shila is listening for incoming contacting connections.
}

type NetworkEndpointConfigJSON struct {
	SizeIngressBuffer               	int           		// Size (shila packets) of the Ingress buffer.
	SizeEgressBuffer               		int           		// Size (shila packets) of the Egress buffer.
	SizeRawIngressStorage          	 	int           		// Size (bytes) of the storage holding raw Ingress data.
	WaitingTimeAfterConnectionIssue	 	int 				// Time to wait after a connection issue has occurred.
	ServerResendInterval           	 	int 				// Time to wait until a server endpoint tries to resend a packet.
	SizeHoldingArea                	 	int           		// Initial size (shila packets) of the holding area.
}

type RouterConfigJSON struct {
	PathSelection 						string				// What type to use for the path selection. (mtu, shortest)
}