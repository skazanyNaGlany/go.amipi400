package main

import (
	"log"
	"os"
	"runtime"
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
var driveDevicesDiscovery components.DriveDevicesDiscovery

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
		log.Fatalln(err)
	}
}

func discoverDriveDevices() {
	log.Println("Getting information about physicall drives")

	if err := driveDevicesDiscovery.Refresh(); err != nil {
		log.Println("Perhaps you need to run this program as root")
		log.Fatalln(err)
	}
}

func printFloppyDevices() {
	floppies := driveDevicesDiscovery.GetFloppies()

	if len(floppies) > 0 {
		log.Println("Physicall floppy drives:")
	}

	for _, devicePathname := range floppies {
		log.Println("\t" + devicePathname)
	}
}

func printCDROMDevices() {
	cdroms := driveDevicesDiscovery.GetCDROMs()

	if len(cdroms) > 0 {
		log.Println("Physicall CDROM drives:")
	}

	for _, devicePathname := range cdroms {
		log.Println("\t" + devicePathname)
	}
}

func DuplicateLog(exeDir string) string {
	logFilePathname, err := goUtils.DuplicateLog(exeDir)

	if err != nil {
		log.Fatalln(err)
	}

	return logFilePathname
}

func CwdToExeOrScript() string {
	exeDir, err := goUtils.CwdToExeOrScript()

	if err != nil {
		log.Fatalln(err)
	}

	return exeDir
}

func checkPlatform() {
	if runtime.GOOS != "linux" {
		log.Fatalln("This app can be used only on Linux.")
	}
}

func main() {
	checkPlatform()
	exeDir := CwdToExeOrScript()
	logFilename := DuplicateLog(exeDir)

	log.Printf("%v v%v\n", AppUnixname, AppVersion)
	log.Printf("Executable directory %v\n", exeDir)
	log.Printf("Log filename %v\n", logFilename)
	log.Println("File system directory " + fileSystemMount)

	createFsDir()
	fileSystem.SetMountDir(fileSystemMount)

	discoverDriveDevices()
	printFloppyDevices()
	printCDROMDevices()

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
