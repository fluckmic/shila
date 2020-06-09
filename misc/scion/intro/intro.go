// Copyright 2018 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
	"io"
	"net"
	"os"
	"shila/core/shila"
)

func main() {
	var err error
	// get local and remote addresses from program arguments:
	port := flag.Uint("port", 0, "[Server] local port to listen on")
	remoteAddr := flag.String("remote", "", "[Client] Remote (i.e. the server's) SCION Address (e.g. 17-ffaa:1:1,[127.0.0.1]:12345)")
	flag.Parse()

	if (*port > 0) == (len(*remoteAddr) > 0) {
		check(fmt.Errorf("Either specify -port for server or -remote for client."))
	}

	if *port > 0 {
		err = runServer(uint16(*port))
		check(err)
	} else {
		runClient(*remoteAddr)
		check(err)
	}
}

func runServer(port uint16) error {
		conn, err := appnet.ListenPort(port)
		if err != nil {
			return err
		}
		err = handleConnection(conn)
		conn.Close()
		return err
}

func handleConnection(conn *snet.Conn) error {

	pr, pw := io.Pipe()
	go decoder(pr)

	buffer := make([]byte, 32*1024)

	for {
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			return err
		}
		pw.Write(buffer[:n])
	}
}

func runClient(address string) error {

	conn, err := appnet.Dial(address)
	if err != nil {
		return err
	}
	defer conn.Close()

	to := conn.RemoteAddr().(*snet.UDPAddr)
	_ = to

	// Send the control msg to the server
	ctrlMsg := controlMessage{
		IPFlow:   				shila.IPFlow{
			Src: net.TCPAddr{
				IP:   net.IPv4(1,2,3,4),
				Port: 4141,
				Zone: "",
			},
			Dst: net.TCPAddr{
				IP:   net.IPv4(9,8,7,6),
				Port: 2727,
				Zone: "",
			},
		},
		Contact: net.UDPAddr{
			IP:   net.IPv4(7,4,7,4),
			Port: 5000,
			Zone: "",
		},
		Payload: []byte("I'm Payload."),
	}
	if err := gob.NewEncoder(io.Writer(conn)).Encode(ctrlMsg); err != nil {
		return shila.PrependError(err, "Failed to transmit control message.")
	}

	return nil
}


// Check just ensures the error is nil, or complains and quits
func check(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, "Fatal error. Exiting.", "err", e)
		os.Exit(1)
	}
}

type controlMessage struct {
	IPFlow  shila.IPFlow
	Contact net.UDPAddr
	Payload []byte
}

func decoder(reader *io.PipeReader) {
	for {
		var ctrlMsg controlMessage
		if err := gob.NewDecoder(reader).Decode(&ctrlMsg); err != nil {
			panic("Cannot fetch control message.")
		} else {
			fmt.Println("Received control message: ", ctrlMsg)
		}
	}
}