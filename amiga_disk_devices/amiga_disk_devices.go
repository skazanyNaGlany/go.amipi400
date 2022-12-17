package main

import (
	"log"
	"os"
	"strings"

	"github.com/skazanyNaGlany/go.amipi400/components"
)

const AppUnixname = "amiga_disk_devices"
const AppVersion = "0.1"
const systemInternalSdCardName = "mmcblk0"
const fileSystemMount = "/tmp/amipi400/amiga_disk_devices"

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

func createFsDir() {
	if err := os.MkdirAll(fileSystemMount, 0777); err != nil {
		log.Panicln(err)
	}
}

func DuplicateLog(exeDir string) string {
	logFilePathname, err := goUtils.DuplicateLog(exeDir)

	if err != nil {
		panic(err)
	}

	return logFilePathname
}

func main() {
	exeDir := goUtils.CwdToExeOrScript()
	logFilename := DuplicateLog(exeDir)

	log.Printf("%v v%v\n", AppUnixname, AppVersion)
	log.Printf("Executable directory %v\n", exeDir)
	log.Printf("Log filename %v\n", logFilename)
	log.Println("File system directory " + fileSystemMount)

	createFsDir()
	fileSystem.SetMountDir(fileSystemMount)

	blockDevices.AddAttachedHandler(attachedBlockDevice)
	blockDevices.AddDetachedHandler(detachedBlockDevice)

	fileSystem.Start()
	blockDevices.Start()

	defer fileSystem.Stop()
	defer blockDevices.Stop()

	runnersBlocker.AddRunner(&blockDevices)
	runnersBlocker.AddRunner(&fileSystem)
	runnersBlocker.BlockUntilRunning()
}
