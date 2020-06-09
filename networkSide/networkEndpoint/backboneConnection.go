package networkEndpoint

import "shila/core/shila"

type BackboneConnections struct {}

func (conns *BackboneConnections) retrieve(key shila.NetworkAddressKey) *BackboneConnection {
	return &BackboneConnection{}
}

func (conns *BackboneConnections) tearDown() error {
	return nil
}

func NewBackboneConnections() BackboneConnections {
	return BackboneConnections{}
}



type BackboneConnection struct {}

func (conn *BackboneConnection) writeIngress(buff []byte) {}

func (conn *BackboneConnection) writeEgress(buff []byte)  {
	/*
	if con, ok := server.backboneConnections[dstKey]; ok {
		writer := io.Writer(con.Backbone)
		_, err := writer.Write(p.Payload)
		if err != nil && server.state.Not(shila.Running) {
			return										// Server turned down anyway.
		}
		// If just the connection was closed, then we ignore the error and drop the packet. The ingress handler
		// has observed the issue as well and will sooner or later remove the closed (or faulty) connection.
	} else {
		// Currently there is no backbone connection available to send the packet, put packet into holding area.
		server.lock.Lock()
		server.holdingArea = append(server.holdingArea, p)
		server.lock.Unlock()
	}
	}
	*/
}