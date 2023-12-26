package utils

import (
	"log"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

type UnixUtils struct{}

var UnixUtilsInstance UnixUtils

func (k *UnixUtils) IsRoot() bool {
	return syscall.Getuid() == 0 && syscall.Geteuid() == 0
}

func (k *UnixUtils) CheckForRoot() {
	if !k.IsRoot() {
		log.Fatalln("Must be run as root.")
	}
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

func (k *UnixUtils) RunFsck(devicePathname string) (string, error) {
	output, err := exec.Command(
		"fsck",
		"-y",
		devicePathname).CombinedOutput()

	return string(output), err
}

func (k *UnixUtils) Sync() (string, error) {
	output, err := exec.Command(
		"sync",
	).CombinedOutput()

	return string(output), err
}

func (k *UnixUtils) Shutdown() (string, error) {
	output, err := exec.Command(
		"shutdown",
		"-h",
		"now",
	).CombinedOutput()

	return string(output), err
}
