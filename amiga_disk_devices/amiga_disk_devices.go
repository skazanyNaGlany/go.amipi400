package main

import (
	"io/fs"
	"log"
	"os"
	"path"
	"strings"

	components_amiga_disk_devices "github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/components"
	drivers_amiga_disk_devices "github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/components/drivers"
	medium_amiga_disk_devices "github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/components/medium"
	interfaces_amiga_disk_devices "github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/interfaces"
	"github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/components/utils"
	"github.com/skazanyNaGlany/go.amipi400/consts"
	"github.com/thoas/go-funk"
)

var blockDevices components.BlockDevices
var fileSystem components_amiga_disk_devices.ADDFileSystem
var runnersBlocker components.RunnersBlocker
var driveDevicesDiscovery components.DriveDevicesDiscovery
var volumeControl components_amiga_disk_devices.VolumeControl
var ledControl components.LEDControl
var asyncFileOps components.AsyncFileOps
var asyncFileOpsDf0 components.AsyncFileOps
var asyncFileOpsDf1 components.AsyncFileOps
var asyncFileOpsDf2 components.AsyncFileOps
var asyncFileOpsDf3 components.AsyncFileOps
var allKeyboardsControl components.AllKeyboardsControl
var cachedAdfsDir = ""

var floppyDevices []string

func ProbeMediumForDriver(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly, formatted bool) (interfaces_amiga_disk_devices.Medium, error) {

	forceInsert := allKeyboardsControl.IsKeysPressed(consts.FORCE_INSERT_KEYS)

	if forceInsert {
		// perform only one special action at a time
		allKeyboardsControl.ClearPressedKeys()
	}

	// try FloppyMediumDriver
	floppyDriver := drivers_amiga_disk_devices.FloppyMediumDriver{}

	floppyDriver.SetCachedAdfsDirectory(cachedAdfsDir)
	floppyDriver.SetVerboseMode(consts.DRIVERS_VERBOSE_MODE)
	floppyDriver.SetDebugMode(consts.DRIVERS_DEBUG_MODE)
	floppyDriver.SetOutsideAsyncFileWriterCallback(outsideAsyncFileWriterCallback)
	floppyDriver.SetPreCacheADFCallback(preCacheADFCallback)

	medium, err := floppyDriver.Probe(
		consts.FILE_SYSTEM_MOUNT,
		name,
		size,
		_type,
		mountpoint,
		label,
		path,
		fsType,
		ptType,
		readOnly,
		forceInsert,
		formatted)

	if err != nil {
		return nil, err
	}

	if medium != nil {
		return medium, nil
	}

	// try CDMediumDriver
	cdDriver := drivers_amiga_disk_devices.CDMediumDriver{}

	cdDriver.SetVerboseMode(consts.DRIVERS_VERBOSE_MODE)
	cdDriver.SetDebugMode(consts.DRIVERS_DEBUG_MODE)

	medium, err = cdDriver.Probe(
		consts.FILE_SYSTEM_MOUNT,
		name,
		size,
		_type,
		mountpoint,
		label,
		path,
		fsType,
		ptType,
		readOnly,
		forceInsert,
		formatted)

	if err != nil {
		return nil, err
	}

	if medium != nil {
		return medium, nil
	}

	// try HardDiskMediumDriver
	hdDriver := drivers_amiga_disk_devices.HardDiskMediumDriver{}

	hdDriver.SetVerboseMode(consts.DRIVERS_VERBOSE_MODE)
	hdDriver.SetDebugMode(consts.DRIVERS_DEBUG_MODE)

	medium, err = hdDriver.Probe(
		consts.FILE_SYSTEM_MOUNT,
		name,
		size,
		_type,
		mountpoint,
		label,
		path,
		fsType,
		ptType,
		readOnly,
		forceInsert,
		formatted)

	if err != nil {
		return nil, err
	}

	if medium != nil {
		return medium, nil
	}

	return nil, nil
}

func formatDeviceIfNeeded(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) bool {

	if _type != "disk" {
		return false
	}

	if readOnly {
		return false
	}

	if !allKeyboardsControl.IsKeysPressed(consts.FORMAT_DEVICE_KEYS) {
		return false
	}

	allKeyboardsControl.ClearPressedKeys()

	log.Println("Formatting device", path)

	n, err := utils.FileUtilsInstance.FileWriteBytes(
		path,
		0,
		consts.EMPTY_DEVICE_HEADER[:],
		os.O_RDWR|os.O_SYNC,
		0,
		nil)

	if err != nil {
		log.Println(err)
		return false
	}

	if n < len(consts.EMPTY_DEVICE_HEADER) {
		log.Println("Cannot format medium in", path)
		return false
	}

	return true
}

