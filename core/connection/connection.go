//
package connection

import (
	"fmt"
	"github.com/bclicn/color"
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
	key			shila.IPFlowKey
	flow        shila.Flow
	kind        shila.FlowType
	state       state
	channels    channels
	lock        sync.Mutex
	touched     time.Time
	kernelSide  *kernelSide.Manager
	networkSide *networkSide.Manager
	router      netflow.Router
}

type channels struct {
	KernelEndpoint  shila.PacketChannels // Kernel end point
	NetworkEndpoint shila.PacketChannels // Network end point
	Contacting      shila.PacketChannels // End point for connection establishment
}

func New(flow shila.Flow, kernelSide *kernelSide.Manager, networkSide *networkSide.Manager, router netflow.Router) *Connection {
	return &Connection{
		key:         flow.IPFlow.Key(),
		flow:        flow,
		state:       newState(),
		lock:        sync.Mutex{},
		touched:     time.Now(),
		kernelSide:  kernelSide,
		networkSide: networkSide,
		router:      router,
	}
}

func (conn *Connection) Close(err error) {

	conn.lock.Lock()
	defer conn.lock.Unlock()

	// Dont need to close a connection multiple times
	if conn.state.current == closed {
		return
	}

	// Tear down all endpoints possibly associated with this connection
	_ = conn.networkSide.TeardownContactingClientEndpoint(conn.flow.IPFlow)
	_ = conn.networkSide.TeardownTrafficSeverEndpoint(conn.flow)
	_ = conn.networkSide.TeardownTrafficClientEndpoint(conn.flow.IPFlow)

	conn.state.set(closed)

	log.Info.Print(conn.Says(shila.PrependError(err, "Closed.").Error()))
}

func (conn *Connection) ProcessPacket(p *shila.Packet) error {

	conn.lock.Lock()
	defer conn.lock.Unlock()

	// From where was the packet received?
	var err error
	switch p.Entrypoint.Role() {
		case shila.IngressKernelEndpoint:		err = conn.processPacketFromKerep(p)
		case shila.EgressKernelEndpoint:		err = conn.processPacketFromKerep(p)
		case shila.ContactNetworkEndpoint: 		err = conn.processPacketFromContactingEndpoint(p)
		case shila.TrafficNetworkEndpoint:		err = conn.processPacketFromTrafficEndpoint(p)
		default:
			err = shila.CriticalError(fmt.Sprint("Unknown entry point label ", p.Entrypoint.Role(), "."))
	}

	if err != nil {
		conn.Close(err)
	}

	return err
}

func (conn *Connection) processPacketFromKerep(p *shila.Packet) error {
	switch conn.state.current {
	case raw:				return conn.processPacketFromKerepStateRaw(p)

	case clientReady:		p.Flow.NetFlow = conn.flow.NetFlow
							// conn.touched = time.Now()
							conn.channels.Contacting.Egress <- p
							return nil

	case clientEstablished:	p.Flow.NetFlow = conn.flow.NetFlow
							// conn.touched = time.Now()
							conn.channels.NetworkEndpoint.Egress <- p
							return nil

	case serverReady: 		// Put packet into egress queue of connection. If the connection is established at one one point, these packets
							// are sent. If not they are lost. (--> Take care, could block if too many packets are in queue
							p.Flow.NetFlow = conn.flow.NetFlow
							conn.channels.NetworkEndpoint.Egress <- p
							conn.setState(serverEstablished)
							return nil

	case serverEstablished: p.Flow.NetFlow = conn.flow.NetFlow
							conn.channels.NetworkEndpoint.Egress <- p
							return nil

	case established:		p.Flow.NetFlow = conn.flow.NetFlow
							conn.touched = time.Now()
							conn.channels.NetworkEndpoint.Egress <- p
							return nil

	case closed: 			return nil

	default: 				return shila.CriticalError("Unknown connection state.")

	}
}

func (conn *Connection) processPacketFromContactingEndpoint(p *shila.Packet) error {
	switch conn.state.current {

	case raw:				return conn.processPacketFromContactingEndpointStateRaw(p)

	case clientReady:		return shila.CriticalError("Both endpoints in client state.") // TODO: TO THINK.

	case clientEstablished: return shila.CriticalError(fmt.Sprint("Invalid connection state ", conn.state.current, "."))

	case serverReady:		conn.channels.KernelEndpoint.Egress <- p
							return nil

	case serverEstablished: conn.channels.KernelEndpoint.Egress <- p
							return nil

	case established: 		conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							return nil

	case closed: 		 	return nil

	default: 				return shila.CriticalError("Unknown connection state.")

	}
}

