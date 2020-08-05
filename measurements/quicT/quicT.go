// Copyright 2017 ETH Zurich
// Copyright 2018 ETH Zurich, Anapaya Systems
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

// Simple application for SCION connectivity using the snet library.
package main

import (
	"context"
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lucas-clemente/quic-go"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/integration"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	sd "github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/squic"
	"github.com/scionproto/scion/go/lib/sock/reliable"
)

const (
	DefaultTimeout = 2 * time.Second
	BufferSize     = 100000 // 100 KBytes
	BytesPerMByte  = 1000000
	MaxNMBytes     = 1 << 16
	ModeServer     = "server"
	ModeClient     = "client"
	KeyPath        = "/home/scion/go/src/github.com/scionproto/scion/gen-certs/tls.key"
	CertPath       = "/home/scion/go/src/github.com/scionproto/scion/gen-certs/tls.pem"
)

var (
	local, remote snet.UDPAddr
	payloadData   []byte

	nMBytes    = flag.Int("n", 10,"Number of MBytes to transfer (default is 10)")
	dispatcher = flag.String("dispatcher", "", "Path to dispatcher socket")
	mode       = flag.String("mode", ModeClient, "Run in "+ModeClient+" or "+ModeServer+" mode")
	sciondAddr = flag.String("sciond", sciond.DefaultSCIONDAddress, "SCIOND address")
	timeout    = flag.Duration("timeout", DefaultTimeout, "Timeout for the ping response")
	verbose    = flag.Bool("v", false, "sets verbose output")
	reverse    = flag.Bool("R", false, "run in reverse mode (server sends, client receives)")
	logConsole string
)

func init() {
	flag.Var(&local, "local", "(Mandatory) address to listen on")
	flag.Var(&remote, "remote", "(Mandatory for clients) address to connect to")
	flag.StringVar(&logConsole, "log.console", "crit",
		"Console logging level: trace|debug|info|warn|error|crit")
}

func main() {
	os.Setenv("TZ", "UTC")
	validateFlags()
	logCfg := log.Config{Console: log.ConsoleConfig{Level: logConsole}}
	if err := log.Setup(logCfg); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s", err)
		flag.Usage()
		os.Exit(1)
	}
	defer log.HandlePanic()
	initNetwork()
	switch *mode {
	case ModeClient:
		c := newClient()
		setSignalHandler(c)
		c.run()
	case ModeServer:
		server{}.run()
	}
}

func validateFlags() {
	flag.Parse()
	if *mode != ModeClient && *mode != ModeServer {
		LogFatal("Unknown mode, must be either '" + ModeClient + "' or '" + ModeServer + "'")
	}
	if *mode == ModeClient {
		if remote.Host == nil {
			LogFatal("Missing remote address")
		}
		if remote.Host.Port == 0 {
			LogFatal("Invalid remote port", "remote port", remote.Host.Port)
		}
	}
	if local.Host == nil {
		LogFatal("Missing local address")
	}
	if *nMBytes < 0 || *nMBytes > MaxNMBytes {
		LogFatal("Invalid nMBytes", "min", 0, "max", MaxNMBytes, "actual", *nMBytes)
	}
}

func LogFatal(msg string, a ...interface{}) {
	log.Crit(msg, a...)
	os.Exit(1)
}

func initNetwork() {
	if err := squic.Init(KeyPath, CertPath); err != nil {
		LogFatal("Unable to initialize QUIC/SCION", "err", err)
	}
	log.Debug("QUIC/SCION successfully initialized")
}

type message struct {
	Data      []byte
}

type rMessage struct {
	NMBytes int
}

func requestRMsg() *rMessage {
	return &rMessage{
		NMBytes: *nMBytes,
	}
}

func requestMsg() *message {
	return &message{
		Data: payloadData,
	}
}

type quicStream struct {
	qstream quic.Stream
	encoder *gob.Encoder
	decoder *gob.Decoder
}

