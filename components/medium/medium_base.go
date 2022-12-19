package medium

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"github.com/winfsp/cgofuse/fuse"
)

// Implements Medium
type MediumBase struct {
	devicePathname   string
	publicPathname   string
	publicName       string
	driver           interfaces.MediumDriver
	fullyCached      bool
	readable         bool
	writable         bool
	creationTime     int64
	accessTime       int64
	modificationTime int64
	handle           *os.File
	mutex            sync.Mutex
	size             int64
}

func (mb *MediumBase) SetCreateTime(creationTime int64) {
	mb.creationTime = creationTime
}

func (mb *MediumBase) SetAccessTime(accessTime int64) {
	mb.accessTime = accessTime
}

func (mb *MediumBase) SetModificationTime(modificationTime int64) {
	mb.modificationTime = modificationTime
}

func (mb *MediumBase) GetCreateTime() int64 {
	return mb.creationTime
}

func (mb *MediumBase) GetAccessTime() int64 {
	return mb.accessTime
}

func (mb *MediumBase) GetModificationTime() int64 {
	return mb.modificationTime
}

func (mb *MediumBase) SetReadable(readable bool) {
	mb.readable = readable
}

func (mb *MediumBase) SetWritable(writable bool) {
	mb.writable = writable
}

func (mb *MediumBase) IsReadable() bool {
	return mb.readable
}

func (mb *MediumBase) IsWritable() bool {
	return mb.writable
}

func (mb *MediumBase) GetPublicName() string {
	return mb.publicName
}

func (mb *MediumBase) IsFullyCached() bool {
	return mb.fullyCached
}

func (mb *MediumBase) SetFullyCached(fullyCached bool) {
	mb.fullyCached = fullyCached
}

func (mb *MediumBase) SetDevicePathname(devicePathname string) {
	mb.devicePathname = devicePathname
}

func (mb *MediumBase) SetPublicPathname(publicPathname string) {
	mb.publicPathname = publicPathname
	mb.publicName = filepath.Base(publicPathname)
}

func (mb *MediumBase) SetDriver(driver interfaces.MediumDriver) {
	mb.driver = driver
}

func (mb *MediumBase) GetDevicePathname() string {
	return mb.devicePathname
}

func (mb *MediumBase) GetPublicPathname() string {
	return mb.publicPathname
}

func (mb *MediumBase) GetDriver() interfaces.MediumDriver {
	return mb.driver
}

func (mb *MediumBase) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	return mb.driver.Getattr(mb, path, stat, fh)
}

func (mb *MediumBase) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	return mb.driver.Read(mb, path, buff, ofst, fh)
}

func (mb *MediumBase) Write(path string, buff []byte, ofst int64, fh uint64) int {
	return mb.driver.Write(mb, path, buff, ofst, fh)
}

func (mb *MediumBase) Close() error {
	if mb.handle != nil {
		return mb.handle.Close()
	}

	return nil
}

func (mb *MediumBase) GetHandle() (*os.File, error) {
	if mb.handle != nil {
		return mb.handle, nil
	}

	isReadable := mb.IsReadable()
	isWritable := mb.IsWritable()

	flag := os.O_SYNC

	if isReadable && isWritable {
		flag |= os.O_RDWR
	} else {
		flag |= os.O_RDONLY
	}

	handle, err := os.OpenFile(
		mb.GetDevicePathname(),
		flag,
		0,
	)

	if err != nil {
		return nil, err
	}

	mb.handle = handle

	return nil, nil
}

func (mb *MediumBase) SetHandle(handle *os.File) {
	mb.handle = handle
}

func (mb *MediumBase) GetMutex() *sync.Mutex {
	return &mb.mutex
}

func (mb *MediumBase) GetSize() int64 {
	return mb.size
}

func (mb *MediumBase) SetSize(size int64) {
	mb.size = size
}
