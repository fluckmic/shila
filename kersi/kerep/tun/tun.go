// TODO: tun package deserves a description.
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
	"unsafe"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

type Device struct {
	Name string
	file *os.File
}

func New(name string) *Device {
	return &Device{name, nil}
}

// TODO: Allocate deserves a description.
func (d *Device) Allocate() error {

	if d.file != nil {
		return Error(fmt.Sprint("Unable to allocate tun device ",
			d.Name, " - ", "Device already allocated."))
	}

	var errno C.int
	var devName = C.CString(d.Name)
	var flags C.int = C.IFF_TUN | C.IFF_NO_PI
	fd := int(C.allocateTun(devName, &errno, flags))
	C.free(unsafe.Pointer(devName))

	if fd < 0 {
		var errorString = C.GoString(C.strerror(errno))
		return Error(fmt.Sprint("Unable to allocate tun device ",
			d.Name, " - ", errorString, "."))
	}

	d.file = os.NewFile(uintptr(fd), d.Name)
	return nil
}

// TODO: De-allocate deserves a description.
func (d *Device) Deallocate() error {
	if d.file == nil {
		return Error(fmt.Sprint("Unable to de-allocate tun device ",
			d.Name, " - ", "Device not allocated."))
	}
	err := d.file.Close()
	d.file = nil
	return err
}

func (d *Device) IsAllocated() bool {
	return d.file != nil
}

func (d *Device) Read(b []byte) (int, error) {
	if !d.IsAllocated() {
		err := Error(fmt.Sprint("Unable to read from tun device ",
			d.Name, " - ", "Device not allocated."))
		return 0, err
	}
	return d.file.Read(b)
}

func (d *Device) Write(b []byte) (int, error) {
	if !d.IsAllocated() {
		err := Error(fmt.Sprint("Unable to write to tun device ",
			d.Name, " - ", "Device not allocated."))
		return 0, err
	}
	return d.file.Write(b)
}
