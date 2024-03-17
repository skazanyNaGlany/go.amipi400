package drivers

import (
	"path/filepath"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/components/medium"
	"github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/interfaces"
	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
)

type CDMediumDriver struct {
	MediumDriverBase
}

func (cdmd *CDMediumDriver) Probe(
	basePath, name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly, force, formatted bool) (interfaces.Medium, error) {
	if !cdmd.isKnownMedium(name, mountpoint, label, path, fsType, ptType) {
		// we are supporting only data CDs, and data CDs will have
		// valid mountpoint (already mounted by the system), label
		// fsType or ptType
		return nil, nil
	}

	if _type != shared.CD_DEVICE_TYPE {
		return nil, nil
	}

	// last chance, try to read at least 2048 bytes (CD sector size) from the medium
	// non-inserted medium or audio CDs will report just error
	// here, or count of the readed bytes will be less than 2048
	data, n, err := utils.FileUtilsInstance.FileReadBytes(
		path,
		0,
		shared.CD_DEVICE_SECTOR_SIZE,
		0,
		0,
		nil,
	)

	if len(data) < shared.CD_DEVICE_SECTOR_SIZE || n < shared.CD_DEVICE_SECTOR_SIZE ||
		err != nil {
		return nil, nil
	}

	// ok should be data CD, perhaps we will need to check for FsType == iso9660 also

	medium := medium.MediumBase{}

	filename := medium.DevicePathnameToPublicFilename(path, shared.CD_ISO_EXTENSION)

	medium.SetDriver(cdmd)
	medium.SetDevicePathname(path)
	medium.SetPublicPathname(
		filepath.Join(basePath, filename),
	)
	medium.SetSize(int64(size))

	// in Linux all devices are readable by default
	medium.SetReadable(true)

	// data CDs will be always read-only
	medium.SetWritable(false)

	now := time.Now().Unix()

	medium.SetCreateTime(now)
	medium.SetAccessTime(now)
	medium.SetModificationTime(now)

	// and that is all, rest of the logic will be handled
	// by MediumDriverBase

	return &medium, nil
}
