//
package tun

/*
 #include <errno.h>
 #include <fcntl.h>
 #include <linux/if_tun.h>
 #include <net/if.h>
 #include <stdio.h>
 #include <stdlib.h>
 #include <string.h>
 #include <sys/ioctl.h>
 #include <sys/stat.h>
 #include <unistd.h>

 int allocateTun(char* devName, int* pErrno, int flags) {

	int fd;
	char* clonedev = "/dev/net/tun";

  	if( (fd = open(clonedev , O_RDWR)) < 0 ) {
    	*pErrno = (int) errno;
		return fd;
 	}

	struct ifreq ifr;
	memset(&ifr, 0, sizeof(ifr));
	ifr.ifr_flags = flags;

    if (*devName) {
		strncpy(ifr.ifr_name, devName, IFNAMSIZ);
    }

	int err;
	if( (err = ioctl(fd, TUNSETIFF, (void *)&ifr)) < 0 ) {
    	close(fd);
		*pErrno = (int) errno;
    	return err;
  	}

	strcpy(devName, ifr.ifr_name);

  	return fd;
 }
*/
import "C"

import (
	"fmt"
	"os"
	"shila/core/shila"
	"unsafe"
)

type Device struct {
	Name 	string
	file 	*os.File
	state   shila.EntityState
}

func New(name string) Device {
	return Device{
		Name: 	name,
		state: 	shila.NewEntityState(),
	}
}

func (device *Device) Allocate() error {

	if device.state.Not(shila.Uninitialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state ", device.state, "."))
	}

	var errno C.int
	var devName = C.CString(device.Name)
	var flags C.int = C.IFF_TUN | C.IFF_NO_PI
	fd := int(C.allocateTun(devName, &errno, flags))
	C.free(unsafe.Pointer(devName))

	if fd < 0 {
		var errorString = C.GoString(C.strerror(errno))
		return shila.CriticalError(errorString)
	}

	device.file = os.NewFile(uintptr(fd), device.Name)
	device.state.Set(shila.Initialized)
	return nil
}

func (device *Device) Deallocate() error {
	device.state.Set(shila.TornDown)
	err := device.file.Close()
	device.file = nil
	return err
}

func (device *Device) Read(b []byte) (int, error) {
	if device.state.Not(shila.Initialized) {
		return -1, shila.CriticalError(fmt.Sprint("Entity in wrong state ", device.state, "."))
	}
	return device.file.Read(b)
}

func (device *Device) Write(b []byte) (int, error) {
	if device.state.Not(shila.Initialized) {
		return -1, shila.CriticalError(fmt.Sprint("Entity in wrong state ", device.state, "."))
	}
	return device.file.Write(b)
}

func (device *Device) Says(str string) string {
	return  fmt.Sprint(device.Identifier(), ": ", str)
}

func (device *Device) Identifier() string {
	return device.Name
}