package connection

import (
	"fmt"
	"shila/core/model"
	"shila/kernelSide"
	"shila/kernelSide/kernelEndpoint"
	"shila/layer"
	"shila/log"
	"shila/networkSide"
	"sync"
	"time"
)

type Connection struct {
	ipConnIdKey model.IPConnectionIdentifierKey
	netConnId   model.NetworkConnectionIdentifier
	state       state
	channels    channels
	lock        sync.Mutex
	touched     time.Time
	sent        bool
	received    bool
	kernelSide  *kernelSide.Manager
	networkSide *networkSide.Manager
	routing     *model.Mapping
}

type channels struct {
	KernelEndpoint  model.PacketChannels // Kernel end point
	NetworkEndpoint model.PacketChannels // Network end point
	Contacting      model.PacketChannels // End point for connection establishment
}

func New(kernelSide *kernelSide.Manager, networkSide *networkSide.Manager, routing *model.Mapping,
	ipId model.IPConnectionIdentifierKey) *Connection {
	return &Connection{
		ipConnIdKey: ipId,
		state:       state{Raw, Raw},
		lock:        sync.Mutex{},
		touched:     time.Now(),
		kernelSide:  kernelSide,
		networkSide: networkSide,
		routing:     routing,
	}
}

func (conn *Connection) Close() {
	conn.lock.Lock()
	defer conn.lock.Unlock()

	// Tear down all endpoints possibly associated with this connection
	// TODO: Rethink
	_ = conn.networkSide.TeardownContactingClientEndpoint(conn.netConnId.Dst)
	_ = conn.networkSide.TeardownTrafficSeverEndpoint(conn.netConnId.Src)
	_ = conn.networkSide.TeardownTrafficClientEndpoint(conn.netConnId.Dst, conn.netConnId.Path)

	conn.state.Set(Closed)
}

func (conn *Connection) ProcessPacket(p *model.Packet) error {

	conn.lock.Lock()
	defer conn.lock.Unlock()

	log.Verbose.Print("Connection {", conn.ipConnIdKey, "} in state {", conn.state.Current(), "} " +
		"starts processing packet from {", p.GetIPConnId().Src.IP, "} to {", p.GetIPConnId().Dst.IP, "}.")

	if p.IPConnIdKey() != conn.ipConnIdKey {
		panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet " +
			"{", p.IPConnIdKey(), "}. - Id mismatch.")) // TODO: Handle panic!
		return nil
	}

	// From where was the packet received?
	var err error
	switch p.GetEntryPoint().Label() {
		case model.KernelEndpoint: 				err = conn.processPacketFromKerep(p)
		case model.ContactingNetworkEndpoint: 	err = conn.processPacketFromContactingEndpoint(p)
		case model.TrafficNetworkEndpoint:		err = conn.processPacketFromTrafficEndpoint(p)
		default:
			panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet " +  // TODO: Handle panic!
			"{", p.IPConnIdKey(), "}. - Unknown entry point label {", p.GetEntryPoint().Label(), "}."))
			return nil
	}

	log.Verbose.Print("Connection {", conn.ipConnIdKey, "} done with processing packet from {", p.GetIPConnId().Src.IP, "} to {",
		p.GetIPConnId().Dst.IP,"}; ending up in state {", conn.state.Current(),"}.")

	return err
}

func (conn *Connection) processPacketFromKerep(p *model.Packet) error {
	switch conn.state.Current() {
	case Raw:				return conn.processPacketFromKerepStateRaw(p)

	case ClientReady:		p.SetNetworkConnId(conn.netConnId)
							conn.touched = time.Now()
							conn.channels.Contacting.Egress <- p
							conn.state.Set(ClientReady)
							return nil

	case ServerReady:		// Put packet into egress queue of connection. If the connection is established at one one point, these packets
							// are sent. If not they are lost. (--> Take care, could block if too many packets are in queue
							p.SetNetworkConnId(conn.netConnId)
							conn.channels.NetworkEndpoint.Egress <- p
							conn.state.Set(ServerReady)
							return nil

	case ClientEstablished:	p.SetNetworkConnId(conn.netConnId)
							conn.touched = time.Now()
							conn.channels.NetworkEndpoint.Egress <- p
							conn.state.Set(ClientEstablished)
							return nil

	case Established:		p.SetNetworkConnId(conn.netConnId)
							conn.touched = time.Now()
							conn.channels.NetworkEndpoint.Egress <- p
							conn.state.Set(Established)
							return nil

	case Closed: 			log.Info.Println("Drop packet - Sent through closed connection.")
							conn.state.Set(Closed)
							return nil

	default: 				panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet " +
							"{", p.IPConnIdKey(), "}. - Unknown connection state.")) // TODO: Handle panic!
							return nil
	}
}

