package components

import (
	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"github.com/winfsp/cgofuse/fuse"
)

// Implements Medium
type MediumBase struct {
	devicePathname string
	publicPathname string
	publicName     string
	driver         interfaces.MediumDriver
}

func (mb *MediumBase) GetDevicePathname() string {
	return mb.devicePathname
}

func (mb *MediumBase) GetPublicPathname() string {
	return mb.publicPathname
}

func (mb *MediumBase) GetPublicName() string {
	return mb.publicName
}

func (mb *MediumBase) GetDriver() interfaces.MediumDriver {
	return mb.driver
}

func (mb *MediumBase) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	return mb.driver.Getattr(path, stat, fh)
}

func (mb *MediumBase) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	return mb.driver.Read(path, buff, ofst, fh)
}

func (mb *MediumBase) Write(path string, buff []byte, ofst int64, fh uint64) int {
	return mb.driver.Write(path, buff, ofst, fh)
}
