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
const fileSystemMount = "/tmp/amiga_disk_devices"

var goUtils components.GoUtils
var blockDevices components.BlockDevices
var fileSystem components.ADDFileSystem
var runnersBlocker components.RunnersBlocker
var driveDevicesDiscovery components.DriveDevicesDiscovery

func isInternalMedium(name string) bool {
	return strings.HasPrefix(name, systemInternalSdCardName)
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

func ProbeMediumForDriver(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) (interfaces.Medium, error) {

	// try FloppyMediumDriver
	floppyDriver := drivers.FloppyMediumDriver{}

	medium, err := floppyDriver.Probe(
		fileSystemMount,
		name,
		size,
		_type,
		mountpoint,
		label,
		path,
		fsType,
		ptType,
		readOnly)

	if err != nil {
		return nil, err
	}

	if medium != nil {
		return medium, nil
	}

	// try CDMediumDriver
	cdDriver := drivers.CDMediumDriver{}

	medium, err = cdDriver.Probe(
		fileSystemMount,
		name,
		size,
		_type,
		mountpoint,
		label,
		path,
		fsType,
		ptType,
		readOnly)

	if err != nil {
		return nil, err
	}

	if medium != nil {
		return medium, nil
	}

	// try HardDiskMediumDriver
	hdDriver := drivers.HardDiskMediumDriver{}

	medium, err = hdDriver.Probe(
		fileSystemMount,
		name,
		size,
		_type,
		mountpoint,
		label,
		path,
		fsType,
		ptType,
		readOnly)

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

	log.Println("Found new block device", path)

	printBlockDevice(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)

	medium, err := ProbeMediumForDriver(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)

	if err != nil {
		log.Println(err)
		return
	}

	if medium == nil {
		log.Println("Unable to find driver for medium", path)
		return
	}

	log.Printf("Medium %v will be handled by %T driver (as %v)\n", path, medium.GetDriver(), medium.GetPublicPathname())

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

	log.Println("Removed block device", path)

	printBlockDevice(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)

	if _, err := fileSystem.RemoveMediumByDevicePathname(path); err != nil {
		log.Println("Unable to close medium:", path, ":", err)
	}
}

func preReadCallback(medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64) {
	log.Println(
		"preReadCallback ",
		path,
		len(buff),
		ofst,
		fh)
}

func postReadCallback(medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64, n int, opTime int64) {
	log.Println(
		"postReadCallback",
		path,
		len(buff),
		ofst,
		fh,
		n,
		opTime)
}

func preWriteCallback(medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64) {
	log.Println(
		"preWriteCallback ",
		path,
		len(buff),
		ofst,
		fh)
}

func postWriteCallback(medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64, n int, opTime int64) {
	log.Println(
		"postWriteCallback",
		path,
		len(buff),
		ofst,
		fh,
		n,
		opTime)
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
	fileSystem.AddPreReadCallback(preReadCallback)
	fileSystem.AddPostReadCallback(postReadCallback)
	fileSystem.AddPreWriteCallback(preWriteCallback)
	fileSystem.AddPostWriteCallback(postWriteCallback)

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
