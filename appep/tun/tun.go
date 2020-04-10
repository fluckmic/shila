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

import "C"
import (
	"fmt"
	"unsafe"
)

type Device struct {
}

func (d *Device) Allocate() (int, error) {

	var errno C.int
	var devName = C.CString("tun27")
	var flags C.int = C.IFF_TUN | C.IFF_NO_PI
	var fd = int(C.allocateTun(devName, &errno, flags))
	C.free(unsafe.Pointer(devName))

	fmt.Println("File descriptor:", fd, "errno:", errno)

	return 0, nil
}