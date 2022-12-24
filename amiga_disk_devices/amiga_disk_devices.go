package main

import (
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/components/drivers"
	"github.com/skazanyNaGlany/go.amipi400/components/medium"
	"github.com/skazanyNaGlany/go.amipi400/interfaces"
)

const AppUnixname = "amiga_disk_devices"
const AppVersion = "0.1"
const systemInternalSdCardName = "mmcblk0"
const poolDeviceName = "loop"
const fileSystemMount = "/tmp/amiga_disk_devices"
const floppyReadMuteSecs = 4
const floppyWriteMuteSecs = 4
const floppyWriteBlinkPowerSecs = 4
const runnersVerboseMode = true
const runnersDebugMode = true

var goUtils components.GoUtils
var unixUtils components.UnixUtils
var blockDevices components.BlockDevices
var fileSystem components.ADDFileSystem
var runnersBlocker components.RunnersBlocker
var driveDevicesDiscovery components.DriveDevicesDiscovery
var volumeControl components.VolumeControl
var ledControl components.LEDControl
var asyncFileOps components.AsyncFileOps
var keyboardControl components.KeyboardControl

func isInternalMedium(name string) bool {
	return strings.HasPrefix(name, systemInternalSdCardName)
}

func isPoolMedium(name string) bool {
	return strings.HasPrefix(name, poolDeviceName)
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

	if isPoolMedium(name) {
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

	medium.AddPreReadCallback(preReadCallback)
	medium.AddPostReadCallback(postReadCallback)
	medium.AddPreWriteCallback(preWriteCallback)
	medium.AddPostWriteCallback(postWriteCallback)

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

	if isPoolMedium(name) {
		return
	}

	log.Println("Removed block device", path)

	printBlockDevice(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)

	if _, err := fileSystem.RemoveMediumByDevicePathname(path); err != nil {
		log.Println("Unable to close medium:", path, ":", err)
	}
}

func onFloppyRead(_medium interfaces.Medium, ofst int64) {
	floppyMedium, isFloppy := _medium.(*medium.FloppyMedium)

	if isFloppy {
		if !floppyMedium.IsFullyCached() {
			// reading from non-cached floppy medium
			volumeControl.MuteForSecs(floppyReadMuteSecs)
		} else {
			// reading from cached floppy medium
			asyncFileOps.FileReadBytesDirect(
				floppyMedium.GetDevicePathname(),
				ofst,
				0,
				0,
				nil,
				4,
				nil)
		}
	}
}

func onFloppyWrite(_medium interfaces.Medium) {
	_, isFloppy := _medium.(*medium.FloppyMedium)

	if isFloppy {
		volumeControl.MuteForSecs(floppyWriteMuteSecs)
	}
}

func onMediumWrite(_medium interfaces.Medium) {
	ledControl.BlinkPowerLEDSecs(floppyWriteBlinkPowerSecs)
}

func preReadCallback(_medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64) {
	// do not put too much logic here, since it will slow down MediumBase.Read
	// or FloppyMediumDriver.Read, same for callbacks related for Write
	onFloppyRead(_medium, ofst)
}

func postReadCallback(_medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64, n int, opTimeMs int64) {
	onFloppyRead(_medium, ofst)
}

func preWriteCallback(_medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64) {
	onFloppyWrite(_medium)
	onMediumWrite(_medium)
}

func postWriteCallback(_medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64, n int, opTimeMs int64) {
	onFloppyWrite(_medium)
	onMediumWrite(_medium)
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

func checkForRoot() {
	if !unixUtils.IsRoot() {
		log.Fatalln("Must be run as root.")
	}
}

func main() {
	checkPlatform()
	checkForRoot()
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

	blockDevices.SetVerboseMode(runnersVerboseMode)
	blockDevices.SetDebugMode(runnersDebugMode)
	fileSystem.SetVerboseMode(runnersVerboseMode)
	fileSystem.SetDebugMode(runnersDebugMode)
	volumeControl.SetVerboseMode(runnersVerboseMode)
	volumeControl.SetDebugMode(runnersDebugMode)
	ledControl.SetVerboseMode(runnersVerboseMode)
	ledControl.SetDebugMode(runnersDebugMode)
	asyncFileOps.SetVerboseMode(runnersVerboseMode)
	asyncFileOps.SetDebugMode(runnersDebugMode)
	keyboardControl.SetVerboseMode(runnersVerboseMode)
	keyboardControl.SetDebugMode(runnersDebugMode)

	blockDevices.AddAttachedHandler(attachedBlockDevice)
	blockDevices.AddDetachedHandler(detachedBlockDevice)

	fileSystem.Start(&fileSystem)
	blockDevices.Start(&blockDevices)
	volumeControl.Start(&volumeControl)
	ledControl.Start(&ledControl)
	asyncFileOps.Start(&asyncFileOps)
	keyboardControl.Start(&keyboardControl)

	defer fileSystem.Stop(&fileSystem)
	defer blockDevices.Stop(&blockDevices)
	defer volumeControl.Stop(&volumeControl)
	defer ledControl.Stop(&ledControl)
	defer asyncFileOps.Stop(&asyncFileOps)
	defer keyboardControl.Stop(&keyboardControl)

	runnersBlocker.AddRunner(&blockDevices)
	runnersBlocker.AddRunner(&fileSystem)
	runnersBlocker.AddRunner(&volumeControl)
	runnersBlocker.AddRunner(&ledControl)
	runnersBlocker.AddRunner(&asyncFileOps)
	runnersBlocker.AddRunner(&keyboardControl)
	runnersBlocker.BlockUntilRunning()
}
