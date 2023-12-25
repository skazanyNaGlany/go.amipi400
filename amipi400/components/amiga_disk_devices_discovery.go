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
	isIdle                          bool
	idleCallback                    interfaces.IdleCallback
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
	affected := 0
	oldIsIdle := addd.isIdle

	for _, pathname := range oldFiles {
		if !funk.ContainsString(files, pathname) {
			affected++
			addd.isIdle = false

			addd.detachedAmigaDiskDeviceCallback(pathname)
		}
	}

	for _, pathname := range files {
		if !funk.ContainsString(oldFiles, pathname) {
			affected++
			addd.isIdle = false

			addd.attachedAmigaDiskDeviceCallback(pathname)
		}
	}

	addd.isIdle = affected == 0

	if addd.isIdle && !oldIsIdle {
		if addd.idleCallback != nil {
			addd.idleCallback(addd)
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

func (addd *AmigaDiskDevicesDiscovery) SetIdleCallback(callback interfaces.IdleCallback) {
	addd.idleCallback = callback
}

func (addd *AmigaDiskDevicesDiscovery) IsIdle() bool {
	return addd.isIdle
}
