package interfaces

import (
	"github.com/winfsp/cgofuse/fuse"
)

type Medium interface {
	GetDevicePathname() string
	GetPublicPathname() string
	GetPublicName() string
	GetDriver() MediumDriver
	SetDevicePathname(devicePathname string)
	SetPublicPathname(publicPathname string)
	SetDriver(driver MediumDriver)
	IsFullyCached() bool
	SetFullyCached(fullyCached bool)
	SetReadable(readable bool)
	SetWritable(writable bool)
	IsReadable() bool
	IsWritable() bool
	SetCreateTime(creationTime int64)
	SetAccessTime(accessTime int64)
	SetModificationTime(modificationTime int64)
	GetCreateTime() int64
	GetAccessTime() int64
	GetModificationTime() int64
	Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int)
	Read(path string, buff []byte, ofst int64, fh uint64) (n int)
	Write(path string, buff []byte, ofst int64, fh uint64) int
}
