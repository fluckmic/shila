package networkEndpoint

import (
	"fmt"
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
	"shila/core/shila"
	"shila/log"
	"sync"
	"time"
)

var _ shila.NetworkServerEndpoint = (*Server)(nil)

type Server struct{
	NetworkEndpointBase
	backboneConnections BackboneConnections
	lAddress            shila.NetworkAddress
	lConnection			*snet.Conn
	lock                sync.Mutex
	holdingArea 		[] *shila.Packet
}

func NewServer(lAddr shila.NetworkAddress, role shila.EndpointRole, issues shila.EndpointIssuePubChannel) shila.NetworkServerEndpoint {
	return &Server{
		NetworkEndpointBase: 	NetworkEndpointBase{
									role:    role,
									ingress: make(shila.PacketChannel, Config.SizeIngressBuffer),
									egress:  make(shila.PacketChannel, Config.SizeEgressBuffer),
									state:   shila.NewEntityState(),
									issues:  issues,
								},
		lAddress:            	lAddr,
		holdingArea:			make([] *shila.Packet, 0, Config.SizeHoldingArea),
		lock:                	sync.Mutex{},
	}
}

func (server *Server) SetupAndRun() (err error) {

	if server.state.Not(shila.Uninitialized) {
		return shila.CriticalError(server.Says(fmt.Sprint("In wrong state {", server.state, "}.")))
	}

	// Setup the backbone connections
	server.backboneConnections = NewBackboneConnections(server)

	// Connection to listen for incoming backbone connections.
	server.lConnection, err = appnet.Listen(server.lAddress.(*snet.UDPAddr).Host)
	if err != nil {
		return shila.PrependError(shila.NetworkConnectionError(err.Error()), "Unable to setup listener.")
	}

	go server.serveIngress() 			// Start listening for incoming backbone connections.
	go server.serveEgress()  			// Start handling incoming packets.
	go server.resendFunctionality() 	// Start the resending functionality

	server.state.Set(shila.Running)
	log.Verbose.Print(server.Says("Setup and running."))
	return
}

func (server *Server) TearDown() (err error) {

	server.state.Set(shila.TornDown)

	err = server.lConnection.Close()            // Close the listening connection. (Server no longer receives incoming connections.)
	err = server.backboneConnections.TearDown() // Properly terminate all existing backbone connections.

	close(server.ingress) 						// Close the ingress channel (Working side no longer processes this endpoint)

	log.Verbose.Print(server.Says("Got torn down."))
	return err
}

func (server *Server) TrafficChannels() shila.PacketChannels {
	return shila.PacketChannels{Ingress: server.ingress, Egress: server.egress}
}

func (server *Server) Role() shila.EndpointRole {
	return server.role
}

func (server *Server) Identifier() string {
	return fmt.Sprint("{", server.Role(), " Server - ", server.lAddress, " <- *}")
}

func (server *Server) Says(str string) string {
	return  fmt.Sprint(server.Identifier(), ": ", str)
}

func (server *Server) serveIngress(){
	buffer := make([]byte, Config.SizeRawIngressStorage)
	for {
		n, from, err := server.lConnection.ReadFrom(buffer)
		if err != nil {
			if server.state.Not(shila.Running) {
				// TODO: Handle error. Definitely a problem to be considered by the server.
				panic(fmt.Sprint("Handle error: ", err.Error()))
			}
			return
		}

		// Does not return any error. Problems in the pipeline are handled internally.
		// In the worst case the input data is just dropped.
		server.backboneConnections.WriteIngress(from, buffer[:n])
	}
}

func (server *Server) serveEgress() {
	for p := range server.egress {
		err := server.backboneConnections.WriteEgress(p)
		if err != nil {
			if server.state.Not(shila.Running) {
				// TODO: Handle error. Definitely a problem to be considered by the server.
				panic(fmt.Sprint("Handle error: ", err.Error()))
			}
			return
		}
	}
}

func (server *Server) resendFunctionality() {
	for {
		time.Sleep(Config.ServerResendInterval)
		server.lock.Lock()
		for _, p := range server.holdingArea {
			if p.TTL > 0 {
				p.TTL--
				server.egress <- p
			} else {
				// Server network endpoint is "not able" to send out the given packet.
				err := shila.NetworkEndpointTimeout("Unable to send packet.")
				server.issues <- shila.EndpointIssuePub { Issuer: server, Flow: p.Flow, Error: err }
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