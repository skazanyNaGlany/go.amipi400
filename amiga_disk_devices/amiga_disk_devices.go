package main

import (
	"io/fs"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	components_amiga_disk_devices "github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/components"
	drivers_amiga_disk_devices "github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/components/drivers"
	medium_amiga_disk_devices "github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/components/medium"
	interfaces_amiga_disk_devices "github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/interfaces"
	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
	"github.com/thoas/go-funk"
)

var blockDevices components.BlockDevices
var fileSystem components_amiga_disk_devices.ADDFileSystem
var runnersBlocker components.RunnersBlocker
var driveDevicesDiscovery components.DriveDevicesDiscovery
var volumeControl components_amiga_disk_devices.VolumeControl
var powerLEDControl components.PowerLEDControl
var numLockLEDControl components.NumLockLEDControl
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

	forceInsert := allKeyboardsControl.IsKeysPressed(shared.FORCE_INSERT_KEYS)

	if forceInsert {
		// perform only one special action at a time
		allKeyboardsControl.ClearPressedKeys()

		powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
		defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)
	}

	// try FloppyMediumDriver
	floppyDriver := drivers_amiga_disk_devices.FloppyMediumDriver{}

	floppyDriver.SetCachedAdfsDirectory(cachedAdfsDir)
	floppyDriver.SetVerboseMode(shared.DRIVERS_VERBOSE_MODE)
	floppyDriver.SetDebugMode(shared.DRIVERS_DEBUG_MODE)
	floppyDriver.SetOutsideAsyncFileWriterCallback(outsideAsyncFileWriterCallback)
	floppyDriver.SetPreCacheADFCallback(preCacheADFCallback)

	medium, err := floppyDriver.Probe(
		shared.FILE_SYSTEM_MOUNT,
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
		if forceInsert {
			powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
			defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)
		}

		return nil, err
	}

	if medium != nil {
		return medium, nil
	}

	// try CDMediumDriver
	cdDriver := drivers_amiga_disk_devices.CDMediumDriver{}

	cdDriver.SetVerboseMode(shared.DRIVERS_VERBOSE_MODE)
	cdDriver.SetDebugMode(shared.DRIVERS_DEBUG_MODE)

	medium, err = cdDriver.Probe(
		shared.FILE_SYSTEM_MOUNT,
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
		if forceInsert {
			powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
			defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)
		}

		return nil, err
	}

	if medium != nil {
		return medium, nil
	}

	// try HardDiskMediumDriver
	hdDriver := drivers_amiga_disk_devices.HardDiskMediumDriver{}

	hdDriver.SetVerboseMode(shared.DRIVERS_VERBOSE_MODE)
	hdDriver.SetDebugMode(shared.DRIVERS_DEBUG_MODE)

	medium, err = hdDriver.Probe(
		shared.FILE_SYSTEM_MOUNT,
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
		if forceInsert {
			powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
			defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)
		}

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

	if !allKeyboardsControl.IsKeysPressed(shared.FORMAT_DEVICE_KEYS) {
		return false
	}

	allKeyboardsControl.ClearPressedKeys()

	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)

	log.Println("Formatting device", path)

	n, err := utils.FileUtilsInstance.FileWriteBytes(
		path,
		0,
		shared.EMPTY_DEVICE_HEADER[:],
		os.O_RDWR|os.O_SYNC,
		0,
		nil)

	if err != nil {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)

		log.Println(err)
		return false
	}

	if n < len(shared.EMPTY_DEVICE_HEADER) {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)

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
			if shared.FLOPPY_MUTE_SOUND_NON_CACHED_READ {
				volumeControl.MuteForSecs(shared.FLOPPY_READ_MUTE_SECS)
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
		if shared.FLOPPY_MUTE_SOUND_NON_CACHED_READ {
			volumeControl.MuteForSecs(shared.FLOPPY_READ_MUTE_SECS)
		}
	} else {
		devicePathname := floppyMedium.GetDevicePathname()
		async := devicePathnameToAsyncFileOps(devicePathname)
		flag := os.O_RDWR

		if !floppyMedium.IsWritable() {
			flag = os.O_RDONLY
		}

		deviceDirectIOHandle, err := floppyMedium.GetDeviceDirectIOHandle()

		if err != nil {
			log.Println(devicePathname, err)

			return
		}

		// reading from cached floppy medium
		// read from real device to move the motor
		async.FileReadBytesDirect(
			devicePathname,
			ofst,
			flag,
			0,
			deviceDirectIOHandle,
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
		if shared.FLOPPY_MUTE_SOUND_NON_CACHED_WRITE {
			volumeControl.MuteForSecs(shared.FLOPPY_WRITE_MUTE_SECS)
		}
	}
}

func onMediumWrite(_medium interfaces_amiga_disk_devices.Medium) {
	powerLEDControl.BlinkPowerLEDSecs(shared.FLOPPY_WRITE_BLINK_POWER_SECS)
}

func onHardDiskRead(_medium interfaces_amiga_disk_devices.Medium) {
	// hard disk does not have its own struct (derived from MediumBase)
	// so we need to check it against the extension
	isHdf := strings.HasSuffix(
		_medium.GetPublicName(),
		shared.HD_HDF_FULL_EXTENSION)

	if !isHdf {
		return
	}

	// reading from hard-disk, blink the power led
	powerLEDControl.BlinkPowerLEDSecs(shared.HARD_DISK_READ_BLINK_POWER_SECS)
}

