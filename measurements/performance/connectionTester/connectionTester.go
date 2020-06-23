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
		err = runServer(uint16(*port))
		check(err)
	} else {
		runClient(*remoteAddr, *name)
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

	// Create control message
	ctrlMsg := controlMessage{
		Name:	name,
	}

	if err := gob.NewEncoder(io.Writer(conn)).Encode(ctrlMsg); err != nil {
		return err
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

func decoder(reader *io.PipeReader) error {
	for i := 0; i < nIncomingMsg; i++ {
		var ctrlMsg controlMessage
		if err := gob.NewDecoder(reader).Decode(&ctrlMsg); err != nil {
			return err
		fmt.Print("Received control message from ", ctrlMsg.Name, ".\n")
		}
	}
	return nil
}