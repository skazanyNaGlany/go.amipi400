package main

import (
	"log"

	"github.com/skazanyNaGlany/go.amipi400/components"
)

const AppUnixname = "amiga_disk_devices"
const AppVersion = "0.1"

var goUtils components.GoUtils
var blockDevices components.BlockDevices
var runnersBlocker components.RunnersBlocker

func attachedBlockDevice(
	name string,
	size int,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) (bool, error) {
	return true, nil
}

func detachedBlockDevice(name string) (bool, error) {
	return true, nil
}

func main() {
	exeDir := goUtils.CwdToExe()
	logFilename := goUtils.DuplicateLog()

	log.Printf("%v v%v\n", AppUnixname, AppVersion)
	log.Printf("Executable directory %v\n", exeDir)
	log.Printf("Log filename %v\n", logFilename)

	blockDevices.AddAttachedHandler(attachedBlockDevice)
	blockDevices.AddDetachedHandler(detachedBlockDevice)

	blockDevices.Start()
	defer blockDevices.Stop()

	runnersBlocker.AddRunner(&blockDevices)
	runnersBlocker.BlockUntilRunning()
}
