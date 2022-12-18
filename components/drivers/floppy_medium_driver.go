package drivers

import (
	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"github.com/winfsp/cgofuse/fuse"
)

// Implements MediumDriver
type FloppyMediumDriver struct{}

func (fmd *FloppyMediumDriver) Probe(pathname string, size uint64, _type string, readOnly bool) (interfaces.Medium, error) {
	return nil, nil
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
