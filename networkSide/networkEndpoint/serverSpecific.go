package networkEndpoint

import (
	"encoding/gob"
	"fmt"
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
	"io"
	"net"
	"shila/core/shila"
	"shila/layer/tcpip"
	"shila/log"
	"shila/networkSide/network"
	"strconv"
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

type MappingBackboneConnections map[shila.NetworkAddressKey]  *networkConnection

func NewServer(lAddr shila.NetworkAddress, role shila.EndpointRole, issues shila.EndpointIssuePubChannel) shila.NetworkServerEndpoint {
	return &Server{
		NetworkEndpointBase: 	NetworkEndpointBase{
									role:    role,
									ingress: make(chan *shila.Packet, Config.SizeIngressBuffer),
									egress:  make(chan *shila.Packet, Config.SizeEgressBuffer),
									state:   shila.NewEntityState(),
									issues:  issues,
								},
		backboneConnections: 	NewBackboneConnections(),
		lAddress:            	lAddr,
		lock:                	sync.Mutex{},
		holdingArea:         	make([] *shila.Packet, 0, Config.SizeHoldingArea),
	}
}

func (server *Server) SetupAndRun() (err error) {

	if server.state.Not(shila.Uninitialized) {
		return shila.CriticalError(server.msg(fmt.Sprint("In wrong state {", server.state, "}.")))
	}

	// Connection to listen for incoming backbone connections.
	server.lConnection, err = appnet.Listen(server.lAddress.(*snet.UDPAddr).Host)
	if err != nil {
		return shila.PrependError(shila.NetworkConnectionError(err.Error()), "Unable to setup listener.")
	}

	go server.serveIncomingConnections() 		// Start listening for incoming backbone connections.
	go server.serveEgress()                  	// Start handling incoming packets.
	go server.resendFunctionality()          	// Start the resendFunctionality functionality.

	server.state.Set(shila.Running)
	log.Verbose.Print(server.msg("Setup and running."))
	return
}

func (server *Server) TearDown() (err error) {

	server.state.Set(shila.TornDown)

	err = server.lConnection.Close() 				// Close the listening connection. (Server no longer receives incoming connections.)
	err = server.backboneConnections.tearDown()		// Properly terminate all existing backbone connections.

	close(server.ingress) 							// Close the ingress channel (Working side no longer processes this endpoint)

	log.Verbose.Print(server.msg("Got torn down."))
	return err
}

func (server *Server) TrafficChannels() shila.PacketChannels {
	return shila.PacketChannels{Ingress: server.ingress, Egress: server.egress}
}

func (server *Server) Role() shila.EndpointRole {
	return server.role
}

func (server *Server) Key() shila.EndpointKey {
	return shila.EndpointKey(shila.GetNetworkAddressKey(server.lAddress))
}

func (server *Server) serveIncomingConnections(){
	buffer := make([]byte, Config.SizeRawIngressStorage)
	for {
		n, from, err := server.lConnection.ReadFrom(buffer)
		if err != nil {
			if server.state.Not(shila.Running) {
				panic("Handle me.") // TODO.
			}
			return
		}
		srcKey := shila.GetNetworkAddressKey(from)
		server.backboneConnections.retrieve(srcKey).writeIngress(buffer[:n])
	}
}

func (server *Server) handleBackboneConnection(backConn *net.TCPConn) {

	// Create the true net lAddress (Server: src <- dst)
	path, _ := network.PathGenerator{}.New("")
	trueNetFlow := shila.NetFlow{
		Src:  backConn.LocalAddr().(*net.TCPAddr),
		Path: path,
		Dst:  backConn.RemoteAddr().(*net.TCPAddr),
	}

	log.Verbose.Print(server.msgFlowRelated(trueNetFlow, "Accepted backbone connection."))

	// Fetch the control message
	var ctrlMsg controlMessage
	if err := gob.NewDecoder(io.Reader(backConn)).Decode(&ctrlMsg); err != nil {
		server.closeBackboneConnectionWithErrorMsg(backConn, trueNetFlow, err, "Cannot fetch control message.")
		return
	}

	// Create the representing lAddress (Content of control message is correct relative to sender)
	representingFlow := shila.Flow{ IPFlow: ctrlMsg.IPFlow.Swap(), NetFlow: trueNetFlow }

	// If endpoint is a contacting endpoint, then the representing lAddress is different from the true net lAddress:
	if server.Role() == shila.ContactingNetworkEndpoint {
		srcAddrTrafficEndpoint, err := server.calculateSrcAddrOfTrafficServerNetworkEndpoint(representingFlow)
		if err != nil {
			server.closeBackboneConnectionWithErrorMsg(backConn, trueNetFlow, err, "Cannot generate src address of traffic server network endpoint.")
			return
		}
		representingFlow.NetFlow.Src = srcAddrTrafficEndpoint
	}

	// Generate the keys
	var keys []shila.NetworkAddressKey
	keys = append(keys, shila.GetNetworkAddressKey(representingFlow.NetFlow.Dst))
	// We need also to be able to send messages to the contact client network endpoint.
	if server.Role() == shila.TrafficNetworkEndpoint {
		// For the moment we can use the same path for this key as for the representing lAddress.
		keys = append(keys, shila.GetNetworkAddressKey(&ctrlMsg.SrcAddrContactEndpoint))
	}

	// Create the connection wrapper
	connection := networkConnection{
		EndpointRole:     server.role,
		TrueNetFlow:      trueNetFlow,
		RepresentingFlow: representingFlow,
		Backbone:         backConn,
	}

	// Add the new backbone connection to the mapping, so that it can be found by the egress handler.
	if err := server.insertBackboneConnection(keys, &connection); err != nil {
		server.closeBackboneConnectionWithErrorMsg(backConn, trueNetFlow, err, "Cannot insert backbone conn.")
		return
	}

	// Start the ingress handler for the backbone connection.
	server.serveIngress(&connection)

	// No longer necessary or possible to serve the ingress, remove the backbone connection from the mapping.
	/*
	server.lock.Lock()
	for _, key := range keys {
		delete(server.backboneConnections, key)
	}
	server.lock.Unlock()
	 */
	return
}

func (server * Server) insertBackboneConnection(keys []shila.NetworkAddressKey, conn *networkConnection) error {

	/*
	server.lock.Lock()
	defer server.lock.Unlock()

	for _, key := range keys {
		if _, ok := server.backboneConnections[key]; ok {
			return shila.TolerableError(fmt.Sprint("Duplicate key {", key, "}."))
		}
	}
	for _, key := range keys {
		server.backboneConnections[key] = conn
	}
	return nil
	*/
	return nil
}

func (server * Server) calculateSrcAddrOfTrafficServerNetworkEndpoint(representingFlow shila.Flow) (shila.NetworkAddress, error) {
	// It is the responsibility of the contact server endpoint to calculate the
	// source address of the corresponding traffic server endpoint.
	ip   := representingFlow.NetFlow.Src.(*net.TCPAddr).IP.String()
	port := strconv.Itoa(representingFlow.IPFlow.Src.Port)
	return network.AddressGenerator{}.New(net.JoinHostPort(ip, port))
}

func (server * Server) closeBackboneConnectionWithErrorMsg(conn *net.TCPConn, flow shila.NetFlow, err error, msg string) {
	log.Error.Print(server.msgFlowRelated(flow, shila.PrependError(err, msg).Error()))
	conn.Close()
	log.Error.Print(server.msgFlowRelated(flow, "Closed backbone connection."))
}

func (server *Server) serveIngress(connection *networkConnection) {

	ingressRaw := make(chan byte, Config.SizeRawIngressBuffer)
	go server.packetize(connection.RepresentingFlow, ingressRaw)

	reader := io.Reader(connection.Backbone)
	storage := make([]byte, Config.SizeRawIngressStorage)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, Config.ReadSizeRawIngress)
		// If the incoming connection suffers from an error, we close it and return. The server instance is still able
		// to receive BackboneConnections as long as it is not shut down by the manager of the network side.
		if err != nil {
			close(ingressRaw) // Stop the packetizing.
			return
		}
		for _, b := range storage[:nBytesRead] {
			ingressRaw <- b
		}
	}
}

