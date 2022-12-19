package drivers

import (
	"path/filepath"
	"strings"

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

	return &medium, nil
}

func (fmd *FloppyMediumDriver) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	return 0
}

func (fmd *FloppyMediumDriver) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	return 0
}

func (fmd *FloppyMediumDriver) Write(path string, buff []byte, ofst int64, fh uint64) int {
	return 0
}
