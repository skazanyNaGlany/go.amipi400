package drivers

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/components/medium"
	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"github.com/winfsp/cgofuse/fuse"
)

const floppyDeviceSectorSize = 512
const floppyDeviceSize = 1474560
const floppyAdfSize = 901120
const floppyAdfExtension = "adf"
const floppyDeviceLastSector = floppyDeviceSize - floppyDeviceSectorSize
const floppyDeviceType = "disk"

var goUtils components.GoUtils

// Implements MediumDriver
type FloppyMediumDriver struct{}

func (fmd *FloppyMediumDriver) Probe(
	basePath, pathname string,
	size uint64,
	_type string,
	readOnly bool) (interfaces.Medium, error) {
	if size != floppyDeviceSize {
		return nil, nil
	}

	if _type != floppyDeviceType {
		return nil, nil
	}

	filename := strings.ReplaceAll(
		pathname,
		"/",
		"__")
	filename = filename + "." + floppyAdfExtension

	medium := medium.FloppyMedium{}

	medium.SetDriver(fmd)
	medium.SetDevicePathname(pathname)
	medium.SetPublicPathname(
		filepath.Join(basePath, filename),
	)

	// in Linux all devices are readable by default
	medium.SetReadable(true)
	medium.SetWritable(!readOnly)

	now := time.Now().Unix()

	medium.SetCreateTime(now)
	medium.SetAccessTime(now)
	medium.SetModificationTime(now)

	return &medium, nil
}

func (fmd *FloppyMediumDriver) Getattr(medium interfaces.Medium, path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	// TODO move to MediumDriverBase
	creationTime := medium.GetCreateTime()
	accessTime := medium.GetAccessTime()
	modificationTime := medium.GetModificationTime()

	isReadable := medium.IsReadable()
	isWritable := medium.IsWritable()

	mask := fmd.generatePermIntMask(
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

func (fmd *FloppyMediumDriver) Read(medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64) (n int) {
	return 0
}

func (fmd *FloppyMediumDriver) Write(medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64) int {
	return 0
}

func (fmd *FloppyMediumDriver) generatePermIntMask(
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
