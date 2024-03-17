package drivers

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/interfaces"
	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
	"github.com/winfsp/cgofuse/fuse"
)

// Implements MediumDriver
type MediumDriverBase struct {
	verboseMode bool
	debugMode   bool
}

func (mdb *MediumDriverBase) Open(
	medium interfaces.Medium,
	path string,
	flags int,
) (errc int, fh uint64) {
	is_writing := 0

	is_writing += flags & os.O_WRONLY
	is_writing += flags & os.O_RDWR
	is_writing += flags & os.O_APPEND
	is_writing += flags & os.O_CREATE
	is_writing += flags & os.O_TRUNC

	if is_writing > 0 && !medium.IsWritable() {
		return -fuse.EPERM, 0
	}

	return 0, 0
}

func (mdb *MediumDriverBase) Getattr(
	medium interfaces.Medium,
	path string,
	stat *fuse.Stat_t,
	fh uint64,
) (int, error) {
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

	return 0, nil
}

func (mdb *MediumDriverBase) Read(
	medium interfaces.Medium,
	path string,
	buff []byte,
	ofst int64,
	fh uint64,
) (int, error) {
	mutex := medium.GetMutex()

	mutex.Lock()
	defer mutex.Unlock()

	handle, err := mdb.OpenMediumHandle(medium)

	if err != nil {
		return 0, err
	}

	medium.SetAccessTime(
		time.Now().Unix())

	lenBuff := int64(len(buff))
	toReadSize := lenBuff
	fileSize := medium.GetSize()

	if ofst+toReadSize > fileSize {
		toReadSize = fileSize - ofst
	}

	data, n, err := utils.FileUtilsInstance.FileReadBytes(
		"",
		ofst,
		toReadSize,
		0,
		0,
		handle)

	if err != nil {
		return 0, err
	}

	copy(buff, data)

	return n, nil
}

func (mdb *MediumDriverBase) Write(
	medium interfaces.Medium,
	path string,
	buff []byte,
	ofst int64,
	fh uint64,
) (int, error) {
	mutex := medium.GetMutex()

	mutex.Lock()
	defer mutex.Unlock()

	if !medium.IsWritable() {
		return 0, errors.New("medium is not writable")
	}

	handle, err := mdb.OpenMediumHandle(medium)

	if err != nil {
		return 0, err
	}

	medium.SetModificationTime(
		time.Now().Unix())

	fileSize := medium.GetSize()
	lenBuff := len(buff)

	if ofst+int64(lenBuff) > fileSize || ofst >= fileSize {
		return 0, errors.New("write outside the medium data")
	}

	n, err := utils.FileUtilsInstance.FileWriteBytes("", ofst, buff, 0, 0, handle)

	if err != nil {
		return 0, err
	}

	return n, nil
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

	goUtils := utils.GoUtilsInstance

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

func (mdb *MediumDriverBase) OpenMediumHandle(
	medium interfaces.Medium,
	readAhead ...int,
) (*os.File, error) {
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

	_readAhead := shared.DEFAULT_READ_AHEAD

	if len(readAhead) == 1 {
		_readAhead = readAhead[0]
	}

	// set read-a-head value for device or file handle
	// for block-device and the file-system
	if err = utils.UnixUtilsInstance.SetDeviceReadAHead(handle, _readAhead); err != nil {
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
func (mdb *MediumDriverBase) isKnownMedium(
	name, mountpoint, label, path, fsType, ptType string,
) bool {
	return mountpoint != "" || label != "" || fsType != "" || ptType != ""
}

func (mdb *MediumDriverBase) SetVerboseMode(verboseMode bool) {
	mdb.verboseMode = verboseMode
}

func (mdb *MediumDriverBase) SetDebugMode(debugMode bool) {
	mdb.debugMode = debugMode
}

func (mdb *MediumDriverBase) GetVerboseMode() bool {
	return mdb.verboseMode
}

func (mdb *MediumDriverBase) GetDebugMode() bool {
	return mdb.debugMode
}
