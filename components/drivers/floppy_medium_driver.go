package drivers

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/components/medium"
	"github.com/skazanyNaGlany/go.amipi400/interfaces"
)

const floppyDeviceSectorSize = 512
const floppyDeviceSize = 1474560
const floppyAdfSize = 901120
const floppyAdfExtension = "adf"
const floppyDeviceLastSector = floppyDeviceSize - floppyDeviceSectorSize
const floppyDeviceType = "disk"

var goUtils components.GoUtils

type FloppyMediumDriver struct {
	MediumDriverBase
}

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
