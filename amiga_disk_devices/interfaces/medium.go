package interfaces

import (
	"os"
	"sync"

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
	Close() error
	GetHandle() (*os.File, error)
	SetHandle(handle *os.File)
	GetMutex() *sync.Mutex
	GetSize() int64
	SetSize(size int64)
	AddPreReadCallback(preReadCallback PreReadCallback)
	AddPostReadCallback(postReadCallback PostReadCallback)
	AddPreWriteCallback(preWriteCallback PreWriteCallback)
	AddPostWriteCallback(postWriteCallback PostWriteCallback)
	AddClosedCallback(closedCallback ClosedCallback)
	CallPreReadCallbacks(_medium Medium, path string, buff []byte, ofst int64, fh uint64)
	CallPreWriteCallbacks(_medium Medium, path string, buff []byte, ofst int64, fh uint64)
	CallPostReadCallbacks(
		_medium Medium,
		path string,
		buff []byte,
		ofst int64,
		fh uint64,
		n int,
		opTimeMs int64,
	)
	CallPostWriteCallbacks(
		_medium Medium,
		path string,
		buff []byte,
		ofst int64,
		fh uint64,
		n int,
		opTimeMs int64,
	)
	CallClosedCallbacks(_medium Medium, err error)
	DevicePathnameToPublicFilename(devicePathname string, extension string) string
	Open(path string, flags int) (errc int, fh uint64)
	Getattr(path string, stat *fuse.Stat_t, fh uint64) (int, error)
	Read(path string, buff []byte, ofst int64, fh uint64) (int, error)
	Write(path string, buff []byte, ofst int64, fh uint64) (int, error)
}
