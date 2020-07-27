package measurements
/*

#include <sys/time.h>
#include <time.h>

int get_now( long* sec, long* nsec)
{
	struct timespec time;

	if (clock_gettime(CLOCK_REALTIME, &time) != 0 )
	{
		return -1;
	}

	*sec  = time.tv_sec;
	*nsec = time.tv_nsec;

	return 0;
}

*/
import "C"
import (
	"bufio"
	"fmt"
	"os"
	"shila/config"
	"shila/log"
	"time"
)

const (
	NPayloadBytesSkipped = 80
)

var (
	egressTimestampWriter *bufio.Writer
	ingressTimestampWriter *bufio.Writer
)

func init() {

	if config.Config.Logging.DoEgressTimestamping {
		file, err := os.Create(config.Config.Logging.EgressTimestampLogPath)
		if err == nil {

			egressTimestampWriter = bufio.NewWriter(file)

			if config.Config.Logging.EgressTimestampLogAdditionalLine != "" {
				_, _ = egressTimestampWriter.WriteString(fmt.Sprintf("%s\n", config.Config.Logging.EgressTimestampLogAdditionalLine))
				egressTimestampWriter.Flush()
			}

			go func() {
				for {
					time.Sleep(time.Duration(config.Config.Logging.TimestampFlushInterval) * time.Second)
					egressTimestampWriter.Flush()
				}
			}()

		}
	}

	if config.Config.Logging.DoIngressTimestamping {
		file, err := os.Create(config.Config.Logging.IngressTimestampLogPath)
		if err == nil {

			ingressTimestampWriter = bufio.NewWriter(file)

			if config.Config.Logging.IngressTimestampLogAdditionalLine != "" {
				_, _ = ingressTimestampWriter.WriteString(fmt.Sprintf("%s\n", config.Config.Logging.IngressTimestampLogAdditionalLine))
				ingressTimestampWriter.Flush()
			}

			go func() {
				for {
					time.Sleep(time.Duration(config.Config.Logging.TimestampFlushInterval) * time.Second)
					ingressTimestampWriter.Flush()
				}
			}()

		}
	}

}

func LogIngressTimestamp(payload []byte) {
	if ingressTimestampWriter != nil {
		logTimestamp(payload, ingressTimestampWriter)
	}
}

func LogEgressTimestamp(payload []byte) {
	if egressTimestampWriter != nil {
		logTimestamp(payload, egressTimestampWriter)
	}
}

func logTimestamp(payload []byte, writer *bufio.Writer) {

	if writer == nil {
		return
	}

	sec, nSec, err := getNow()
	if err != nil {
		log.Error.Print("Unable to log timestamp. ", err.Error())
		return
	}

	A, B, C, D, E, err := getPacketIdentifier(payload)
	if err != nil {
		//log.Error.Print("Unable to log timestamp. ", err.Error())
		return
	}

	_, _ = writer.WriteString(fmt.Sprintf("%v, %v, %v, %v, %v, %v, %v\n", sec, nSec, A, B, C, D, E))

}

func getNow() (int64, int64, error) {

	var sec  C.long
	var nSec C.long
	ret := int(C.get_now(&sec, &nSec))

	if ret < 0 {
		return -1,-1, Error("Error in get_now().")
	} else {
		return int64(sec), int64(nSec), nil
	}
}

func getPacketIdentifier(payload []byte) (byte, byte, byte, byte, byte, error) {

	for i := NPayloadBytesSkipped; i < len(payload) - 5; i++ {

		if payload[i] == 1 {
			//	   A			 B			   C			 D			   E
			return payload[i+5], payload[i+4], payload[i+3], payload[i+2], payload[i+1], nil
		}
	}

	return 0, 0, 0, 0, 0, Error("Unable to fetch packet identifier.")
}

type Error string
func (e Error) Error() string {
	return string(e)
}