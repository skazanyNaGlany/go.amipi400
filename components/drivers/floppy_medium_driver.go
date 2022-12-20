package drivers

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/components/medium"
	"github.com/skazanyNaGlany/go.amipi400/interfaces"
)

const floppyDeviceSize = 1474560
const floppyAdfSize = 901120
const floppyAdfExtension = "adf"
const floppyDeviceType = "disk"

// const floppyDeviceSectorSize = 512
// const floppyDeviceLastSector = floppyDeviceSize - floppyDeviceSectorSize

var goUtils components.GoUtils

type FloppyMediumDriver struct {
	MediumDriverBase
}

func (fmd *FloppyMediumDriver) Probe(
	basePath, name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) (interfaces.Medium, error) {
	if fmd.isKnownMedium(name, mountpoint, label, path, fsType, ptType) {
		return nil, nil
	}

	if size != floppyDeviceSize {
		return nil, nil
	}

	if _type != floppyDeviceType {
		return nil, nil
	}

	filename := strings.ReplaceAll(
		path,
		"/",
		"__")
	filename = filename + "." + floppyAdfExtension

	medium := medium.FloppyMedium{}

	medium.SetDriver(fmd)
	medium.SetDevicePathname(path)
	medium.SetPublicPathname(
		filepath.Join(basePath, filename),
	)
	medium.SetSize(floppyAdfSize)

	// in Linux all devices are readable by default
	medium.SetReadable(true)
	medium.SetWritable(!readOnly)

	now := time.Now().Unix()

	medium.SetCreateTime(now)
	medium.SetAccessTime(now)
	medium.SetModificationTime(now)

	return &medium, nil
}
