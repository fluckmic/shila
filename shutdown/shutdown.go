// TODO: Borrowed from scion lib - add proper mentioning!

// Package shutdown provides a way to handle shutdown errors.
// 1. It gives the main goroutine an opportunity to cleanly shut down in case of a shutdown error.
// 2. If main goroutine is non-responsive it terminates the process.
// 3. To improve debugging, after the first shutdown error the other goroutines
//    are given a grace period so that we have more logs to investigate.
//
// Also implemented is a clean shutdown option, for non-error events that
// trigger clean application shutdown (e.g., a signal received from the user).
//
// The main program should call shutdown.Init() when it's starting.
//
// Any library producing shutdown errors should call shutdown.Check() when it starts.
package shutdown

import (
	"os"
	"os/signal"
	"shila/log"
	"sync"
	"time"
)

const (
	// DelayInterval is the interval between when a package signals that a
	// shutdown event has occurred, and when drainers of the shutdown channel are
	// informed. This allows for collecting more informative logs before
	// tearing the application down.
	DelayInterval = time.Second
	// GraceInterval is the time in which the main goroutine should shut
	// the application down.
	GraceInterval = 5 * time.Second
)

var (
	initialized bool

	fatalClosed   bool
	orderlyClosed bool

	fatalMtx   sync.Mutex
	orderlyMtx sync.Mutex

	// Used for signals asking for clean shutdown within the app
	orderlyChannel chan struct{}
	// Used to catch signals from the os (SIGTERM, SIGINT)
	signalChannel chan os.Signal
	// Used for signals asking for forceful termination
	fatalChannel chan struct{}
)

type Error string

func (e Error) Error() string {
	return string(e)
}

func Init() {
	orderlyChannel = make(chan struct{})
	fatalChannel = make(chan struct{})

	signalChannel = make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, os.Kill)

	go func() {
		defer signal.Reset(os.Interrupt, os.Kill)
		select {
		case <-signalChannel:
			Orderly(GraceInterval)
		}
	}()

	initialized = true
}

// Not checks whether the package was initialized.
// MUST be called when a library using the shutdown package,
// e.g. produces shutdown errors, is initialized.
func Check() {
	if !initialized {
		// TODO!
		panic(Error("Library uses shutdown package which wasn't initialized."))
	}
}

// Fatal produces a shutdown error. This function never exits.
func Fatal(err error) {

	log.Error.Println("Fatal error.", err.Error())

	// Grace period to gather more logs in case that
	// the first shutdown error wasn't the most informative one.
	time.Sleep(DelayInterval)

	// Ask main goroutine to shut down the application.
	fatalMtx.Lock()
	if !fatalClosed {
		close(fatalChannel)
		fatalClosed = true

		// If the main goroutine fatals out correctly,
		// this won't get a chance to run.
		time.AfterFunc(GraceInterval, func() {
			log.Error.Fatalln("Main goroutine did not shut down within {", GraceInterval, "} - Forcing shutdown.")
		})
	}
	fatalMtx.Unlock()

	select {}
}

// Orderly closes the orderly channel, thus informing channel
// drainers (usually the main goroutine) that the application should be cleanly
// shut down. If the application does not shut down in the specified duration,
// it is forcefully torn down.
//
// Shutdown blocks forever.
func Orderly(d time.Duration) {

	log.Info.Print("Shutdown initiated. - Wait {", d, "} until forceful shutdown.")

	// Inform drainer if not informed already
	orderlyMtx.Lock()
	if !orderlyClosed {
		close(orderlyChannel)
		orderlyClosed = true

		// If the main goroutine shuts down everything in time,
		// this won't get a chance to run.
		time.AfterFunc(d, func() {
			log.Error.Fatal("Main goroutine did not shut down within {", d, "} - Forcing shutdown.")
		})
	}
	orderlyMtx.Unlock()

	select {}
}

// FatalChan returns a read-only channel that is closed
// when a shutdown condition has occurred.
func FatalChan() <-chan struct{} {
	return fatalChannel
}

// OrderlyChan returns a read-only channel that is closed
// when the application should be cleanly shut down.
func OrderlyChan() <-chan struct{} {
	return orderlyChannel
}
