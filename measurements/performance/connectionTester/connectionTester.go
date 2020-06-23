package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
	"io"
	"os"
)

const (
	nIncomingMsg = 3
)

func main() {

	var err error

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
		err = runServer(uint16(*port), *name)
		check(err)
	} else {
		runClient(*remoteAddr, *name)
		check(err)
	}
}

func runServer(port uint16, name string) error {

	conn, err := appnet.ListenPort(port)
		if err != nil {
			return err
		}
		err = handleConnection(conn, name)
		conn.Close()

		return err
}

func handleConnection(conn *snet.Conn, name string) error {

	pr, pw := io.Pipe()
	go decoder(pr, conn, name)

	buffer := make([]byte, 32*1024)

	for {
		n, from, err := conn.ReadFrom(buffer)
		if err != nil {
			return err
		}
		_ = from
		pw.Write(buffer[:n])
	}
}

func runClient(address string, name string) error {

	conn, err := appnet.Dial(address)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Send control message
	if err := gob.NewEncoder(io.Writer(conn)).Encode(controlMessage{ Name: name}); err != nil {
		return err
	}

	// Receive control message
	var ctrlMsgR controlMessage
	if err := gob.NewDecoder(io.Reader(conn)).Decode(&ctrlMsgR); err != nil {
		return err
	} else {
		fmt.Print(name, "exchanged control message with", ctrlMsgR.Name, ".\n")
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

func decoder(reader *io.PipeReader, conn *snet.Conn, name string) error {
	for i := 0; i < nIncomingMsg; i++ {
		var ctrlMsg controlMessage
		if err := gob.NewDecoder(reader).Decode(&ctrlMsg); err != nil {
			return err
		}
		if err := gob.NewEncoder(io.Writer(conn)).Encode(controlMessage{Name: name}); err != nil {
			return err
		}
	}
	return nil
}