func attachedBlockDeviceCallback(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	if utils.BlockDeviceUtilsInstance.IsInternalMedium(name) {
		return
	}

	if utils.BlockDeviceUtilsInstance.IsPoolMedium(name) {
		return
	}

	log.Println("Found new block device", path)

	utils.BlockDeviceUtilsInstance.PrintBlockDevice(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)

	formatted := false

	if formatDeviceIfNeeded(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly) {
		// affect these variables so the driver will
		// think the medium is "unknown"
		mountpoint = ""
		label = ""
		fsType = ""
		ptType = ""

		formatted = true
	}

	_medium, err := ProbeMediumForDriver(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly, formatted)

	if err != nil {
		log.Println(err)
		return
	}

	if _medium == nil {
		log.Println("Unable to find driver for medium", path)
		return
	}

	log.Printf("Medium %v will be handled by %T driver (as %v)\n", path, _medium.GetDriver(), _medium.GetPublicPathname())

	_medium.AddPreReadCallback(preReadCallback)
	_medium.AddPostReadCallback(postReadCallback)
	_medium.AddPreWriteCallback(preWriteCallback)
	_medium.AddPostWriteCallback(postWriteCallback)
	_medium.AddClosedCallback(closedCallback)

	floppyMedium, isFloppy := _medium.(*medium_amiga_disk_devices.FloppyMedium)

	if isFloppy {
		if floppyMedium.GetCachedAdfPathname() == "" {
			if consts.FLOPPY_MUTE_SOUND_NON_CACHED_READ {
				volumeControl.MuteForSecs(consts.FLOPPY_READ_MUTE_SECS)
			}
		}
	}

	fileSystem.AddMedium(_medium)
}

func detachedBlockDeviceCallback(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	if utils.BlockDeviceUtilsInstance.IsInternalMedium(name) {
		return
	}

	if utils.BlockDeviceUtilsInstance.IsPoolMedium(name) {
		return
	}

	log.Println("Removed block device", path)

	utils.BlockDeviceUtilsInstance.PrintBlockDevice(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)

	if _, err := fileSystem.RemoveMediumByDevicePathname(path); err != nil {
		log.Println("Unable to close medium:", path, ":", err)
	}
}

func devicePathnameToAsyncFileOps(devicePathname string) *components.AsyncFileOps {
	index := funk.IndexOfString(floppyDevices, devicePathname)

	if index == 0 {
		return &asyncFileOpsDf0
	} else if index == 1 {
		return &asyncFileOpsDf1
	} else if index == 2 {
		return &asyncFileOpsDf2
	} else if index == 3 {
		return &asyncFileOpsDf3
	}

	return &asyncFileOps
}

func onFloppyRead(_medium interfaces_amiga_disk_devices.Medium, ofst int64) {
	floppyMedium, isFloppy := _medium.(*medium_amiga_disk_devices.FloppyMedium)

	if !isFloppy {
		return
	}

	if !floppyMedium.IsFullyCached() {
		// reading from non-cached floppy medium
		if consts.FLOPPY_MUTE_SOUND_NON_CACHED_READ {
			volumeControl.MuteForSecs(consts.FLOPPY_READ_MUTE_SECS)
		}
	} else {
		devicePathname := floppyMedium.GetDevicePathname()
		async := devicePathnameToAsyncFileOps(devicePathname)
		flag := os.O_RDWR

		if !floppyMedium.IsWritable() {
			flag = os.O_RDONLY
		}

		// reading from cached floppy medium
		// read from real device to move the motor
		async.FileReadBytesDirect(
			devicePathname,
			ofst,
			flag,
			0,
			nil,
			1,
			nil)
	}
}

func onFloppyWrite(_medium interfaces_amiga_disk_devices.Medium) {
	floppyMedium, isFloppy := _medium.(*medium_amiga_disk_devices.FloppyMedium)

	if !isFloppy {
		return
	}

	if !floppyMedium.IsFullyCached() {
		if consts.FLOPPY_MUTE_SOUND_NON_CACHED_WRITE {
			volumeControl.MuteForSecs(consts.FLOPPY_WRITE_MUTE_SECS)
		}
	}
}

func onMediumWrite(_medium interfaces_amiga_disk_devices.Medium) {
	ledControl.BlinkPowerLEDSecs(consts.FLOPPY_WRITE_BLINK_POWER_SECS)
}

