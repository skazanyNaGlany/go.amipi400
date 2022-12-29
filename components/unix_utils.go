package components

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

type UnixUtils struct{}

var UnixUtilsInstance UnixUtils

func (k *UnixUtils) IsRoot() bool {
	return syscall.Getuid() == 0 && syscall.Geteuid() == 0
}

func (k *UnixUtils) SetDeviceReadAHead(handle *os.File, readAHead int) error {
	// set read-a-head value for block-device
	if err := unix.IoctlSetInt(int(handle.Fd()), unix.BLKRASET, readAHead); err != nil {
		return err
	}

	// set read-a-head value for file-system
	if err := unix.IoctlSetInt(int(handle.Fd()), unix.BLKFRASET, readAHead); err != nil {
		return err
	}

	return nil
}