func preReadCallback(_medium interfaces_amiga_disk_devices.Medium, path string, buff []byte, ofst int64, fh uint64) {
	// do not put too much logic here, since it will slow down MediumBase.Read
	// or FloppyMediumDriver.Read, same for callbacks related for Write
	onFloppyRead(_medium, ofst)
	onHardDiskRead(_medium)
}

func postReadCallback(_medium interfaces_amiga_disk_devices.Medium, path string, buff []byte, ofst int64, fh uint64, n int, opTimeMs int64) {
	// onFloppyRead(_medium, ofst)
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
	powerLEDControl.BlinkPowerLEDSecs(shared.FLOPPY_WRITE_BLINK_POWER_SECS)
}

func outsideAsyncFileWriterCallback(name string, offset int64, buff []byte, flag int, perm fs.FileMode, useHandle *os.File, oneTimeFinal bool) {
	powerLEDControl.BlinkPowerLEDSecs(shared.FLOPPY_WRITE_BLINK_POWER_SECS)

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

	if size < shared.CACHED_ADFS_QUOTA {
		// quota not exceeded
		return nil
	}

	// exceeded the quota
	log.Printf("Exceeded the quota for %v (max %v bytes)\n", cachedAdfsDir, shared.CACHED_ADFS_QUOTA)
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
	if err := os.MkdirAll(shared.FILE_SYSTEM_MOUNT, 0777); err != nil {
		log.Fatalln(err)
	}

	cachedAdfsDir = path.Join(exeDir, shared.CACHED_ADFS)

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

func stopServices() {
	fileSystem.Stop(&fileSystem)
	blockDevices.Stop(&blockDevices)
	volumeControl.Stop(&volumeControl)
	powerLEDControl.Stop(&powerLEDControl)
	asyncFileOps.Stop(&asyncFileOps)
	asyncFileOpsDf0.Stop(&asyncFileOpsDf0)
	asyncFileOpsDf1.Stop(&asyncFileOpsDf1)
	asyncFileOpsDf2.Stop(&asyncFileOpsDf2)
	asyncFileOpsDf3.Stop(&asyncFileOpsDf3)
	allKeyboardsControl.Stop(&allKeyboardsControl)
}

func gracefulShutdown() {
	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan

	stopServices()
}

func main() {
	utils.GoUtilsInstance.CheckPlatform()
	utils.UnixUtilsInstance.CheckForRoot()
	utils.SysUtilsInstance.CheckForExecutables(
		shared.AMIGA_DISK_DEVICES_NEEDED_EXECUTABLES)

	exeDir := utils.GoUtilsInstance.MustCwdToExeOrScript()
	logFilename := utils.GoUtilsInstance.MustDuplicateLog(exeDir)

	initCreateDirs(exeDir)

	log.Printf("%v v%v\n", shared.AMIGA_DISK_DEVICES_UNIXNAME, shared.AMIGA_DISK_DEVICES_VERSION)
	log.Printf("Executable directory %v\n", exeDir)
	log.Printf("Log filename %v\n", logFilename)
	log.Println("File system directory " + shared.FILE_SYSTEM_MOUNT)
	log.Println("Cached ADFs directory " + cachedAdfsDir)

	fileSystem.SetMountDir(shared.FILE_SYSTEM_MOUNT)

	discoverDriveDevices()
	printFloppyDevices()
	printCDROMDevices()

	blockDevices.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	blockDevices.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	fileSystem.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	fileSystem.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	volumeControl.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	volumeControl.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	powerLEDControl.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	powerLEDControl.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	asyncFileOps.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	asyncFileOps.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	asyncFileOpsDf0.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	asyncFileOpsDf0.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	asyncFileOpsDf1.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	asyncFileOpsDf1.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	asyncFileOpsDf2.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	asyncFileOpsDf2.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	asyncFileOpsDf3.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	asyncFileOpsDf3.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	allKeyboardsControl.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	allKeyboardsControl.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	numLockLEDControl.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	numLockLEDControl.SetDebugMode(shared.RUNNERS_DEBUG_MODE)

	blockDevices.AddAttachedCallback(attachedBlockDeviceCallback)
	blockDevices.AddDetachedCallback(detachedBlockDeviceCallback)
	allKeyboardsControl.SetKeyEventCallback(keyEventCallback)

	fileSystem.Start(&fileSystem)
	blockDevices.Start(&blockDevices)
	volumeControl.Start(&volumeControl)
	powerLEDControl.Start(&powerLEDControl)
	asyncFileOps.Start(&asyncFileOps)
	asyncFileOpsDf0.Start(&asyncFileOpsDf0)
	asyncFileOpsDf1.Start(&asyncFileOpsDf1)
	asyncFileOpsDf2.Start(&asyncFileOpsDf2)
	asyncFileOpsDf3.Start(&asyncFileOpsDf3)
	allKeyboardsControl.Start(&allKeyboardsControl)
	numLockLEDControl.Start(&numLockLEDControl)

	runnersBlocker.AddRunner(&blockDevices)
	runnersBlocker.AddRunner(&fileSystem)
	runnersBlocker.AddRunner(&volumeControl)
	runnersBlocker.AddRunner(&powerLEDControl)
	runnersBlocker.AddRunner(&asyncFileOps)
	runnersBlocker.AddRunner(&asyncFileOpsDf0)
	runnersBlocker.AddRunner(&asyncFileOpsDf1)
	runnersBlocker.AddRunner(&asyncFileOpsDf2)
	runnersBlocker.AddRunner(&asyncFileOpsDf3)
	runnersBlocker.AddRunner(&allKeyboardsControl)
	runnersBlocker.AddRunner(&numLockLEDControl)

	go gracefulShutdown()

	runnersBlocker.BlockUntilRunning()
}
