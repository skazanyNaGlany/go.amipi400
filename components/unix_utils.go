package components

import "syscall"

type UnixUtils struct{}

func (k *UnixUtils) IsRoot() bool {
	return syscall.Getuid() == 0 && syscall.Geteuid() == 0
}