func onHardDiskRead(_medium interfaces_amiga_disk_devices.Medium) {
	// hard disk does not have its own struct (derived from MediumBase)
	// so we need to check it against the extension
	isHdf := strings.HasSuffix(
		_medium.GetPublicName(),
		consts.HD_HDF_FULL_EXTENSION)

	if !isHdf {
		return
	}

	// reading from hard-disk, blink the power led
	ledControl.BlinkPowerLEDSecs(consts.HARD_DISK_READ_BLINK_POWER_SECS)
}

func preReadCallback(_medium interfaces_amiga_disk_devices.Medium, path string, buff []byte, ofst int64, fh uint64) {
	// do not put too much logic here, since it will slow down MediumBase.Read
	// or FloppyMediumDriver.Read, same for callbacks related for Write
	onFloppyRead(_medium, ofst)
	onHardDiskRead(_medium)
}

func postReadCallback(_medium interfaces_amiga_disk_devices.Medium, path string, buff []byte, ofst int64, fh uint64, n int, opTimeMs int64) {
	onFloppyRead(_medium, ofst)
	onHardDiskRead(_medium)
}

func preWriteCallback(_medium interfaces_amiga_disk_devices.Medium, path string, buff []byte, ofst int64, fh uint64) {
	onFloppyWrite(_medium)
	onMediumWrite(_medium)
}

func postWriteCallback(_medium interfaces_amiga_disk_devices.Medium, path string, buff []byte, ofst int64, fh uint64, n int, opTimeMs int64) {
	onFloppyWrite(_medium)
	onMediumWrite(_medium)
}

func closedCallback(_medium interfaces_amiga_disk_devices.Medium, err error) {
}

func fileWriteBytesCallback(name string, offset int64, buff []byte, flag int, perm fs.FileMode, useHandle *os.File, n int, err error) {
	ledControl.BlinkPowerLEDSecs(consts.FLOPPY_WRITE_BLINK_POWER_SECS)
}

func outsideAsyncFileWriterCallback(name string, offset int64, buff []byte, flag int, perm fs.FileMode, useHandle *os.File, oneTimeFinal bool) {
	ledControl.BlinkPowerLEDSecs(consts.FLOPPY_WRITE_BLINK_POWER_SECS)

	async := devicePathnameToAsyncFileOps(name)

	if oneTimeFinal {
		async.FileWriteBytesOneTimeFinal(name, offset, buff, flag, perm, useHandle, fileWriteBytesCallback)
	} else {
		async.FileWriteBytes(name, offset, buff, flag, perm, useHandle, 0, fileWriteBytesCallback)
	}
}

func keyEventCallback(sender any, key string, pressed bool) {
}

func preCacheADFCallback(_medium interfaces_amiga_disk_devices.Medium, targetADFpathname string) error {
	size, err := utils.FileUtilsInstance.GetDirSize(cachedAdfsDir)

	if err != nil {
		return err
	}

	if size < consts.CACHED_ADFS_QUOTA {
		// quota not exceeded
		return nil
	}

	// exceeded the quota
	log.Printf("Exceeded the quota for %v (max %v bytes)\n", cachedAdfsDir, consts.CACHED_ADFS_QUOTA)
	log.Println("Trying to find oldest file to delete it")

	oldest, err := utils.FileUtilsInstance.GetDirOldestFile(cachedAdfsDir)

	if err != nil {
		return err
	}

	if oldest == nil {
		return nil
	}

	pathname := path.Join(cachedAdfsDir, oldest.Name())

	log.Printf("Deleting oldest file %v\n", pathname)

	if err = os.Remove(pathname); err != nil {
		return err
	}

	return nil
}

func initCreateDirs(exeDir string) {
	if err := os.MkdirAll(consts.FILE_SYSTEM_MOUNT, 0777); err != nil {
		log.Fatalln(err)
	}

	cachedAdfsDir = path.Join(exeDir, consts.CACHED_ADFS)

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
	floppyDevices = driveDevicesDiscovery.GetFloppies()

	if len(floppyDevices) > 0 {
		log.Println("Physicall floppy drives:")
	}

	for i, devicePathname := range floppyDevices {
		log.Println("\t", i, devicePathname)
	}
}

func printCDROMDevices() {
	cdroms := driveDevicesDiscovery.GetCDROMs()

	if len(cdroms) > 0 {
		log.Println("Physicall CDROM drives:")
	}

	for i, devicePathname := range cdroms {
		log.Println("\t", i, devicePathname)
	}
}

