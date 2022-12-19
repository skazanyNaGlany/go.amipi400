package drivers

import (
	"strconv"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"github.com/winfsp/cgofuse/fuse"
)

var fileUtils components.FileUtils

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

	return 0
}

func (mdb *MediumDriverBase) Read(medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64) (n int) {
	mutex := medium.GetMutex()

	mutex.Lock()
	defer mutex.Unlock()

	handle, err := medium.GetHandle()

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

	data, n, err := fileUtils.FileReadBytes(
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

	handle, err := medium.GetHandle()

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

	n, err := fileUtils.FileWriteBytes("", ofst, buff, 0, 0, handle)

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