func (conn *Connection) processPacketFromContactingEndpoint(p *model.Packet) error {
	switch conn.state.Current() {

	case Raw:				return conn.processPacketFromContactingEndpointStateRaw(p)

	case ClientReady:		log.Info.Println("Drop packet - Both endpoints in client state.")
							conn.state.Set(ClientReady)
							return nil

	case ServerReady:		conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.state.Set(ServerReady)
							return nil

	case Established: 		conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.state.Set(Established)
							return nil

	case Closed: 			log.Info.Println("Drop packet - Sent through closed connection.")
							conn.state.Set(Closed)
							return nil

	default: 				panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet " +
							"{", p.IPConnIdKey(), "}. - Unknown connection state.")) // TODO: Handle panic!
							return nil
	}
}

func (conn *Connection) processPacketFromTrafficEndpoint(p *model.Packet) error {
	switch conn.state.Current() {

	case Raw:				conn.state.Set(Raw)
							panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet " + // TODO: Handle panic!
							"{", p.IPConnIdKey(), "}. - Invalid connection state {", conn.state.Current(), "}."))
							return nil

							// A packet from the traffic endpoint is just received if network connection is established
	case ClientReady:		conn.state.Set(ClientReady)
							panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet " + // TODO: Handle panic!
								"{", p.IPConnIdKey(), "}. - Invalid connection state {", conn.state.Current(), "}."))
							return nil

	case ServerReady:		conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.state.Set(Established)
							return nil

	case ClientEstablished: // The very first packet received through the traffic endpoint holds the MPTCP receiver key
							// which we need later to be able to get the network destination address for the sub flows.
							if key, ok, err := layer.GetMPTCPSenderKey(p.GetRawPayload()); ok {
								if err == nil {
									if err := conn.routing.InsertFromMPTCPEndpointKey(key,
										conn.netConnId.Src, conn.netConnId.Dst, conn.netConnId.Path); err != nil {
										panic(fmt.Sprint("Error in fetching receiver key in connection {",
										conn.ipConnIdKey, "}. - ", err.Error())) // TODO: Handle panic!
										return nil
									}
								} else {
									panic(fmt.Sprint("Error in fetching receiver key in connection {", conn.ipConnIdKey,
									"} beside having packet {", p.GetIPConnId(), "} containing it. - ", err.Error())) // TODO: Handle panic!
									return nil
								}
							}

							conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.state.Set(Established)
							return nil

	case Established: 		conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.state.Set(Established)
							return nil

	case Closed: 			log.Info.Println("Drop packet - Sent through closed connection.")
							conn.state.Set(Closed)
							return nil

	default: 				panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet " +
							"{", p.IPConnIdKey(), "}. - Unknown connection state.")) // TODO: Handle panic!
							return nil
	}
}

