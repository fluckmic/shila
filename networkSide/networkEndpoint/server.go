package networkEndpoint

import (
	"fmt"
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
	"shila/config"
	"shila/core/shila"
	"shila/log"
	"shila/measurements"
	"sync"
	"time"
)

var _ shila.NetworkServerEndpoint = (*Server)(nil)

type Server struct{
	Base
	key					shila.NetworkAddressKey
	backboneConnections ServerBackboneConnections
	lAddress            shila.NetworkAddress
	lConnection         *snet.Conn
	lock                sync.Mutex
	holdingArea         [] *shila.Packet
}

func NewServer(lAddr shila.NetworkAddress, role shila.EndpointRole, issues shila.EndpointIssuePubChannel) shila.NetworkServerEndpoint {
	return &Server{
		Base: 			Base{
									Role:    role,
									Ingress: make(shila.PacketChannel, config.Config.NetworkEndpoint.SizeIngressBuffer),
									Egress:  make(shila.PacketChannel, config.Config.NetworkEndpoint.SizeEgressBuffer),
									State:   shila.NewEntityState(),
									Issues:  issues,
								},
		key:			shila.GetNetworkAddressKey(lAddr),
		lAddress:       lAddr,
		holdingArea:	make([] *shila.Packet, 0, config.Config.NetworkEndpoint.SizeHoldingArea),
		lock:           sync.Mutex{},
	}
}

func (server *Server) SetupAndRun() (err error) {

	if server.State.Not(shila.Uninitialized) {
		return shila.CriticalError(server.Says(fmt.Sprint("In wrong State ", server.State, ".")))
	}

	// Setup the backbone connections
	server.backboneConnections = NewBackboneConnections(server)

	// Connection to listen for incoming backbone connections.
	server.lConnection, err = appnet.Listen(server.lAddress.(*snet.UDPAddr).Host)
	if err != nil {
		return shila.PrependError(ConnectionError(err.Error()), "Unable to setup listener.")
	}

	go server.serveIngress() 			// Start listening for incoming backbone connections.
	go server.serveEgress()  			// Start handling incoming packets.
	go server.resendFunctionality() 	// Start the resending functionality

	server.State.Set(shila.Running)
	log.Verbose.Print(server.Says("Setup and running."))
	return
}

func (server *Server) TearDown() (err error) {

	server.State.Set(shila.TornDown)

	err = server.lConnection.Close()            // Close the listening connection. (Server no longer receives incoming connections.)
	err = server.backboneConnections.TearDown() // Properly terminate all existing backbone connections.

	close(server.Ingress) // Close the Ingress channel (Working side no longer processes this endpoint)

	log.Verbose.Print(server.Says("Got torn down."))
	return err
}

func (server *Server) TrafficChannels() shila.PacketChannels {
	return shila.PacketChannels{Ingress: server.Ingress, Egress: server.Egress}
}

func (server *Server) Role() shila.EndpointRole {
	return server.Base.Role
}

func (server *Server) Identifier() string {
	return fmt.Sprint("Server ", server.Role(), " (", server.lAddress, " <- *)")
}

func (server *Server) Says(str string) string {
	return  fmt.Sprint(server.Identifier(), ": ", str)
}

func (server *Server) Key() shila.NetworkAddressKey {
	return server.key
}

func (server *Server) serveIngress(){


	buffer := make([]byte, config.Config.NetworkEndpoint.SizeRawIngressStorage)
	for {
		n, from, err := server.lConnection.ReadFrom(buffer)
		if err != nil {
			go server.handleConnectionIssue(err)
			return
		}
		go func() {
			// ...probably create a timestamp for it..
			if config.Config.Logging.DoIngressTimestamping {
				measurements.LogIngressTimestamp(buffer[:n])
			}
		}()

		// Does not return any error. Problems in the pipeline are handled internally.
		// In the worst case the input data is just dropped.
		server.backboneConnections.WriteIngress(from, buffer[:n])

		// Server upon reception of data:
		// 1. Determine scion address of sender
		// 2. Check if there is an existing backbone connection?
		// 		If not, create backbone connection:
		//		2.1 Fetch control message and finalize setup of backbone connection
		//		If contact server endpoint:
		// 			Set receiving in port to the one it will be for the traffic server endpoint.
		//			This info is required to find the right server in egress processing.
		//		If traffic server endpoint:
		//			Make backbone this backbone connection also findable for traffic coming
		//			from the client contacting endpoint.
		//			This info is required to find the right backbone connection.
		// 3. Hand data to corresponding backbone connection
		// 4.
	}
}

func (server *Server) serveEgress() {
	for p := range server.Egress {
		err := server.backboneConnections.WriteEgress(p)
		if err != nil {
			go server.handleConnectionIssue(err)
			return
		}
	}
}

func (server *Server) resendFunctionality() {
	for {
		time.Sleep(time.Duration(config.Config.NetworkEndpoint.ServerResendInterval) * time.Second)
		server.lock.Lock()
		for _, p := range server.holdingArea {
			if p.TTL > 0 {
				p.TTL--
				server.Egress <- p
			} else {
				// Server network endpoint is "not able" to send out the given packet.
				err := ConnectionError("Unable to send packet.")
				server.Issues <- shila.EndpointIssuePub { Issuer: server, Key: p.Flow.TCPFlow.Key(), Error: err } // TODO!
			}
		}
		server.holdingArea = server.holdingArea[:0]
		server.lock.Unlock()
	}
}

func (server *Server) addToHoldingArea(packet *shila.Packet) {
	server.lock.Lock()
	defer server.lock.Unlock()
	server.holdingArea = append(server.holdingArea, packet)
}

func (server *Server) handleConnectionIssue(err error) {
	// Wait a little bit - maybe the server is going to die anyway.
	time.Sleep(time.Duration(config.Config.NetworkEndpoint.WaitingTimeAfterConnectionIssue) * time.Second)
	if server.State.Is(shila.Running) {
		log.Error.Println(server.Says(fmt.Sprint("Publishes issue - ", err.Error())))
		server.Issues <- shila.EndpointIssuePub{Issuer: server, Error: ConnectionError(err.Error())}
	}
}