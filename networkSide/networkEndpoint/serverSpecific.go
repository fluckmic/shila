package networkEndpoint

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"shila/core/shila"
	"shila/layer/tcpip"
	"shila/log"
	"shila/networkSide/network"
	"strconv"
	"sync"
	"time"
)

var _ shila.NetworkServerEndpoint = (*Server)(nil)

type Server struct{
	Base
	backboneConnections map[shila.NetworkAddressAndPathKey]  networkConnection
	flow	            shila.Flow
	listener            net.Listener
	lock                sync.Mutex
	holdingArea         []*shila.Packet
}

func NewServer(flow shila.Flow, label shila.EndpointLabel, endpointIssues shila.EndpointIssuePubChannel) shila.NetworkServerEndpoint {
	return &Server{
		Base: 			Base{
								label: 			label,
								state: 			shila.NewEntityState(),
								endpointIssues: endpointIssues,
						},
		backboneConnections: make(map[shila.NetworkAddressAndPathKey]  networkConnection),
		flow:             	 flow,
		lock:                sync.Mutex{},
		holdingArea:         make([]*shila.Packet, 0, Config.SizeHoldingArea),
	}
}

func (s *Server) SetupAndRun() error {

	if s.state.Not(shila.Uninitialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", s.state, "}."))
	}

	// set up the listener
	src := s.flow.NetFlow.Src.(*net.TCPAddr)
	listener, err := net.ListenTCP(src.Network(), src)
	if err != nil {
		return shila.ThirdPartyError(fmt.Sprint("Unable to setup and run server {", s.Label(), "} listening on {", s.Key(), "}. - ", err.Error()))
	}

	// Create the channels
	s.ingress = make(chan *shila.Packet, Config.SizeIngressBuffer)
	s.egress  = make(chan *shila.Packet, Config.SizeEgressBuffer)

	// Start listening for incoming backbone connections.
	s.listener = listener
	go s.serveIncomingConnections()

	log.Verbose.Print("Server {", s.Label(), "} started to listen for incoming backbone connections on {", s.Key(), "}.")

	// Start to handle incoming packets
	go s.serveEgress()

	// Start the resending functionality
	go s.resending()

	s.state.Set(shila.Running)
	return nil
}

func (s *Server) TearDown() error {

	s.state.Set(shila.TornDown)

	// Close the listener
	// Server no longer listens for incoming connections
	err := s.listener.Close()

	// Close all incoming connections
	// Terminates all workers processing incoming an connection and the corresponding packetizer
	for _, conn := range s.backboneConnections {
		err = conn.Backbone.Close()
	}

	// Close the ingress channel
	// Working side no longer processes this endpoint
	close(s.ingress)

	return err
}

func (s *Server) TrafficChannels() shila.PacketChannels {
	return shila.PacketChannels{Ingress: s.ingress, Egress: s.egress}
}

func (s *Server) Label() shila.EndpointLabel {
	return s.label
}

func (s *Server) Key() shila.EndpointKey {
	return shila.EndpointKey(shila.GetNetworkAddressKey(s.flow.NetFlow.Src))
}

func (s *Server) serveIncomingConnections(){
	for {
		if connection, err := s.listener.Accept(); err != nil {
			return
		} else {
			go s.handleBackboneConnection(connection.(*net.TCPConn))
		}
	}
}

