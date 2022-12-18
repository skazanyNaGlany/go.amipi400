package main

import (
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/components/drivers"
	"github.com/skazanyNaGlany/go.amipi400/interfaces"
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

// Check if the medium is known to the system
func isKnownMedium(name, mountpoint, label, path, fsType, ptType string) bool {
	return mountpoint != "" || label != "" || fsType != "" || ptType != ""
}

func printBlockDevice(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	log.Println("\tName:          " + name)
	log.Println("\tSize:          " + strconv.FormatUint(size, 10))
	log.Println("\tType:          " + _type)
	log.Println("\tMountpoint:    " + mountpoint)
	log.Println("\tLabel:         " + label)
	log.Println("\tPathname:      " + path)
	log.Println("\tFsType:        " + fsType)
	log.Println("\tPtType:        " + ptType)
	log.Println("\tRead-only:     " + strconv.FormatBool(readOnly))
}

func ProbeMediumForDriver(pathname string, size uint64, _type string, readOnly bool) (interfaces.Medium, error) {
	floppyDriver := drivers.FloppyMediumDriver{}

	medium, err := floppyDriver.Probe(pathname, size, _type, readOnly)

	if err != nil {
		return nil, err
	}

	if medium != nil {
		return medium, nil
	}

	return nil, nil
}

func attachedBlockDevice(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	if isInternalMedium(name) {
		return
	}

	if isKnownMedium(name, mountpoint, label, path, fsType, ptType) {
		return
	}

	log.Println("Found new block device", path)
	printBlockDevice(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)

	medium, err := ProbeMediumForDriver(path, size, _type, readOnly)

	if err != nil {
		log.Println(err)
		return
	}

	if medium == nil {
		log.Println("Unable to find driver for medium", path)
		return
	}

	log.Printf("Medium %v will be handled by %T driver\n", path, medium.GetDriver())

	fileSystem.AddMedium(medium)
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
	printBlockDevice(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
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
