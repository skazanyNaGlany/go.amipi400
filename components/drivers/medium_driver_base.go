package drivers

import (
	"os"
	"strconv"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"github.com/winfsp/cgofuse/fuse"
	"golang.org/x/sys/unix"
)

const defaultReadAhead = 256

// Implements MediumDriver
type MediumDriverBase struct{}

func (mdb *MediumDriverBase) Getattr(medium interfaces.Medium, path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	// TODO move to MediumDriverBase
	creationTime := medium.GetCreateTime()
	accessTime := medium.GetAccessTime()
	modificationTime := medium.GetModificationTime()

	isReadable := medium.IsReadable()
	isWritable := medium.IsWritable()

	mask := mdb.generatePermIntMask(
		isReadable, isWritable, false,
		isReadable, isWritable, false,
		isReadable, isWritable, false,
	)

	stat.Mode = fuse.S_IFREG | mask
	stat.Size = medium.GetSize()
	stat.Ctim = fuse.Timespec{Sec: creationTime}
	stat.Atim = fuse.Timespec{Sec: accessTime}
	stat.Mtim = fuse.Timespec{Sec: modificationTime}
	stat.Nlink = 1

	return 0
}

func (mdb *MediumDriverBase) Read(medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64) (n int) {
	mutex := medium.GetMutex()

	mutex.Lock()
	defer mutex.Unlock()

	handle, err := mdb.getMediumHandle(medium)

	if err != nil {
		return -fuse.EIO
	}

	medium.SetAccessTime(
		time.Now().Unix())

	lenBuff := len(buff)
	toReadSize := lenBuff
	fileSize := medium.GetSize()

	if ofst+int64(toReadSize) > int64(fileSize) {
		toReadSize = int(fileSize) - int(ofst)
	}

	data, n, err := components.FileUtilsInstance.FileReadBytes(
		"",
		ofst,
		uint64(toReadSize),
		0,
		0,
		handle)

	if err != nil {
		return -fuse.EIO
	}

	copy(buff, data)

	return n
}

func (mdb *MediumDriverBase) Write(medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64) int {
	mutex := medium.GetMutex()

	mutex.Lock()
	defer mutex.Unlock()

	if !medium.IsWritable() {
		return -fuse.EROFS
	}

	handle, err := mdb.getMediumHandle(medium)

	if err != nil {
		return -fuse.EIO
	}

	medium.SetModificationTime(
		time.Now().Unix())

	fileSize := medium.GetSize()
	lenBuff := len(buff)

	if ofst+int64(lenBuff) > fileSize || ofst >= fileSize {
		return -fuse.ENOSPC
	}

	n, err := components.FileUtilsInstance.FileWriteBytes("", ofst, buff, 0, 0, handle)

	if err != nil {
		return -fuse.EIO
	}

	return n
}

func (mdb *MediumDriverBase) generatePermIntMask(
	userCanRead bool,
	userCanWrite bool,
	userCanExecute bool,
	groupCanRead bool,
	groupCanWrite bool,
	groupCanExecute bool,
	otherCanRead bool,
	otherCanWrite bool,
	otherCanExecute bool,
) uint32 {
	binString := ""

	binString += goUtils.BoolToStrInt(userCanRead)
	binString += goUtils.BoolToStrInt(userCanWrite)
	binString += goUtils.BoolToStrInt(userCanExecute)

	binString += goUtils.BoolToStrInt(groupCanRead)
	binString += goUtils.BoolToStrInt(groupCanWrite)
	binString += goUtils.BoolToStrInt(groupCanExecute)

	binString += goUtils.BoolToStrInt(otherCanRead)
	binString += goUtils.BoolToStrInt(otherCanWrite)
	binString += goUtils.BoolToStrInt(otherCanExecute)

	mask, _ := strconv.ParseUint(binString, 2, 64)

	return uint32(mask)
}

func (mdb *MediumDriverBase) getMediumHandle(medium interfaces.Medium, readAhead ...int) (*os.File, error) {
	handle, err := medium.GetHandle()

	if err != nil {
		return nil, err
	}

	if handle != nil {
		return handle, nil
	}

	isReadable := medium.IsReadable()
	isWritable := medium.IsWritable()

	flag := os.O_SYNC

	if isReadable && isWritable {
		flag |= os.O_RDWR
	} else {
		flag |= os.O_RDONLY
	}

	handle, err = os.OpenFile(
		medium.GetDevicePathname(),
		flag,
		0,
	)

	if err != nil {
		return nil, err
	}

	_readAhead := defaultReadAhead

	if len(readAhead) == 1 {
		_readAhead = readAhead[0]
	}

	// set read-a-head value for block-device
	if err = unix.IoctlSetInt(int(handle.Fd()), unix.BLKRASET, _readAhead); err != nil {
		handle.Close()

		return nil, err
	}

	// set read-a-head value for file-system
	if err = unix.IoctlSetInt(int(handle.Fd()), unix.BLKFRASET, _readAhead); err != nil {
		handle.Close()

		return nil, err
	}

	medium.SetHandle(handle)

	return handle, nil
}

func (mdb *MediumDriverBase) CloseMedium(medium interfaces.Medium) error {
	handle, err := medium.GetHandle()

	if err != nil {
		return err
	}

	if handle == nil {
		// handle not open yet, or already closed
		return nil
	}

	err = handle.Close()

	medium.SetHandle(nil)

	return err
}

// Check if the medium is known to the system
func (mdb *MediumDriverBase) isKnownMedium(name, mountpoint, label, path, fsType, ptType string) bool {
	return mountpoint != "" || label != "" || fsType != "" || ptType != ""
}