func (s *Server) handleBackboneConnection(backboneConnection *net.TCPConn) {

	reader := io.Reader(backboneConnection)
	lenBuffer := make([]byte, 8)
	if _, err := io.ReadFull(reader, lenBuffer); err != nil {
		s.closeBackboneConnection(backboneConnection, err); return
	}
	len := binary.BigEndian.Uint64(lenBuffer)
	buffer := make([]byte, len)
	if _, err := io.ReadFull(reader, buffer); err != nil {
		s.closeBackboneConnection(backboneConnection, err);	return
	}

	var receivedFlow shila.Flow
	decoder := gob.NewDecoder(bytes.NewReader(buffer))
	if err := decoder.Decode(&receivedFlow); err != nil {
		s.closeBackboneConnection(backboneConnection, err); return
	}

		/*
		// The very first thing we do for a accepted backbone connection is to see
		// whether we can get the corresponding flow.
		IPFlowString, err := bufio.NewReader(backboneConnection).ReadString('\n')
		if  err != nil {
			s.closeBackboneConnection(backboneConnection, err); return
		}
		IPFlow, err := shila.GetIPFlowFromString(strings.TrimSuffix(IPFlowString, "\n"))
		if err != nil {
			s.closeBackboneConnection(backboneConnection, err); return
		}
		 */
	// If we aren't able to get the flow, for whatever reason, we just throw away the backbone connection request.
	// The server remains ready to receive incoming requests.

	// Fetch the network address from the client side as well as the path taken.
	dstAddr, err := network.AddressGenerator{}.New(backboneConnection.RemoteAddr().String())
	if err != nil {
		s.closeBackboneConnection(backboneConnection, err); return
	}
	path, err	:= network.PathGenerator{}.New("")
	if err != nil {
		s.closeBackboneConnection(backboneConnection, err); return
	}

	// Determine the network address of this network endpoint depending on the functionality
	var srcAddr shila.NetworkAddress
	if s.Label() == shila.ContactingNetworkEndpoint {
		localAddr 	 := backboneConnection.LocalAddr().(*net.TCPAddr)
		srcAddr, _   = network.AddressGenerator{}.New(net.JoinHostPort(localAddr.IP.String(), strconv.Itoa(receivedFlow.IPFlow.Dst.Port)))
	} else if s.Label() == shila.TrafficNetworkEndpoint {
		srcAddr = s.flow.NetFlow.Src
	} else {
		s.closeBackboneConnection(backboneConnection, shila.CriticalError(fmt.Sprint("Wrong server label."))); return
	}

	connection := networkConnection{
		Identifier: shila.Flow{IPFlow: receivedFlow.IPFlow, NetFlow: shila.NetFlow{
			Src:  srcAddr,
			Path: path,
			Dst:  dstAddr,
		}},
		Backbone:   backboneConnection,
	}

	// Generate the keys
	var keys []shila.NetworkAddressAndPathKey
	keys = append(keys, shila.GetNetworkAddressAndPathKey(dstAddr, path))

	/*
	// Before sending any traffic data, the traffic client endpoint sends the source address of the
	// corresponding contacting client endpoint.
	if s.Label() == shila.TrafficNetworkEndpoint {
		if srcAddrReceived, err := bufio.NewReader(backboneConnection).ReadString('\n'); err != nil {
			s.closeBackboneConnection(backboneConnection, err); return
		} else {
			contactSrcAddr, _ := network.AddressGenerator{}.New(strings.TrimSuffix(srcAddrReceived,"\n"))
			keys = append(keys, shila.GetNetworkAddressAndPathKey(contactSrcAddr, path))
		}
	}
	 */

	// Add the new backboneConnection to the mapping, such that it can be found by the egress handler.
	s.lock.Lock()
	for _, key := range keys {
		if _, ok := s.backboneConnections[key]; ok {
			s.lock.Unlock()
			s.closeBackboneConnection(backboneConnection, err); return
		} else {
			s.backboneConnections[key] = connection
			log.Verbose.Print("Server {", s.Label(), "} listening on {", s.Key(), "} started handling a new backbone backboneConnection {", key, "}.")
		}
	}
	s.lock.Unlock()

	// Start the ingress handler for the backboneConnection.
	s.serveIngress(connection)

	// No longer necessary or possible to serve the ingress, remove the backboneConnection from the mapping.
	s.lock.Lock()
	for _, key := range keys {
		log.Verbose.Print("Server {", s.Label(), "} listening on {", s.Key(), "} removed backbone backboneConnection {", key, "}.")
		delete(s.backboneConnections, key)
	}
	s.lock.Unlock()

	return
}

func (s *Server) closeBackboneConnection(connection *net.TCPConn, err error) {
	connection.Close()
	log.Error.Print("Closed backbone connection in Server {", s.Label(), ",", s.Key(), "}. - ", err.Error())
}

func (s *Server) serveIngress(connection networkConnection) {

	// Prepare everything for the packetizer
	ingressRaw := make(chan byte, Config.SizeRawIngressBuffer)
	go s.packetize(connection.Identifier, ingressRaw)

	reader := io.Reader(connection.Backbone)
	storage := make([]byte, Config.SizeRawIngressStorage)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, Config.ReadSizeRawIngress)
		// If the incoming connection suffers from an error, we close it and return.
		// The server instance is still able to receive backboneConnections as long as it is not
		// shut down by the manager of the network side.
		if err != nil {
			close(ingressRaw) // Stop the packetizing.
			return
		}
		for _, b := range storage[:nBytesRead] {
			ingressRaw <- b
		}
	}
}

func (s *Server) packetize(flow shila.Flow, ingressRaw chan byte) {
	for {
		if rawData, err := tcpip.PacketizeRawData(ingressRaw, Config.SizeRawIngressStorage); rawData != nil {
				s.ingress <- shila.NewPacketWithNetFlow(s, flow.IPFlow, flow.NetFlow, rawData)
		} else {
			if err == nil {
				// All good, ingress raw closed.
				return
			}
			s.endpointIssues <- shila.EndpointIssuePub{
				Publisher: 	s,
				Flow:		flow,
				Error:     	shila.PrependError(err, "Error in raw data packetizer."),
			}
			return
		}
	}
}

func (s *Server) resending() {
	for {
		time.Sleep(Config.ServerResendInterval)
		for _, p := range s.holdingArea {
			if p.TTL > 0 {
				p.TTL--
				s.egress <- p
			} else {
				s.endpointIssues <- shila.EndpointIssuePub{
					Publisher: s,
					Flow:      p.Flow,
					Error:     shila.ThirdPartyError("Unable to write data."),
				}
			}
		}
	}
}

func (s *Server) serveEgress() {
	for p := range s.egress {
		// Retrieve key to get the correct connection
		key := p.Flow.NetFlow.DstAndPathKey()
		if con, ok := s.backboneConnections[key]; ok {
			writer := io.Writer(con.Backbone)
			_, err := writer.Write(p.Payload)
			if err != nil && s.state.Not(shila.Running) {
				// Server turned down anyway.
				return
			}
			// If just the connection was closed, then we ignore the error and drop the packet.
			// The ingress handler has observed the issue as well and will sooner or later remove
			// the closed (or faulty) connection.
		} else {
			// Currently there is no backbone connection available to send the packet.
			// It's TTL value is decreased and it is put into the holding area.
			s.holdingArea = append(s.holdingArea, p)
			log.Verbose.Print("Server {", s.Label(), "} listening on {", s.Key(), "} directs packet for " +
				"backbone connection key {", key, "} into holding area.")
		}
	}
}

func (s *Server) Flow() shila.Flow {
	return s.flow
}