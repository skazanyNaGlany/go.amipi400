package components

import "syscall"

type UnixUtils struct{}

var UnixUtilsInstance UnixUtils

func (k *UnixUtils) IsRoot() bool {
	return syscall.Getuid() == 0 && syscall.Geteuid() == 0
}