func (conn *Connection) processPacketFromKerepStateRaw(p *model.Packet) error {

	// Assign the channels from the device through which the packet was received.
	var ep interface{} = p.GetEntryPoint()
	if entryPoint, ok := ep.(*kernelEndpoint.Device); ok {
		conn.channels.KernelEndpoint.Ingress = entryPoint.TrafficChannels().Ingress // ingress from kernel end point
		conn.channels.KernelEndpoint.Egress  = entryPoint.TrafficChannels().Egress  // egress towards kernel end point
	} else {
		conn.state.Set(Closed)
		panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet " +
		"{", p.IPConnIdKey(), "}. - Invalid entry point.")) // TODO: Handle panic!
		return nil
	}

	// Create the packet header which is associated with the connection
	// If the packet contains a receiver token, then the new connection is a MPTCP sub flow.
	if token, ok, err := layer.GetMPTCPReceiverToken(p.GetRawPayload()); ok {
		if err == nil {
			if netConnId, ok := conn.routing.RetrieveFromMPTCPEndpointToken(token); ok {
				conn.netConnId = netConnId
				p.SetNetworkConnId(netConnId)
			} else {
				panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet " +
				"{", p.IPConnIdKey(), "}. - No network connection id available in routing for receiver token {", token,
				"} beside having packet {", p.GetIPConnId(), "} containing it.")) // TODO: Handle panic!
				return nil
			}
		} else {
			panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet " +
			"{", p.IPConnIdKey(), "}. - Error while fetching receiver token. - ", err.Error())) // TODO: Handle panic!
			return nil
		}
	// For a MPTCP main flow the network header can probably be extracted from the IP options
	} else if netConnId, ok, err := layer.GetNetworkHeaderFromIPOptions(p.GetRawPayload()); ok {
		if err == nil {
			conn.netConnId = netConnId
			p.SetNetworkConnId(netConnId)
		} else {
			panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet {", p.IPConnIdKey(),
			"}. - Unable to get IP options. - ", err.Error())) // TODO: Handle panic!
			return nil
		}
	// For a MPTCP main flow the network header is probably available in the routing table
	} else if netConnId, ok := conn.routing.RetrieveFromIPAddressPortKey(p.IPConnIdDstKey()); ok {
		conn.netConnId = netConnId
		p.SetNetworkConnId(netConnId)
	// No valid option to get network header :(
	} else {
		panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet {", p.IPConnIdKey(),
		"}. - No valid option to create network connection identifier.")) // TODO: Handle panic!
		return nil
	}

	// Create the contacting connection
	contactingNetConnId, channels, err := conn.networkSide.EstablishNewContactingClientEndpoint(conn.netConnId)
	if err != nil {
		conn.state.Set(Closed)
		panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet {", p.IPConnIdKey(),
		"}. - Unable to establish contacting client endpoint. - ", err.Error())) // TODO: Handle panic!
		return nil
	}
	conn.channels.Contacting = channels

	// The contacting network connection id contains as src the local network endpoint which
	// was used by the network to established the contacting connection. Once a traffic connection
	// is established, we send this information to the corresponding server side.
	conn.netConnId.Src = contactingNetConnId.Src

	// Send the packet via the contacting channel
	conn.touched = time.Now()
	conn.channels.Contacting.Egress <- p

	// Try to connect to the address via path, a corresponding server should be there listening
	go func() {
		if trafficNetConnId, channels, err := conn.networkSide.EstablishNewTrafficClientEndpoint(conn.netConnId); err != nil {
			conn.Close()
			log.Info.Print(fmt.Sprint("Connection {", conn.ipConnIdKey, "} was not able to establish a traffic" +
			"backbone connection to {", conn.netConnId.Dst, " via ", conn.netConnId.Path, "}. - ", err.Error()))
		} else {
			conn.netConnId 				  = trafficNetConnId
			conn.channels.NetworkEndpoint = channels
			conn.lock.Lock()
			defer conn.lock.Unlock()
			conn.state.Set(ClientEstablished)
			// The contacting client endpoint is no longer needed.
			_ = conn.networkSide.TeardownContactingClientEndpoint(conn.netConnId.Dst)
			log.Info.Print(fmt.Sprint("Connection {", conn.ipConnIdKey, "} was able to establish a traffic" +
			"backbone connection to {", conn.netConnId.Dst, " via ", conn.netConnId.Path, "}."))
		}
	}()

	// Set new state
	conn.state.Set(ClientReady)

	return nil
}

func (conn *Connection) processPacketFromContactingEndpointStateRaw(p *model.Packet) error {

	// Get the kernel endpoint from the kernel side manager
	dstKey := p.IPConnIdDstIPKey()
	if channels, ok := conn.kernelSide.GetTrafficChannels(dstKey); ok {
		conn.channels.KernelEndpoint = channels
	} else {
		conn.state.Set(Closed)
		panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet {", p.IPConnIdKey(),
		"}. - No kernel endpoint {", dstKey, "}.")) // TODO: Handle panic!
		return nil
	}

	// Send packet to kernel endpoint
	// --> 	Could still be that connection cannot be established, since we have no idea if there is actually a server listening
	conn.channels.KernelEndpoint.Egress <- p

	// If the packet is received through the contacting endpoint (server), then it's network connection id
	// is already set. This is the responsibility of the corresponding network server implementation.
	conn.netConnId = p.GetNetworkConnId()

	// Request new incoming connection from network side.
	// ! The receiving network endpoint is responsible to correctly set the destination network address! !
	if channels, err := conn.networkSide.EstablishNewTrafficServerEndpoint(conn.netConnId); err != nil {
		conn.state.Set(Closed)
		panic(fmt.Sprint("Connection {", conn.ipConnIdKey, "} can not process packet {", p.IPConnIdKey(),
		"}. - Unable to establish server endpoint. -", err.Error())) // TODO: Handle panic!
		return nil
	} else {
		conn.channels.NetworkEndpoint = channels
	}

	// Set new state
	conn.state.Set(ServerReady)

	return nil
}