func newQuicStream(qstream quic.Stream) *quicStream {
	return &quicStream{
		qstream,
		gob.NewEncoder(qstream),
		gob.NewDecoder(qstream),
	}
}

func (qs quicStream) WriteRMsg(msg *rMessage) error {
	return qs.encoder.Encode(msg)
}

func (qs quicStream) ReadRMsg() (*rMessage, error) {
	var msg rMessage
	err := qs.decoder.Decode(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, err
}

func (qs quicStream) WriteMsg(msg *message) error {
	return qs.encoder.Encode(msg)
}

func (qs quicStream) ReadMsg() (*message, error) {
	var msg message
	err := qs.decoder.Decode(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, err
}

type client struct {
	*quicStream
	qsess quic.Session
}

func newClient() *client {
	return &client{}
}

// run dials to a remote SCION address and repeatedly sends ping messages
// while receiving pong messages. For each successful ping-pong, a message
// with the round trip time is printed.
func (c *client) run() {
	// Needs to happen before Dial, as it will 'copy' the remote to the connection.
	// If remote is not in local AS, we need a path!
	c.setupPath()
	defer c.Close()

	ds := reliable.NewDispatcher(*dispatcher)
	sciondConn, err := sd.NewService(*sciondAddr).Connect(context.Background())
	if err != nil {
		LogFatal("Unable to initialize SCION network", "err", err)
	}
	network := snet.NewNetworkWithPR(local.IA, ds, sd.Querier{
		Connector: sciondConn,
		IA:        local.IA,
	}, sd.RevHandler{Connector: sciondConn})

	// Connect to remote address. Note that currently the SCION library
	// does not support automatic binding to local addresses, so the local
	// IP address needs to be supplied explicitly. When supplied a local
	// port of 0, Dial will assign a random free local port.

	c.qsess, err = squic.Dial(network, local.Host, &remote, addr.SvcNone, nil)
	if err != nil {
		LogFatal("Unable to dial", "err", err)
	}

	qstream, err := c.qsess.OpenStreamSync()
	if err != nil {
		LogFatal("quic OpenStream failed", "err", err)
	}
	c.quicStream = newQuicStream(qstream)
	log.Debug("Quic stream opened", "local", &local, "remote", &remote)

	if *reverse {
		// Server sends, client receives
		log.Info("Instance is receiving party.")

		if err = c.quicStream.WriteRMsg(requestRMsg()); err != nil {
			LogFatal("Failed to write hi message.", "err", err)
		}

		blackHoleReceive(c.quicStream)

	} else {
		// Client sends, server receives
		defer log.HandlePanic()
		log.Info("Instance is sending party.")

		goodput := determineGoodput(c.quicStream)
		fmt.Printf("%v MBytes/s\n", goodput)
	}

}

func (c *client) Close() error {
	var err error
	if c.qstream != nil {
		err = c.qstream.Close()
	}
	if err == nil && c.qsess != nil {
		// Note closing the session here is fine since we know that all the traffic went through.
		// If you are not sure that this is the case you should probably not close the session.
		// E.g. if you are just sending something to a server and closing the session immediately
		// it might be that the server does not see the message.
		// See also: https://github.com/lucas-clemente/quic-go/issues/464
		err = c.qsess.Close()
	}
	return err
}

func (c client) setupPath() {
	if !remote.IA.Equal(local.IA) {
		path := choosePath()
		if path == nil {
			LogFatal("No paths available to remote destination")
		}
		remote.Path = path.Path()
		remote.NextHop = path.OverlayNextHop()
	}
}

type server struct {
}

// run listens on a SCION address and replies to any ping message.
// On any error, the server exits.
func (s server) run() {
	ds := reliable.NewDispatcher(*dispatcher)
	sciondConn, err := sd.NewService(*sciondAddr).Connect(context.Background())
	if err != nil {
		LogFatal("Unable to initialize SCION network", "err", err)
	}
	network := snet.NewNetworkWithPR(local.IA, ds, &sd.Querier{
		Connector: sciondConn,
		IA:        local.IA,
	}, sd.RevHandler{Connector: sciondConn})
	if err != nil {
		LogFatal("Unable to initialize SCION network", "err", err)
	}
	qsock, err := squic.Listen(network, local.Host, addr.SvcNone, nil)
	if err != nil {
		LogFatal("Unable to listen", "err", err)
	}
	if len(os.Getenv(integration.GoIntegrationEnv)) > 0 {
		// Needed for integration test ready signal.
		fmt.Printf("Port=%d\n", qsock.Addr().(*net.UDPAddr).Port)
		fmt.Printf("%s%s\n", integration.ReadySignal, local.IA)
	}
	log.Info("Listening", "local", qsock.Addr())
	for {
		qsess, err := qsock.Accept()
		if err != nil {
			log.Error("Unable to accept quic session", "err", err)
			// Accept failing means the socket is unusable.
			break
		}
		log.Info("Quic session accepted", "src", qsess.RemoteAddr())
		go func() {
			defer log.HandlePanic()
			s.handleClient(qsess)
		}()
	}
}

func (s server) handleClient(qsess quic.Session) {
	defer qsess.Close()
	qstream, err := qsess.AcceptStream()
	if err != nil {
		log.Error("Unable to accept quic stream", "err", err)
		return
	}
	defer qstream.Close()

	qs := newQuicStream(qstream)

	if *reverse {
		// Server sends, client receives
		log.Info("Instance is sending party.")

		if rMsg, err := qs.ReadRMsg(); err != nil {
			LogFatal("Failed to receive r message", "err", err)
		} else {
			nMBytes = &rMsg.NMBytes
		}

		goodput := determineGoodput(qs)
		fmt.Printf("%v MBytes/s\n", goodput)

	} else {
		// Client sends, server receives
		defer log.HandlePanic()
		blackHoleReceive(qs)
	}
}

func choosePath() snet.Path {
	var pathIndex uint64

	sdConn, err := sd.NewService(*sciondAddr).Connect(context.Background())
	if err != nil {
		LogFatal("Unable to initialize SCION network", "err", err)
	}
	paths, err := sdConn.Paths(context.Background(), remote.IA, local.IA, sd.PathReqFlags{})
	if err != nil {
		LogFatal("Failed to lookup paths", "err", err)
	}

	//fmt.Printf("Using path:\n  %s\n", fmt.Sprintf("%s", paths[pathIndex]))
	return paths[pathIndex]
}

func blackHoleReceive(qs *quicStream) {
	log.Info("Black hole receiving started.")
	for {
		_, err := qs.ReadMsg()
		if err != nil {
			log.Error("Unable to read", "err", err)
			break
		}
	}
	log.Info("Black hole receiving stopped.")
}

func goodputSend(qs *quicStream) {

	log.Info("Goodput sending started.")

	// Allocate the dummy load which is send over and over..
	allocatePayloadData()

	nCycles := *nMBytes * (BytesPerMByte / BufferSize)

	log.Info("Number of sending cycles", "nCycles", nCycles)

	reqMsg := requestMsg()
	for i := 0; i < nCycles; i++ {
		err := qs.WriteMsg(reqMsg)
		if err != nil {
			log.Error("Unable to write", "err", err)
			continue
		}
	}

	log.Info("Goodput sending stopped.")
}

func determineGoodput(qs *quicStream) float64 {

	before := time.Now()

	goodputSend(qs)

	after := time.Now()
	elapsed := after.Sub(before).Round(time.Microsecond)

	return float64(*nMBytes) / elapsed.Seconds()
}

func allocatePayloadData() {
	payloadData = make([]byte, 0, BufferSize)
	for i := 0; i < BufferSize; i++ {
		payloadData = append(payloadData, byte(rand.Int()))
	}
}

func setSignalHandler(closer io.Closer) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer log.HandlePanic()
		<-c
		closer.Close()
		os.Exit(1)
	}()
}