func (server *Server) packetize(flow shila.Flow, ingressRaw chan byte) {
	for {
		if rawData, err := tcpip.PacketizeRawData(ingressRaw, Config.SizeRawIngressStorage); rawData != nil {
				server.ingress <- shila.NewPacketWithNetFlow(server, flow.IPFlow.Swap(), flow.NetFlow.Swap(), rawData)
		} else {
			if err == nil {
				// All good, ingress raw closed.
				return
			}
			err := shila.PrependError(shila.ParsingError(err.Error()), "Issue in raw data packetizer.")
			server.issues <- shila.EndpointIssuePub{	Issuer: server, Flow: flow, Error: err }
			return
		}
	}
}

func (server *Server) resendFunctionality() {
	/*
	for {
		time.Sleep(Config.ServerResendInterval)
		server.lock.Lock()
		for _, p := range server.holdingArea {
			if p.TTL > 0 {
				p.TTL--
				server.egress <- p
			} else {
				// Server network endpoint is not able to send out the given packet.
				err := shila.NetworkEndpointTimeout("Unable to send packet.")
				server.issues <- shila.EndpointIssuePub { Issuer: server,	Flow: p.Flow, Error: err }
			}
		}
		server.holdingArea = server.holdingArea[:0]
		server.lock.Unlock()
	}
	*/
}

func (server *Server) serveEgress() {
	for p := range server.egress {
		dstKey := p.Flow.NetFlow.DstKey() // Retrieve key to get the correct backbone connection
		server.backboneConnections.retrieve(dstKey).writeEgress(p.Payload)
	}
}

func (server *Server) msg(str string) string {
	return fmt.Sprint("Server {", server.Role(), " - ", server.lAddress, " <- *}: ", str)
}

func (server *Server) msgFlowRelated(flow shila.NetFlow, str string) string {
	return fmt.Sprint("Server {", server.Role(), " - ", flow.Src.String(), " <- ", flow.Dst.String(),"}: ", str)
}
