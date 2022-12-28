package medium

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"github.com/winfsp/cgofuse/fuse"
)

// Implements Medium
type MediumBase struct {
	devicePathname     string
	publicPathname     string
	publicName         string
	driver             interfaces.MediumDriver
	readable           bool
	writable           bool
	creationTime       int64
	accessTime         int64
	modificationTime   int64
	handle             *os.File
	mutex              sync.Mutex
	size               int64
	preReadCallbacks   []interfaces.PreReadCallback
	postReadCallbacks  []interfaces.PostReadCallback
	preWriteCallbacks  []interfaces.PreWriteCallback
	postWriteCallbacks []interfaces.PostWriteCallback
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

func (mb *MediumBase) Getattr(path string, stat *fuse.Stat_t, fh uint64) (int, error) {
	return mb.driver.Getattr(mb, path, stat, fh)
}

func (mb *MediumBase) Read(path string, buff []byte, ofst int64, fh uint64) (int, error) {
	mb.CallPreReadCallbacks(mb, path, buff, ofst, fh)

	startTime := time.Now().UnixMilli()
	result, err := mb.driver.Read(mb, path, buff, ofst, fh)
	totalTime := time.Now().UnixMilli() - startTime

	mb.CallPostReadCallbacks(mb, path, buff, ofst, fh, result, totalTime)

	return result, err
}

func (mb *MediumBase) Write(path string, buff []byte, ofst int64, fh uint64) (int, error) {
	mb.CallPreWriteCallbacks(mb, path, buff, ofst, fh)

	startTime := time.Now().UnixMilli()
	result, err := mb.driver.Write(mb, path, buff, ofst, fh)
	totalTime := time.Now().UnixMilli() - startTime

	mb.CallPostWriteCallbacks(mb, path, buff, ofst, fh, result, totalTime)

	return result, err
}

func (mb *MediumBase) Close() error {
	return mb.driver.CloseMedium(mb)
}

func (mb *MediumBase) GetHandle() (*os.File, error) {
	return mb.handle, nil
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

func (mb *MediumBase) AddPreReadCallback(preReadCallback interfaces.PreReadCallback) {
	mb.preReadCallbacks = append(mb.preReadCallbacks, preReadCallback)
}

func (mb *MediumBase) AddPostReadCallback(postReadCallback interfaces.PostReadCallback) {
	mb.postReadCallbacks = append(mb.postReadCallbacks, postReadCallback)
}

func (mb *MediumBase) AddPreWriteCallback(preWriteCallback interfaces.PreWriteCallback) {
	mb.preWriteCallbacks = append(mb.preWriteCallbacks, preWriteCallback)
}

func (mb *MediumBase) AddPostWriteCallback(postWriteCallback interfaces.PostWriteCallback) {
	mb.postWriteCallbacks = append(mb.postWriteCallbacks, postWriteCallback)
}

func (mb *MediumBase) CallPreReadCallbacks(_medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64) {
	for _, callback := range mb.preReadCallbacks {
		callback(_medium, path, buff, ofst, fh)
	}
}

func (mb *MediumBase) CallPreWriteCallbacks(_medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64) {
	for _, callback := range mb.preWriteCallbacks {
		callback(_medium, path, buff, ofst, fh)
	}
}

func (mb *MediumBase) CallPostReadCallbacks(_medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64, n int, opTimeMs int64) {
	for _, callback := range mb.postReadCallbacks {
		callback(_medium, path, buff, ofst, fh, n, opTimeMs)
	}
}

func (mb *MediumBase) CallPostWriteCallbacks(_medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64, n int, opTimeMs int64) {
	for _, callback := range mb.postWriteCallbacks {
		callback(_medium, path, buff, ofst, fh, n, opTimeMs)
	}
}

func (mb *MediumBase) DevicePathnameToPublicFilename(devicePathname string, extension string) string {
	filename := strings.ReplaceAll(
		devicePathname,
		"/",
		"__")

	return filename + "." + extension
}
