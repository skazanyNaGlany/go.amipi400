package main

import (
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/components/drivers"
	"github.com/skazanyNaGlany/go.amipi400/components/medium"
	"github.com/skazanyNaGlany/go.amipi400/consts"
	"github.com/skazanyNaGlany/go.amipi400/interfaces"
)

const AMIGA_DISK_DEVICES_UNIXNAME = "amiga_disk_devices"
const AMIGA_DISK_DEVICES_VERSION = "0.1"
const SYSTEM_INTERNAL_SD_CARD_NAME = "mmcblk0"
const POOL_DEVICE_NAME = "loop"
const FILE_SYSTEM_MOUNT = "/tmp/amiga_disk_devices"
const CACHED_ADFS = "./cached_adfs"
const FLOPPY_READ_MUTE_SECS = 4
const FLOPPY_WRITE_MUTE_SECS = 4
const FLOPPY_WRITE_BLINK_POWER_SECS = 4
const RUNNERS_VERBOSE_MODE = true
const RUNNERS_DEBUG_MODE = true
const DRIVERS_VERBOSE_MODE = true
const DRIVERS_DEBUG_MODE = true
const FORCE_INSERT_KEY = "L_SHIFT"

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
var cachedAdfsDir = ""

func isInternalMedium(name string) bool {
	return strings.HasPrefix(name, SYSTEM_INTERNAL_SD_CARD_NAME)
}

func isPoolMedium(name string) bool {
	return strings.HasPrefix(name, POOL_DEVICE_NAME)
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

	forceInsert := keyboardControl.IsKeyPressed(FORCE_INSERT_KEY)

	// perform only one special action at a time
	keyboardControl.ClearPressedKeys()

	// try FloppyMediumDriver
	floppyDriver := drivers.FloppyMediumDriver{
		CachedAdfsDirectory:   cachedAdfsDir,
		CachedAdfsHeaderMagic: strings.ToUpper(consts.AMIPI400_APP_UNIXNAME),
	}

	floppyDriver.SetVerboseMode(DRIVERS_VERBOSE_MODE)
	floppyDriver.SetDebugMode(DRIVERS_DEBUG_MODE)

	medium, err := floppyDriver.Probe(
		FILE_SYSTEM_MOUNT,
		name,
		size,
		_type,
		mountpoint,
		label,
		path,
		fsType,
		ptType,
		readOnly,
		forceInsert)

	if err != nil {
		return nil, err
	}

	if medium != nil {
		return medium, nil
	}

	// try CDMediumDriver
	cdDriver := drivers.CDMediumDriver{}

	cdDriver.SetVerboseMode(DRIVERS_VERBOSE_MODE)
	cdDriver.SetDebugMode(DRIVERS_DEBUG_MODE)

	medium, err = cdDriver.Probe(
		FILE_SYSTEM_MOUNT,
		name,
		size,
		_type,
		mountpoint,
		label,
		path,
		fsType,
		ptType,
		readOnly,
		forceInsert)

	if err != nil {
		return nil, err
	}

	if medium != nil {
		return medium, nil
	}

	// try HardDiskMediumDriver
	hdDriver := drivers.HardDiskMediumDriver{}

	hdDriver.SetVerboseMode(DRIVERS_VERBOSE_MODE)
	hdDriver.SetDebugMode(DRIVERS_DEBUG_MODE)

	medium, err = hdDriver.Probe(
		FILE_SYSTEM_MOUNT,
		name,
		size,
		_type,
		mountpoint,
		label,
		path,
		fsType,
		ptType,
		readOnly,
		forceInsert)

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
			volumeControl.MuteForSecs(FLOPPY_READ_MUTE_SECS)
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
		volumeControl.MuteForSecs(FLOPPY_WRITE_MUTE_SECS)
	}
}

func onMediumWrite(_medium interfaces.Medium) {
	ledControl.BlinkPowerLEDSecs(FLOPPY_WRITE_BLINK_POWER_SECS)
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

func initCreateDirs(exeDir string) {
	if err := os.MkdirAll(FILE_SYSTEM_MOUNT, 0777); err != nil {
		log.Fatalln(err)
	}

	cachedAdfsDir = path.Join(exeDir, CACHED_ADFS)

	if err := os.MkdirAll(cachedAdfsDir, 0777); err != nil {
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

	initCreateDirs(exeDir)

	log.Printf("%v v%v\n", AMIGA_DISK_DEVICES_UNIXNAME, AMIGA_DISK_DEVICES_VERSION)
	log.Printf("Executable directory %v\n", exeDir)
	log.Printf("Log filename %v\n", logFilename)
	log.Println("File system directory " + FILE_SYSTEM_MOUNT)
	log.Println("Cached ADFs directory " + cachedAdfsDir)

	fileSystem.SetMountDir(FILE_SYSTEM_MOUNT)

	discoverDriveDevices()
	printFloppyDevices()
	printCDROMDevices()

	blockDevices.SetVerboseMode(RUNNERS_VERBOSE_MODE)
	blockDevices.SetDebugMode(RUNNERS_DEBUG_MODE)
	fileSystem.SetVerboseMode(RUNNERS_VERBOSE_MODE)
	fileSystem.SetDebugMode(RUNNERS_DEBUG_MODE)
	volumeControl.SetVerboseMode(RUNNERS_VERBOSE_MODE)
	volumeControl.SetDebugMode(RUNNERS_DEBUG_MODE)
	ledControl.SetVerboseMode(RUNNERS_VERBOSE_MODE)
	ledControl.SetDebugMode(RUNNERS_DEBUG_MODE)
	asyncFileOps.SetVerboseMode(RUNNERS_VERBOSE_MODE)
	asyncFileOps.SetDebugMode(RUNNERS_DEBUG_MODE)
	keyboardControl.SetVerboseMode(RUNNERS_VERBOSE_MODE)
	keyboardControl.SetDebugMode(RUNNERS_DEBUG_MODE)

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
