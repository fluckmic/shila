package connection

import (
	"fmt"
	"shila/core/netflow"
	"shila/core/shila"
	"shila/kernelSide"
	"shila/kernelSide/kernelEndpoint"
	"shila/log"
	"shila/networkSide"
	"sync"
	"time"
)

type Connection struct {
	flow        shila.Flow
	state       state
	channels    channels
	lock        sync.Mutex
	touched     time.Time
	kernelSide  *kernelSide.Manager
	networkSide *networkSide.Manager
	router      *netflow.Router
}

type channels struct {
	KernelEndpoint  shila.PacketChannels // Kernel end point
	NetworkEndpoint shila.PacketChannels // Network end point
	Contacting      shila.PacketChannels // End point for connection establishment
}

func New(flow shila.Flow, kernelSide *kernelSide.Manager, networkSide *networkSide.Manager, router *netflow.Router) *Connection {
	return &Connection{
		flow:        flow,
		state:       newState(),
		lock:        sync.Mutex{},
		touched:     time.Now(),
		kernelSide:  kernelSide,
		networkSide: networkSide,
		router:      router,
	}
}

func (conn *Connection) Close() {

	conn.lock.Lock()
	defer conn.lock.Unlock()

	// Tear down all endpoints possibly associated with this connection
	_ = conn.networkSide.TeardownContactingClientEndpoint(conn.flow.IPFlow)
	_ = conn.networkSide.TeardownTrafficSeverEndpoint(conn.flow)
	_ = conn.networkSide.TeardownTrafficClientEndpoint(conn.flow.IPFlow)

	conn.state.set(closed)

	log.Info.Print("Closed connection {", conn.Key(), "}.")
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
		conn.setState(closed)
	}

	return err
}

func (conn *Connection) processPacketFromKerep(p *shila.Packet) error {
	switch conn.state.current {
	case raw:				return conn.processPacketFromKerepStateRaw(p)

	case clientReady:		p.Flow.NetFlow = conn.flow.NetFlow
							conn.touched = time.Now()
							conn.channels.Contacting.Egress <- p
							conn.setState(clientReady)
							return nil

	case serverReady: 		// Put packet into egress queue of connection. If the connection is established at one one point, these packets
							// are sent. If not they are lost. (--> Take care, could block if too many packets are in queue
							p.Flow.NetFlow = conn.flow.NetFlow
							conn.channels.NetworkEndpoint.Egress <- p
							conn.setState(serverReady)
							return nil

	case clientEstablished:	p.Flow.NetFlow = conn.flow.NetFlow
							conn.touched = time.Now()
							conn.channels.NetworkEndpoint.Egress <- p
							conn.setState(clientEstablished)
							return nil

	case established:		p.Flow.NetFlow = conn.flow.NetFlow
							conn.touched = time.Now()
							conn.channels.NetworkEndpoint.Egress <- p
							conn.setState(established)
							return nil

	case closed: 			conn.setState(closed)
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
							conn.setState(serverReady)
							return nil

	case established: 		conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.setState(established)
							return nil

	case closed: 			// Packets sent through a closed connection are dropped.
							conn.setState(closed)
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
							conn.setState(established)

							log.Info.Print("Established new connection {", conn.Key(), "}.")

							return nil

	case clientEstablished: // The very first packet received through the traffic endpoint holds the MPTCP endpoint key
							// of destination (from the connection point of view) which we need later to be able to get
							// the network destination address for the subflow.
							if err := conn.router.InsertFromSynAckMpCapable(p, conn.flow.NetFlow); err != nil {
								return Error(fmt.Sprint("Unable to update router. - ", err.Error()))
							}

							conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.setState(established)

							log.Info.Print("Established new connection {", conn.Key(), "}.")

							return nil

	case established: 		conn.touched 	= time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.setState(established)
							return nil

	case closed: 			// Packets sent through a closed connection are dropped.
							conn.setState(closed)
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

	// Get the network flow
	var err error
	conn.flow.NetFlow, conn.flow.Kind, err = conn.router.Route(p)
	if err != nil {
		return Error(fmt.Sprint("Unable to get network flow."))
	}
	p.Flow.NetFlow = conn.flow.NetFlow

	// Create the contacting connection
	contactingNetFlow, channels, err := conn.networkSide.EstablishNewContactingClientEndpoint(conn.flow)
	if err != nil {
		conn.setState(closed)
		panic("Implement custom error type or feedback to user") // TODO!
		return nil
	}
	conn.channels.Contacting = channels

	// The contacting network connection id contains as src the local network endpoint which
	// was used by the network to established the contacting connection. Once a traffic connection
	// is established, we send this information to the corresponding server side.
	conn.flow.NetFlow.Src = contactingNetFlow.Src

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
			conn.setState(clientEstablished)
			defer conn.lock.Unlock()
			// The contacting client endpoint is no longer needed.
			_ = conn.networkSide.TeardownContactingClientEndpoint(conn.flow.IPFlow)
		}
	}()

	// set new state
	conn.state.set(clientReady)

	return nil
}

func (conn *Connection) processPacketFromContactingEndpointStateRaw(p *shila.Packet) error {

	// Not the kernel endpoint from the kernel side manager
	dstKey := p.Flow.IPFlow.DstIPKey()
	if channels, ok := conn.kernelSide.GetTrafficChannels(dstKey); ok {
		conn.channels.KernelEndpoint = channels
	} else {
		conn.state.set(closed)
		return Error(fmt.Sprint("No kernel endpoint {", dstKey, "}."))
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
		conn.state.set(closed)
		return Error(fmt.Sprint("Unable to establish server endpoint. - ", err.Error()))
	} else {
		conn.channels.NetworkEndpoint = channels
	}

	// set new state
	conn.state.set(serverReady)

	return nil
}

func (conn *Connection) setState(state stateIdentifier) {
	conn.state.set(state)
	if conn.state.previous != conn.state.current {
		log.Verbose.Print("Connection {", conn.Key(), "} changed state from {", conn.state.previous,
			"} to {", conn.state.current, "}.")
	}
}