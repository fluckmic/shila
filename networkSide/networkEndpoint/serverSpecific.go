package networkEndpoint

import (
	"fmt"
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
	"shila/core/shila"
	"shila/log"
	"sync"
)

var _ shila.NetworkServerEndpoint = (*Server)(nil)

type Server struct{
	NetworkEndpointBase
	backboneConnections BackboneConnections
	lAddress            shila.NetworkAddress
	lConnection			*snet.Conn
	lock                sync.Mutex
	holdingArea         [] *shila.Packet
}

func NewServer(lAddr shila.NetworkAddress, role shila.EndpointRole, issues shila.EndpointIssuePubChannel) shila.NetworkServerEndpoint {
	return &Server{
		NetworkEndpointBase: 	NetworkEndpointBase{
									role:    role,
									ingress: make(chan *shila.Packet, Config.SizeIngressBuffer),
									egress:  make(chan *shila.Packet, Config.SizeEgressBuffer),
									state:   shila.NewEntityState(),
									issues:  issues,
								},
		lAddress:            	lAddr,
		lock:                	sync.Mutex{},
		holdingArea:         	make([] *shila.Packet, 0, Config.SizeHoldingArea),
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

	go server.serveIngress() // Start listening for incoming backbone connections.
	go server.serveEgress()  // Start handling incoming packets.

	server.state.Set(shila.Running)
	log.Verbose.Print(server.Says("Setup and running."))
	return
}

func (server *Server) TearDown() (err error) {

	server.state.Set(shila.TornDown)

	err = server.lConnection.Close()            // Close the listening connection. (Server no longer receives incoming connections.)
	err = server.backboneConnections.TearDown() // Properly terminate all existing backbone connections.

	close(server.ingress) 							// Close the ingress channel (Working side no longer processes this endpoint)

	log.Verbose.Print(server.Says("Got torn down."))
	return err
}

func (server *Server) TrafficChannels() shila.PacketChannels {
	return shila.PacketChannels{Ingress: server.ingress, Egress: server.egress}
}

func (server *Server) Role() shila.EndpointRole {
	return server.role
}

func (server *Server) Identifier() shila.EndpointIdentifier {
	return shila.EndpointIdentifier(shila.GetNetworkAddressKey(server.lAddress))
}

func (server *Server) Says(str string) string {
	return fmt.Sprint("Server {", server.Role(), " - ", server.lAddress, " <- *}: ", str)
}

func (server *Server) serveIngress(){
	buffer := make([]byte, Config.SizeRawIngressStorage)
	for {
		n, from, err := server.lConnection.ReadFrom(buffer)
		if err != nil {
			if server.state.Not(shila.Running) {
				panic("Handle me.") // TODO.
			}
			return
		}
		server.backboneConnections.WriteIngress(from, buffer[:n])
	}
}

func (server *Server) serveEgress() {
	for p := range server.egress {
		to := p.Flow.NetFlow.Dst
		server.backboneConnections.WriteEgress(to, p.Payload)
	}
}