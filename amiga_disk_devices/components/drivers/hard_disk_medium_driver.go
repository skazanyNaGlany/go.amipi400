package drivers

import (
	"path/filepath"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/components/medium"
	"github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/interfaces"
	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
)

type HardDiskMediumDriver struct {
	MediumDriverBase
}

func (hdmd *HardDiskMediumDriver) Probe(
	basePath, name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly, force, formatted bool) (interfaces.Medium, error) {
	if !force {
		if hdmd.isKnownMedium(name, mountpoint, label, path, fsType, ptType) {
			return nil, nil
		}
	}

	if size < shared.HD_DEVICE_MIN_SIZE {
		return nil, nil
	}

	if _type != shared.HD_DEVICE_TYPE {
		return nil, nil
	}

	hasDosHeader, err := hdmd.hasDOSheader(path)

	if err != nil {
		return nil, err
	}

	if !hasDosHeader {
		return nil, nil
	}

	medium := medium.MediumBase{}

	filename := medium.DevicePathnameToPublicFilename(path, shared.HD_HDF_EXTENSION)

	medium.SetDriver(hdmd)
	medium.SetDevicePathname(path)
	medium.SetPublicPathname(
		filepath.Join(basePath, filename),
	)
	medium.SetSize(int64(size))

	// in Linux all devices are readable by default
	medium.SetReadable(true)
	medium.SetWritable(!readOnly)

	now := time.Now().Unix()

	medium.SetCreateTime(now)
	medium.SetAccessTime(now)
	medium.SetModificationTime(now)

	// and that is all, rest of the logic will be handled
	// by MediumDriverBase

	return &medium, nil
}

func (hdmd *HardDiskMediumDriver) hasDOSheader(path string) (bool, error) {
	data, n, err := utils.FileUtilsInstance.FileReadBytes(path, 0, shared.HD_DEVICE_SECTOR_SIZE, 0, 0, nil)

	if err != nil {
		return false, err
	}

	if len(data) < shared.HD_DEVICE_SECTOR_SIZE || n < shared.HD_DEVICE_SECTOR_SIZE {
		return false, nil
	}

	// check for RDSK
	if data[0] == 'R' &&
		data[1] == 'D' &&
		data[2] == 'S' &&
		data[3] == 'K' {
		return true, nil
	}

	// check for DOS
	if data[0] == 'D' &&
		data[1] == 'O' &&
		data[2] == 'S' {
		return true, nil
	}

	return false, nil
}
