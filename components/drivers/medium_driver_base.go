package drivers

import (
	"strconv"

	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"github.com/winfsp/cgofuse/fuse"
)

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
	stat.Size = floppyAdfSize
	stat.Ctim = fuse.Timespec{Sec: creationTime}
	stat.Atim = fuse.Timespec{Sec: accessTime}
	stat.Mtim = fuse.Timespec{Sec: modificationTime}

	return 0
}

func (mdb *MediumDriverBase) Read(medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64) (n int) {
	return 0
}

func (mdb *MediumDriverBase) Write(medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64) int {
	return 0
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
