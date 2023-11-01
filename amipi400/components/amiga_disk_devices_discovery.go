package components

import (
	"time"

	"github.com/skazanyNaGlany/go.amipi400/amipi400/interfaces"
	"github.com/skazanyNaGlany/go.amipi400/shared/components"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
	"github.com/thoas/go-funk"
)

type AmigaDiskDevicesDiscovery struct {
	components.RunnerBase
	attachedAmigaDiskDeviceCallback interfaces.AttachedAmigaDiskDeviceCallback
	detachedAmigaDiskDeviceCallback interfaces.DetachedAmigaDiskDeviceCallback
	mountpoint                      string
	currentFiles                    []string
}

func (addd *AmigaDiskDevicesDiscovery) loop() {
	var oldFiles []string

	for addd.IsRunning() {
		time.Sleep(time.Millisecond * 10)

		addd.currentFiles = utils.FileUtilsInstance.GetDirFiles(addd.mountpoint, false)

		addd.callCallbacks(addd.currentFiles, oldFiles)

		oldFiles = addd.currentFiles
	}

	addd.SetRunning(false)
}

func (addd *AmigaDiskDevicesDiscovery) HasFile(pathname string) bool {
	return funk.ContainsString(addd.currentFiles, pathname)
}

func (addd *AmigaDiskDevicesDiscovery) callCallbacks(files []string, oldFiles []string) {
	for _, pathname := range oldFiles {
		if !funk.ContainsString(files, pathname) {
			addd.detachedAmigaDiskDeviceCallback(pathname)
		}
	}

	for _, pathname := range files {
		if !funk.ContainsString(oldFiles, pathname) {
			addd.attachedAmigaDiskDeviceCallback(pathname)
		}
	}
}

func (addd *AmigaDiskDevicesDiscovery) SetAttachedAmigaDiskDeviceCallback(callback interfaces.AttachedAmigaDiskDeviceCallback) {
	addd.attachedAmigaDiskDeviceCallback = callback
}

func (addd *AmigaDiskDevicesDiscovery) SetDetachedAmigaDiskDeviceCallback(callback interfaces.DetachedAmigaDiskDeviceCallback) {
	addd.detachedAmigaDiskDeviceCallback = callback
}

func (addd *AmigaDiskDevicesDiscovery) Run() {
	addd.currentFiles = make([]string, 0)

	addd.loop()
}

func (addd *AmigaDiskDevicesDiscovery) SetMountpoint(mountpoint string) {
	addd.mountpoint = mountpoint
}
