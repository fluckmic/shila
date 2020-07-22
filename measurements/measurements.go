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
	"fmt"
	"shila/log"
)

const (
	NPayloadBytesSkipped = 80
)

func LogTimestamp(payload []byte) {

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

	fmt.Printf("%v, %v, %v, %v, %v, %v, %v\n", sec, nSec, A, B, C, D, E)

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