func main() {
	utils.GoUtilsInstance.CheckPlatform()
	utils.UnixUtilsInstance.CheckForRoot()

	exeDir := utils.GoUtilsInstance.MustCwdToExeOrScript()
	logFilename := utils.GoUtilsInstance.MustDuplicateLog(exeDir)

	initCreateDirs(exeDir)

	log.Printf("%v v%v\n", consts.AMIGA_DISK_DEVICES_UNIXNAME, consts.AMIGA_DISK_DEVICES_VERSION)
	log.Printf("Executable directory %v\n", exeDir)
	log.Printf("Log filename %v\n", logFilename)
	log.Println("File system directory " + consts.FILE_SYSTEM_MOUNT)
	log.Println("Cached ADFs directory " + cachedAdfsDir)

	fileSystem.SetMountDir(consts.FILE_SYSTEM_MOUNT)

	discoverDriveDevices()
	printFloppyDevices()
	printCDROMDevices()

	blockDevices.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	blockDevices.SetDebugMode(consts.RUNNERS_DEBUG_MODE)
	fileSystem.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	fileSystem.SetDebugMode(consts.RUNNERS_DEBUG_MODE)
	volumeControl.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	volumeControl.SetDebugMode(consts.RUNNERS_DEBUG_MODE)
	ledControl.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	ledControl.SetDebugMode(consts.RUNNERS_DEBUG_MODE)
	asyncFileOps.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	asyncFileOps.SetDebugMode(consts.RUNNERS_DEBUG_MODE)
	asyncFileOpsDf0.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	asyncFileOpsDf0.SetDebugMode(consts.RUNNERS_DEBUG_MODE)
	asyncFileOpsDf1.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	asyncFileOpsDf1.SetDebugMode(consts.RUNNERS_DEBUG_MODE)
	asyncFileOpsDf2.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	asyncFileOpsDf2.SetDebugMode(consts.RUNNERS_DEBUG_MODE)
	asyncFileOpsDf3.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	asyncFileOpsDf3.SetDebugMode(consts.RUNNERS_DEBUG_MODE)
	allKeyboardsControl.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	allKeyboardsControl.SetDebugMode(consts.RUNNERS_DEBUG_MODE)

	blockDevices.AddAttachedCallback(attachedBlockDeviceCallback)
	blockDevices.AddDetachedCallback(detachedBlockDeviceCallback)
	allKeyboardsControl.SetKeyEventCallback(keyEventCallback)

	fileSystem.Start(&fileSystem)
	blockDevices.Start(&blockDevices)
	volumeControl.Start(&volumeControl)
	ledControl.Start(&ledControl)
	asyncFileOps.Start(&asyncFileOps)
	asyncFileOpsDf0.Start(&asyncFileOpsDf0)
	asyncFileOpsDf1.Start(&asyncFileOpsDf1)
	asyncFileOpsDf2.Start(&asyncFileOpsDf2)
	asyncFileOpsDf3.Start(&asyncFileOpsDf3)
	allKeyboardsControl.Start(&allKeyboardsControl)

	defer fileSystem.Stop(&fileSystem)
	defer blockDevices.Stop(&blockDevices)
	defer volumeControl.Stop(&volumeControl)
	defer ledControl.Stop(&ledControl)
	defer asyncFileOps.Stop(&asyncFileOps)
	defer asyncFileOpsDf0.Stop(&asyncFileOpsDf0)
	defer asyncFileOpsDf1.Stop(&asyncFileOpsDf1)
	defer asyncFileOpsDf2.Stop(&asyncFileOpsDf2)
	defer asyncFileOpsDf3.Stop(&asyncFileOpsDf3)
	defer allKeyboardsControl.Stop(&allKeyboardsControl)

	runnersBlocker.AddRunner(&blockDevices)
	runnersBlocker.AddRunner(&fileSystem)
	runnersBlocker.AddRunner(&volumeControl)
	runnersBlocker.AddRunner(&ledControl)
	runnersBlocker.AddRunner(&asyncFileOps)
	runnersBlocker.AddRunner(&asyncFileOpsDf0)
	runnersBlocker.AddRunner(&asyncFileOpsDf1)
	runnersBlocker.AddRunner(&asyncFileOpsDf2)
	runnersBlocker.AddRunner(&asyncFileOpsDf3)
	runnersBlocker.AddRunner(&allKeyboardsControl)

	runnersBlocker.BlockUntilRunning()
}