func (conn *Connection) processPacketFromTrafficEndpoint(p *shila.Packet) error {
	switch conn.state.current {

	case raw:				return shila.CriticalError(fmt.Sprint("Invalid connection state ", conn.state.current, "."))

							// A packet from the traffic endpoint is just received if network connection is established
	case clientReady:		return shila.CriticalError(fmt.Sprint("Invalid connection state ", conn.state.current, "."))

	case clientEstablished: // The very first packet received through the traffic endpoint holds the MPTCP endpoint key
							// of destination (from the connection point of view) which we need later to be able to get
							// the network destination address for the subflow.
							if err := conn.router.InsertFromSynAckMpCapable(p, conn.flow.NetFlow); err != nil {
								return shila.TolerableError(fmt.Sprint("Unable to update router.", err.Error()))
							}

							conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.setState(established)

							log.Info.Print(conn.Says(color.Green("Successfully established!")))

							return nil

	case serverReady:		// Packets sent to the traffic endpoint before the connection is established are ignored.
							return nil

	case serverEstablished: conn.touched = time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							conn.setState(established)

							log.Info.Print(conn.Says(color.Green("Successfully established!")))
							return nil

	case established: 		conn.touched 	= time.Now()
							conn.channels.KernelEndpoint.Egress <- p
							return nil

	case closed: 			return nil

	default: 				return shila.CriticalError("Unknown connection state.")
	}
}

func (conn *Connection) processPacketFromKerepStateRaw(p *shila.Packet) error {

	// Assign the channels from the device through which the packet was received.
	var ep interface{} = p.Entrypoint
	if entryPoint, ok := ep.(*kernelEndpoint.Device); ok {
		conn.channels.KernelEndpoint.Ingress = entryPoint.TrafficChannels().Ingress // ingress from kernel end point
		conn.channels.KernelEndpoint.Egress  = entryPoint.TrafficChannels().Egress  // egress towards kernel end point
	} else {
		return shila.CriticalError("Invalid entry point.")
	}

	// Get the network flow
	var err error
	conn.flow.NetFlow, conn.kind, err = conn.router.Route(p)
	if err != nil {
		return shila.TolerableError(fmt.Sprint("Unable to get network flow.", err.Error()))
	}
	p.Flow.NetFlow = conn.flow.NetFlow

	// Create the contacting connection
	contactingNetFlow, channels, err := conn.networkSide.EstablishNewContactingClientEndpoint(conn.flow)
	if err != nil {
		return shila.TolerableError(fmt.Sprint( "Unable to establish contacting connection.", err.Error()))
	}
	conn.channels.Contacting = channels

	// The contacting network flow contains as src the local network endpoint which
	// was used by the network to established the contacting connection. Once a traffic connection
	// is established, we send this information to the corresponding server side.
	conn.flow.NetFlow.Src = contactingNetFlow.Src

	// Send the packet via the contacting channel
	conn.touched = time.Now()
	conn.channels.Contacting.Egress <- p

	// Try to connect to the address via path, a corresponding server should be there listening
	go func() {
		// Wait a certain amount of time to give the server endpoint time to establish itself
		time.Sleep(Config.WaitingTimeTrafficConnEstablishment)
		if trafficNetFlow, channels, err := conn.networkSide.EstablishNewTrafficClientEndpoint(conn.flow); err != nil {
			conn.Close(err)
		} else {
			conn.lock.Lock()
			conn.flow.NetFlow = trafficNetFlow
			conn.channels.NetworkEndpoint = channels
			conn.setState(clientEstablished)
			// The contacting client endpoint is no longer needed.
			_ = conn.networkSide.TeardownContactingClientEndpoint(conn.flow.IPFlow)
			conn.lock.Unlock()
		}
	}()

	// set new state
	conn.state.set(clientReady)

	return nil
}

func (conn *Connection) processPacketFromContactingEndpointStateRaw(p *shila.Packet) error {

	// Get the kernel endpoint from the kernel side manager
	packetDstKey := p.Flow.IPFlow.DstIPKey()
	if channels, ok := conn.kernelSide.GetTrafficChannels(packetDstKey); ok {
		conn.channels.KernelEndpoint = channels
	} else {
		conn.state.set(closed)
		return shila.TolerableError(fmt.Sprint("Cant process packet. No kernel endpoint for ", packetDstKey, ".")) // TODO: TO THINK.
	}

	// Send packet to kernel endpoint
	// --> 	Could still be that connection cannot be established, since we have no idea if there is actually a server listening
	conn.channels.KernelEndpoint.Egress <- p

	// If the packet is received through the contacting endpoint (server), then it's network connection id
	// is already set. This is the responsibility of the corresponding network server implementation.
	conn.flow.NetFlow = p.Flow.NetFlow.Swap()

	// Request new incoming connection from network side.
	// ! The receiving network endpoint is responsible to correctly set the destination network address! !
	if channels, err := conn.networkSide.EstablishNewTrafficServerEndpoint(conn.flow.NetFlow.Src, conn.key); err != nil {
		conn.state.set(closed)
		return shila.TolerableError(fmt.Sprint("Unable to establish server endpoint.", err.Error()))
	} else {
		conn.channels.NetworkEndpoint = channels
	}

	// set new state
	conn.setState(serverReady)

	return nil
}

func (conn *Connection) setState(state stateIdentifier) {
	conn.state.set(state)
	if conn.state.previous != conn.state.current {
		log.Verbose.Println(conn.Says(fmt.Sprint("State change from ", conn.state.previous, " to ", conn.state.current, ".")))
	}
}

func (conn *Connection) Identifier() string {
	return fmt.Sprint(conn.kind, " Connection (", conn.flow.NetFlow.Src, " <-> ", conn.flow.NetFlow.Dst, ")")
}

func (conn *Connection) Says(str string) string {
	return  fmt.Sprint(conn.Identifier(), ": ", str)
}