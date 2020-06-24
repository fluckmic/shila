package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"io"
	"os"
)

const (
	nIncomingMsg = 1
)

func main() {

	// Get local and remote addresses from program arguments:
	port 		:= flag.Uint("port", 0, "[Server] Local port to listen on.")
	remoteAddr 	:= flag.String("remote", "", "[Client] Remote SCION Address.")
	name 		:= flag.String("name", "", "Name of the entity.")

	flag.Parse()

	if len(*name) < 1 {
		os.Exit(1)
	}

	if (*port > 0) == (len(*remoteAddr) > 0) {
		os.Exit(1)
	}

	if *port > 0 {
		err := runServer(uint16(*port), *name)
		check(err)
	} else {
		err := runClient(*remoteAddr, *name)
		check(err)
	}
}

func runServer(port uint16, name string) error {

	conn, err := appnet.ListenPort(port)
		if err != nil {
			return err
		}

	fmt.Print("Server ", name, " ready to receive ", nIncomingMsg, " incoming messages.\n")

	for i := 0; i < nIncomingMsg; i++ {

		// Receive control message
		var ctrlMsgR controlMessage
		if err := gob.NewDecoder(io.Reader(conn)).Decode(&ctrlMsgR); err != nil {
			return err
		} else {
			fmt.Print("Server ", name, " received control message from ", ctrlMsgR.Name, ".\n")
		}

		// Send control message
		var buffer bytes.Buffer
		if err := gob.NewEncoder(&buffer).Encode(controlMessage{Name: name}); err != nil {
			return err
		}

		if _, err := conn.WriteTo(buffer.Bytes(), conn.RemoteAddr()); err != nil {
			return err
		}

		fmt.Print("Server ", name, " sent control message back to ", ctrlMsgR.Name, ".\n")
	}

	conn.Close()
	return nil
}

func runClient(address string, name string) error {

	conn, err := appnet.Dial(address)
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Print("Client ", name, " connected to ", address, ".\n")

	// Send control message
	if err := gob.NewEncoder(io.Writer(conn)).Encode(controlMessage{ Name: name}); err != nil {
		return err
	}

	fmt.Print("Client ", name, " sent control message to ", address, ".\n")

	// Receive control message
	var ctrlMsgR controlMessage
	if err := gob.NewDecoder(io.Reader(conn)).Decode(&ctrlMsgR); err != nil {
		return err
	} else {
		fmt.Print("Client ", name, " received control message from ", ctrlMsgR.Name, ".\n")
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
	Name string
}