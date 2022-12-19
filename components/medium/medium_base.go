package medium

import (
	"path/filepath"

	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"github.com/winfsp/cgofuse/fuse"
)

// Implements Medium
type MediumBase struct {
	devicePathname string
	publicPathname string
	publicName     string
	driver         interfaces.MediumDriver
	fullyCached    bool
}

func (mb *MediumBase) GetPublicName() string {
	return mb.publicName
}

func (mb *MediumBase) IsFullyCached() bool {
	return mb.fullyCached
}

func (mb *MediumBase) SetFullyCached(fullyCached bool) {
	mb.fullyCached = fullyCached
}

func (mb *MediumBase) SetDevicePathname(devicePathname string) {
	mb.devicePathname = devicePathname
}

func (mb *MediumBase) SetPublicPathname(publicPathname string) {
	mb.publicPathname = publicPathname
	mb.publicName = filepath.Base(publicPathname)
}

func (mb *MediumBase) SetDriver(driver interfaces.MediumDriver) {
	mb.driver = driver
}

func (mb *MediumBase) GetDevicePathname() string {
	return mb.devicePathname
}

func (mb *MediumBase) GetPublicPathname() string {
	return mb.publicPathname
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
