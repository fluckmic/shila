package connection

import (
	"fmt"
	"shila/core/shila"
	"shila/kernelSide"
	"shila/kernelSide/kernelEndpoint"
	"shila/layer"
	"shila/log"
	"shila/networkSide"
	"sync"
	"time"
)

type Connection struct {
	flow        shila.Flow
	state       state
	kind        kind
	channels    channels
	lock        sync.Mutex
	touched     time.Time
	kernelSide  *kernelSide.Manager
	networkSide *networkSide.Manager
	routing     *shila.Mapping
}

type channels struct {
	KernelEndpoint  shila.PacketChannels // Kernel end point
	NetworkEndpoint shila.PacketChannels // Network end point
	Contacting      shila.PacketChannels // End point for connection establishment
}

func New(flow shila.Flow, kernelSide *kernelSide.Manager, networkSide *networkSide.Manager, routing *shila.Mapping) *Connection {
	return &Connection{
		flow:        flow,
		state:       state{raw, raw},
		kind:        unknown,
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
	_ = conn.networkSide.TeardownContactingClientEndpoint(conn.flow.IPFlow)
	_ = conn.networkSide.TeardownTrafficSeverEndpoint(conn.flow)
	_ = conn.networkSide.TeardownTrafficClientEndpoint(conn.flow.IPFlow)

	conn.state.Set(closed)

	log.Verbose.Print("Connection {", conn.flow.IPFlow.Key(), "} closed.")
}

func (conn *Connection) ProcessPacket(p *shila.Packet) error {

	conn.lock.Lock()
	defer conn.lock.Unlock()

	if p.Flow.IPFlow.Key() != conn.flow.IPFlow.Key() {
		return Error(fmt.Sprint("IP flow mismatch."))
	}

	// From where was the packet received?
	var err error
	switch p.Entrypoint.Label() {
		case shila.KernelEndpoint: 				err = conn.processPacketFromKerep(p)
		case shila.ContactingNetworkEndpoint: 	err = conn.processPacketFromContactingEndpoint(p)
		case shila.TrafficNetworkEndpoint:		err = conn.processPacketFromTrafficEndpoint(p)
		default:
			err = Error(fmt.Sprint("Unknown entry point label {", p.Entrypoint.Label(), "}."))
	}

	if err != nil {
		conn.state.Set(closed)
	}

	return err
}

func (conn *Connection) processPacketFromKerep(p *shila.Packet) error {
	switch conn.state.current {
	case raw:				return conn.processPacketFromKerepStateRaw(p)

	case clientReady:		p.Flow.NetFlow = conn.flow.NetFlow
							conn.touched = time.Now()
							conn.channels.Contacting.Egress <- p
							conn.state.Set(clientReady)
							return nil

	case serverReady: 		// Put packet into egress queue of connection. If the connection is established at one one point, these packets
							// are sent. If not they are lost. (--> Take care, could block if too many packets are in queue
							p.Flow.NetFlow = conn.flow.NetFlow
							conn.channels.NetworkEndpoint.Egress <- p
							conn.state.Set(serverReady)
							return nil

	case clientEstablished:	p.Flow.NetFlow = conn.flow.NetFlow
							conn.touched = time.Now()
							conn.channels.NetworkEndpoint.Egress <- p
							conn.state.Set(clientEstablished)
							return nil

	case established:		p.Flow.NetFlow = conn.flow.NetFlow
							conn.touched = time.Now()
							conn.channels.NetworkEndpoint.Egress <- p
							conn.state.Set(established)
							return nil

	case closed: 			conn.state.Set(closed)
							return nil

	default: 				return Error(fmt.Sprint("Unknown connection state."))

	}
}

func (conn *Connection) processPacketFromContactingEndpoint(p *shila.Packet) error {
	switch conn.state.current {

	case raw:				return conn.processPacketFromContactingEndpointStateRaw(p)

	case clientReady:		return Error(fmt.Sprint("Both endpoints in client state."))

	case serverReady:		conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.state.Set(serverReady)
							return nil

	case established: 		conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.state.Set(established)
							return nil

	case closed: 			// Packets sent through a closed connection are dropped.
							conn.state.Set(closed)
						 	return nil

	default: 				return Error(fmt.Sprint("Unknown connection state."))

	}
}

func (conn *Connection) processPacketFromTrafficEndpoint(p *shila.Packet) error {
	switch conn.state.current {

	case raw:				return Error(fmt.Sprint("Invalid connection state."))

							// A packet from the traffic endpoint is just received if network connection is established
	case clientReady:		return Error(fmt.Sprint("Invalid connection state."))

	case serverReady:		conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.state.Set(established)

							log.Info.Print("established new connection {", conn.Key(), "}.")

							return nil

	case clientEstablished: // The very first packet received through the traffic endpoint holds the MPTCP endpoint key
							// of destination (from the connection point of view) which we need later to be able to get
							// the network destination address for the sub flows.
							if key, ok, err := layer.GetMPTCPSenderKey(p.Payload); ok {
								if err == nil {
									if err := conn.routing.InsertFromMPTCPEndpointKey(key, conn.flow.NetFlow); err != nil {
										return Error(fmt.Sprint("Unable to insert MPTCP endpoint key. - ", err.Error()))
									}
								} else {
									return Error(fmt.Sprint("Error in fetching MPTCP endpoint key. - ", err.Error()))
								}
							}

							conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.state.Set(established)

							log.Info.Print("Established new connection {", conn.Key(), "}.")

							return nil

	case established: 		conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.state.Set(established)
							return nil

	case closed: 			// Packets sent through a closed connection are dropped.
							conn.state.Set(closed)
							return nil

	default: 				return Error(fmt.Sprint("Unknown connection state."))
	}
}

func (conn *Connection) processPacketFromKerepStateRaw(p *shila.Packet) error {

	// Assign the channels from the device through which the packet was received.
	var ep interface{} = p.Entrypoint
	if entryPoint, ok := ep.(*kernelEndpoint.Device); ok {
		conn.channels.KernelEndpoint.Ingress = entryPoint.TrafficChannels().Ingress // ingress from kernel end point
		conn.channels.KernelEndpoint.Egress  = entryPoint.TrafficChannels().Egress  // egress towards kernel end point
	} else {
		return Error(fmt.Sprint("Invalid entry point."))
	}

	// Create the packet network flow which is associated with the connection
	// If the packet contains a receiver token, then the new connection is a MPTCP sub flow.
	if token, ok, err := layer.GetMPTCPReceiverToken(p.Payload); ok {
		if err == nil {
			if netFlow, ok := conn.routing.RetrieveFromMPTCPEndpointToken(token); ok {
				conn.flow.NetFlow = netFlow
				conn.kind = 		sub
				p.Flow.NetFlow 	  = netFlow
			} else {
				return Error(fmt.Sprint("No network flow for MPTCP receiver token {", token, "}" +
					" beside having the packet containing it."))
			}
		} else {
			return Error(fmt.Sprint("Unable to fetch MPTCP receiver token. - ", err.Error()))
		}
	// For a MPTCP main flow the network flow can probably be extracted from the IP options
	} else if netFlow, ok, err := layer.GetNetFlowFromIPOptions(p.Payload); ok {
		if err == nil {
			conn.flow.NetFlow = netFlow
			conn.kind = main
			p.Flow.NetFlow 	  = netFlow
		} else {
			return Error(fmt.Sprint("Unable to get IP options. - ", err.Error()))
		}
	// For a MPTCP main flow the network flow is probably available in the routing table
	} else if netFlow, ok := conn.routing.RetrieveFromIPAddressPortKey(p.Flow.IPFlow.DstKey()); ok {
		conn.flow.NetFlow = netFlow
		conn.kind = main
		p.Flow.NetFlow 	  = netFlow
	// No valid option to get network flow :(
	} else {
		return Error(fmt.Sprint("No valid option to create network connection identifier."))
	}

	// Create the contacting connection
	contactingNetConnId, channels, err := conn.networkSide.EstablishNewContactingClientEndpoint(conn.flow)
	if err != nil {
		conn.state.Set(closed)
		panic("Implement custom error type.") // TODO!
		return nil
	}
	conn.channels.Contacting = channels

	// The contacting network connection id contains as src the local network endpoint which
	// was used by the network to established the contacting connection. Once a traffic connection
	// is established, we send this information to the corresponding server side.
	conn.flow.NetFlow.Src = contactingNetConnId.Src

	// Send the packet via the contacting channel
	conn.touched = time.Now()
	conn.channels.Contacting.Egress <- p

	// Try to connect to the address via path, a corresponding server should be there listening
	go func() {
		if trafficNetFlow, channels, err := conn.networkSide.EstablishNewTrafficClientEndpoint(conn.flow); err != nil {
			panic("Implement feedback to user!") // TODO!
			conn.Close()
		} else {
			conn.flow.NetFlow = trafficNetFlow
			conn.channels.NetworkEndpoint = channels
			conn.lock.Lock()
			conn.state.Set(clientEstablished)
			defer conn.lock.Unlock()
			// The contacting client endpoint is no longer needed.
			_ = conn.networkSide.TeardownContactingClientEndpoint(conn.flow.IPFlow)
		}
	}()

	// Set new state
	conn.state.Set(clientReady)

	return nil
}

func (conn *Connection) processPacketFromContactingEndpointStateRaw(p *shila.Packet) error {

	// Get the kernel endpoint from the kernel side manager
	dstKey := p.Flow.IPFlow.DstIPKey()
	if channels, ok := conn.kernelSide.GetTrafficChannels(dstKey); ok {
		conn.channels.KernelEndpoint = channels
	} else {
		conn.state.Set(closed)
		panic(fmt.Sprint("Connection {", conn.flow.IPFlow.Key(), "} can not process packet {", p.Flow.IPFlow.Key(),
		"}. - No kernel endpoint {", dstKey, "}.")) // TODO: Handle panic!
		return nil
	}

	// Send packet to kernel endpoint
	// --> 	Could still be that connection cannot be established, since we have no idea if there is actually a server listening
	conn.channels.KernelEndpoint.Egress <- p

	// If the packet is received through the contacting endpoint (server), then it's network connection id
	// is already set. This is the responsibility of the corresponding network server implementation.
	conn.flow.NetFlow = p.Flow.NetFlow

	// Request new incoming connection from network side.
	// ! The receiving network endpoint is responsible to correctly set the destination network address! !
	if channels, err := conn.networkSide.EstablishNewTrafficServerEndpoint(conn.flow); err != nil {
		conn.state.Set(closed)
		panic(fmt.Sprint("Connection {", conn.flow.IPFlow.Key(), "} can not process packet {", p.Flow.IPFlow.Key(),
		"}. - Unable to establish server endpoint. -", err.Error())) // TODO: Handle panic!
		return nil
	} else {
		conn.channels.NetworkEndpoint = channels
	}

	// Set new state
	conn.state.Set(serverReady)

	return nil
}
