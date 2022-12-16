package main

import (
	"log"
	"strings"

	"github.com/skazanyNaGlany/go.amipi400/components"
)

const AppUnixname = "amiga_disk_devices"
const AppVersion = "0.1"
const systemInternalSdCardName = "mmcblk0"

var goUtils components.GoUtils
var blockDevices components.BlockDevices
var fileSystem components.ADDFileSystem
var runnersBlocker components.RunnersBlocker

func isInternalMedium(name string) bool {
	return strings.HasPrefix(name, systemInternalSdCardName)
}

func attachedBlockDevice(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	if isInternalMedium(name) {
		return
	}

	log.Println("Found new block device", name)
}

func detachedBlockDevice(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	if isInternalMedium(name) {
		return
	}

	log.Println("Removed block device", name)
}

func main() {
	exeDir := goUtils.CwdToExe()
	logFilename := goUtils.DuplicateLog()

	log.Printf("%v v%v\n", AppUnixname, AppVersion)
	log.Printf("Executable directory %v\n", exeDir)
	log.Printf("Log filename %v\n", logFilename)

	blockDevices.AddAttachedHandler(attachedBlockDevice)
	blockDevices.AddDetachedHandler(detachedBlockDevice)

	fileSystem.Start()
	defer fileSystem.Stop()

	blockDevices.Start()
	defer blockDevices.Stop()

	runnersBlocker.AddRunner(&blockDevices)
	runnersBlocker.AddRunner(&fileSystem)
	runnersBlocker.BlockUntilRunning()
}
