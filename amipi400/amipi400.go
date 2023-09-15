package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	components_amipi400 "github.com/skazanyNaGlany/go.amipi400/amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
	"github.com/thoas/go-funk"
)

var runnersBlocker components.RunnersBlocker
var allKeyboardsControl components.AllKeyboardsControl
var amigaDiskDevicesDiscovery components_amipi400.AmigaDiskDevicesDiscovery
var emulator components_amipi400.AmiberryEmulator
var driveDevicesDiscovery components.DriveDevicesDiscovery
var commander components_amipi400.AmiberryCommander
var blockDevices components.BlockDevices
var mounted = make(map[string]string) // [devicePathname]mountpoint

func adfPathnameToDFIndex(pathname string) int {
	floppyDevices := driveDevicesDiscovery.GetFloppies()

	// get basename and convert it
	// to the device pathname
	baseName := filepath.Base(pathname)
	baseName = strings.ReplaceAll(baseName, "__", "/")
	baseName = strings.Replace(baseName, shared.FLOPPY_ADF_FULL_EXTENSION, "", 1)

	index := funk.IndexOfString(floppyDevices, baseName)

	if index < 0 {
		return 0
	}

	if index >= shared.MAX_ADFS {
		return 0
	}

	return index
}

func isoPathnameToCDIndex(pathname string) int {
	cdromDevices := driveDevicesDiscovery.GetCDROMs()

	// get basename and convert it
	// to the device pathname
	baseName := filepath.Base(pathname)
	baseName = strings.ReplaceAll(baseName, "__", "/")
	baseName = strings.Replace(baseName, shared.CD_ISO_FULL_EXTENSION, "", 1)

	index := funk.IndexOfString(cdromDevices, baseName)

	if index < 0 {
		return 0
	}

	if index >= shared.MAX_CDS {
		return 0
	}

	return index
}

func getHdfFreeSlot() int {
	for index := 0; index < shared.MAX_HDFS; index++ {
		if emulator.GetHdf(index) == "" {
			return index
		}
	}

	return -1
}

func getHdfSlot(pathname string) int {
	for index := 0; index < shared.MAX_HDFS; index++ {
		if emulator.GetHdf(index) == pathname {
			return index
		}
	}

	return -1
}

func attachAdf(index int, pathname string) bool {
	strIndex := fmt.Sprint(index)

	if emulator.GetAdf(index) != "" {
		log.Println("ADF already attached at DF" + strIndex + ", eject it first")

		return false
	}

	log.Println("Attaching", pathname, "to DF"+strIndex)

	emulator.AttachAdf(index, pathname)

	return true
}

func attachIso(index int, pathname string) bool {
	strIndex := fmt.Sprint(index)

	if emulator.GetIso(index) != "" {
		log.Println("ISO already attached at CD" + strIndex + ", eject it first")

		return false
	}

	log.Println("Attaching", pathname, "to CD"+strIndex)

	emulator.AttachCd(index, pathname)

	return true
}

func attachHdf(index int, bootPriority int, pathname string) bool {
	if emulator.GetHdf(index) != "" {
		log.Printf("HDF already attached at DH%v, eject it first\n", index)

		return false
	}

	log.Printf("Attaching %v to DH%v (boot priority %v)\n", pathname, index, bootPriority)

	emulator.AttachHdf(index, bootPriority, pathname)

	utils.UnixUtilsInstance.Sync()
	emulator.HardReset()

	return true
}

func detachIso(index int, pathname string) bool {
	strIndex := fmt.Sprint(index)

	currentIsoPathname := emulator.GetIso(index)

	if currentIsoPathname == "" {
		log.Println("ISO not attached to CD" + strIndex + ", cannot eject")

		return false
	}

	if currentIsoPathname != pathname {
		log.Println(pathname + " not attached to CD" + strIndex + ", cannot eject")

		return false
	}

	log.Println("Detaching", pathname, "from CD"+strIndex)

	emulator.DetachCd(index)

	return true
}

func detachAdf(index int, pathname string) bool {
	strIndex := fmt.Sprint(index)

	currentAdfPathname := emulator.GetAdf(index)

	if currentAdfPathname == "" {
		log.Println("ADF not attached to DF" + strIndex + ", cannot eject")

		return false
	}

	if currentAdfPathname != pathname {
		log.Println(pathname + " not attached to DF" + strIndex + ", cannot eject")

		return false
	}

	log.Println("Detaching", pathname, "from DF"+strIndex)

	emulator.DetachAdf(index)

	return true
}

func detachHdf(index int, pathname string) bool {
	strIndex := fmt.Sprint(index)

	currentHdfPathname := emulator.GetHdf(index)

	if currentHdfPathname == "" {
		log.Println("HDF not attached to DH" + strIndex + ", cannot eject")

		return false
	}

	if currentHdfPathname != pathname {
		log.Println(pathname + " not attached to DH" + strIndex + ", cannot eject")

		return false
	}

	log.Println("Detaching", pathname, "from DH"+strIndex)

	emulator.DetachHdf(index)

	utils.UnixUtilsInstance.Sync()
	emulator.HardReset()

	return true
}

func attachAmigaDiskDeviceAdf(pathname string) {
	index := adfPathnameToDFIndex(pathname)

	oldVolume := emulator.GetFloppySoundVolumeDisk(index)
	emulator.SetFloppySoundVolumeDisk(index, 0)

	if !attachAdf(index, pathname) {
		emulator.SetFloppySoundVolumeDisk(index, oldVolume)
	}
}

func attachAmigaDiskDeviceIso(pathname string) {
	index := isoPathnameToCDIndex(pathname)

	attachIso(index, pathname)
}

func detachAmigaDiskDeviceAdf(pathname string) {
	index := adfPathnameToDFIndex(pathname)

	oldVolume := emulator.GetFloppySoundVolumeDisk(index)
	emulator.SetFloppySoundVolumeDisk(index, 0)

	if !detachAdf(index, pathname) {
		emulator.SetFloppySoundVolumeDisk(index, oldVolume)
	}
}

func detachAmigaDiskDeviceIso(pathname string) {
	index := isoPathnameToCDIndex(pathname)

	detachIso(index, pathname)
}

func attachAmigaDiskDeviceHdf(pathname string) {
	index := getHdfFreeSlot()

	if index == -1 {
		log.Println("Cannot find free HDF slot, eject other HDF")

		return
	}

	attachHdf(index, 0, pathname)
}

func detachAmigaDiskDeviceHdf(pathname string) {
	index := getHdfSlot(pathname)

	if index == -1 {
		log.Println("HDF", pathname, "not attached")

		return
	}

	detachHdf(index, pathname)
}

func attachedAmigaDiskDeviceCallback(pathname string) {
	isAdf := strings.HasSuffix(pathname, shared.FLOPPY_ADF_FULL_EXTENSION)

	if isAdf {
		attachAmigaDiskDeviceAdf(pathname)

		return
	}

	isHdf := strings.HasSuffix(pathname, shared.HD_HDF_FULL_EXTENSION)

	if isHdf {
		attachAmigaDiskDeviceHdf(pathname)

		return
	}

	isIso := strings.HasSuffix(pathname, shared.CD_ISO_FULL_EXTENSION)

	if isIso {
		attachAmigaDiskDeviceIso(pathname)

		return
	}

	log.Fatalln(pathname, "not supported")
}

func detachedAmigaDiskDeviceCallback(pathname string) {
	isAdf := strings.HasSuffix(pathname, shared.FLOPPY_ADF_FULL_EXTENSION)

	if isAdf {
		detachAmigaDiskDeviceAdf(pathname)

		return
	}

	isHdf := strings.HasSuffix(pathname, shared.HD_HDF_FULL_EXTENSION)

	if isHdf {
		detachAmigaDiskDeviceHdf(pathname)

		return
	}

	isIso := strings.HasSuffix(pathname, shared.CD_ISO_FULL_EXTENSION)

	if isIso {
		detachAmigaDiskDeviceIso(pathname)

		return
	}

	log.Fatalln(pathname, "not supported")
}

func isSoftResetKeys() bool {
	releasedKeys := allKeyboardsControl.GetReleasedKeys()
	currentTimestamp := time.Now().UnixMilli()
	goodCount := 0

	for key, pressedTimestamp := range releasedKeys {
		if funk.ContainsString(shared.SOFT_RESET_KEYS, key) {
			if currentTimestamp-pressedTimestamp < shared.SOFT_RESET_KEYS_MAX_MS {
				goodCount++
			}
		}
	}

	return goodCount == len(shared.SOFT_RESET_KEYS)
}

func isHardResetKeys() bool {
	releasedKeys := allKeyboardsControl.GetReleasedKeys()
	currentTimestamp := time.Now().UnixMilli()
	goodCount := 0

	for key, pressedTimestamp := range releasedKeys {
		if funk.ContainsString(shared.HARD_RESET_KEYS, key) {
			if currentTimestamp-pressedTimestamp >= shared.HARD_RESET_KEYS_MIN_MS {
				goodCount++
			}
		}
	}

	return goodCount == len(shared.HARD_RESET_KEYS)
}

func isToggleAutoHeightKeys() bool {
	return allKeyboardsControl.IsKeysReleased(shared.TOGGLE_AUTO_HEIGHT_KEYS)
}

func keyEventCallback(sender any, key string, pressed bool) {
	if isSoftResetKeys() {
		allKeyboardsControl.ClearPressedKeys()
		allKeyboardsControl.ClearReleasedKeys()

		utils.UnixUtilsInstance.Sync()
		emulator.SoftReset()
	} else if isHardResetKeys() {
		allKeyboardsControl.ClearPressedKeys()
		allKeyboardsControl.ClearReleasedKeys()

		// TODO detach and unmount all mediums
		utils.UnixUtilsInstance.Sync()
		emulator.HardReset()
	} else if isToggleAutoHeightKeys() {
		allKeyboardsControl.ClearPressedKeys()
		allKeyboardsControl.ClearReleasedKeys()

		emulator.ToggleAutoHeight()
	}
}

func getDirectoryFirstFile(pathname, extension string) string {
	files := utils.FileUtilsInstance.GetDirFiles(pathname)

	sort.Strings(files)

	for _, pathname := range files {
		if strings.HasSuffix(pathname, extension) {
			return pathname
		}
	}

	return ""
}

func parseMediumLabel(label string, re *regexp.Regexp) (int, int, error) {
	var err error
	var index int64
	var bootPriority int64

	matches := utils.RegExInstance.FindNamedMatches(re, label)

	strIndex, ok1 := matches["index"]
	strBootPriority, ok2 := matches["boot_priority"]

	_ = ok1
	_ = ok2

	// index
	if strIndex == "X" {
		index = -1
	} else {
		index, err = strconv.ParseInt(strIndex, 10, 32)

		if err != nil {
			return 0, 0, err
		}
	}

	// boot priority (optional)
	if strBootPriority != "" {
		bootPriority, err = strconv.ParseInt(strBootPriority, 10, 32)

		if err != nil {
			return 0, 0, err
		}
	}

	return int(index), int(bootPriority), nil
}

func fixMountMedium(devicePathname, label, fsType string) (string, error) {
	log.Println(devicePathname, label, "running Fsck")

	output, err := utils.UnixUtilsInstance.RunFsck(devicePathname)

	if err != nil {
		// fail or not, try to mount it anyway
		log.Println(err)
	}

	log.Println("Fsck output:")
	utils.GoUtilsInstance.LogPrintLines(output)

	target := filepath.Join(shared.AP4_ROOT_MOUNTPOINT, label)

	log.Println(devicePathname, label, "mounting as", target)

	if err := os.MkdirAll(target, 0777); err != nil {
		return "", err
	}

	if err := syscall.Mount(devicePathname, target, fsType, syscall.MS_SYNC, ""); err != nil {
		return "", err
	}

	mounted[devicePathname] = target

	return target, nil
}

func unmountMedium(devicePathname string, mountpoint string, flags int) {
	log.Println("Unmount", mountpoint)

	syscall.Unmount(mountpoint, flags)
	delete(mounted, devicePathname)
}

func attachDFMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	var err error

	index, _, err := parseMediumLabel(label, shared.AP4_MEDIUM_DF_REG_EX)

	if err != nil || index == -1 {
		log.Println(path, label, "cannot get index for medium: ", err)

		return
	}

	if emulator.GetAdf(index) != "" {
		log.Printf("ADF already attached at DF%v, eject it first\n", index)

		return
	}

	if mountpoint != "" {
		unmountMedium(path, mountpoint, 0)
		mountpoint = ""
	}

	// mount the medium if not mounted
	if mountpoint == "" {
		mountpoint, err = fixMountMedium(path, label, fsType)

		if err != nil {
			log.Println(path, label, err)

			return
		}
	}

	// find first .adf file and attach it to the emulator
	firstAdfpathname := getDirectoryFirstFile(mountpoint, shared.FLOPPY_ADF_FULL_EXTENSION)

	if firstAdfpathname == "" {
		log.Println(path, label, "contains no", shared.FLOPPY_ADF_EXTENSION, "files")

		return
	}

	oldVolume := emulator.GetFloppySoundVolumeDisk(index)
	emulator.SetFloppySoundVolumeDisk(index, shared.FLOPPY_DISK_IN_DRIVE_SOUND_VOLUME)

	// TODO unmount when attachAdf return false and the medium
	// was not mounted prevoiusly mountpoint == "", so it means
	// the file was not attached
	if !attachAdf(index, firstAdfpathname) {
		emulator.SetFloppySoundVolumeDisk(index, oldVolume)
	}
}

func attachDHMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	var err error

	index, bootPriority, err := parseMediumLabel(label, shared.AP4_MEDIUM_DH_REG_EX)

	if err != nil || index == -1 {
		log.Println(path, label, "cannot get index for medium: ", err)

		return
	}

	if emulator.GetHdf(index) != "" {
		log.Printf("HDF already attached at DH%v, eject it first\n", index)

		return
	}

	// TODO unmount all
	if mountpoint != "" {
		unmountMedium(path, mountpoint, 0)
		mountpoint = ""
	}

	// mount the medium if not mounted
	if mountpoint == "" {
		mountpoint, err = fixMountMedium(path, label, fsType)

		if err != nil {
			log.Println(path, label, err)

			return
		}
	}

	// find first .hdf file and attach it to the emulator
	firstHdfpathname := getDirectoryFirstFile(mountpoint, shared.HD_HDF_FULL_EXTENSION)

	if firstHdfpathname == "" {
		log.Println(path, label, "contains no", shared.HD_HDF_EXTENSION, "files")

		return
	}

	// TODO unmount when attachHdf return false and the medium
	// was not mounted prevoiusly mountpoint == "", so it means
	// the file was not attached
	attachHdf(index, bootPriority, firstHdfpathname)
}

func attachCDMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	var err error

	index, _, err := parseMediumLabel(label, shared.AP4_MEDIUM_CD_REG_EX)

	if err != nil || index == -1 {
		log.Println(path, label, "cannot get index for medium: ", err)

		return
	}

	if emulator.GetIso(index) != "" {
		log.Printf("ISO already attached at CD%v, eject it first\n", index)

		return
	}

	if mountpoint != "" {
		unmountMedium(path, mountpoint, 0)
		mountpoint = ""
	}

	// mount the medium if not mounted
	if mountpoint == "" {
		mountpoint, err = fixMountMedium(path, label, fsType)

		if err != nil {
			log.Println(path, label, err)

			return
		}
	}

	// find first .iso file and attach it to the emulator
	firstIsoPathname := getDirectoryFirstFile(mountpoint, shared.CD_ISO_FULL_EXTENSION)

	if firstIsoPathname == "" {
		log.Println(path, label, "contains no", shared.CD_ISO_EXTENSION, "files")

		return
	}

	// TODO unmount when attachIso return false and the medium
	// was not mounted prevoiusly mountpoint == "", so it means
	// the file was not attached
	attachIso(index, firstIsoPathname)
}

func attachMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	if shared.AP4_MEDIUM_DF_REG_EX.MatchString(label) {
		attachDFMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
	} else if shared.AP4_MEDIUM_DH_REG_EX.MatchString(label) {
		attachDHMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
	} else if shared.AP4_MEDIUM_CD_REG_EX.MatchString(label) {
		attachCDMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
	}
}

func detachDFMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	_mountpoint, exists := mounted[path]

	if !exists || _mountpoint == "" {
		log.Println(path, label, "not mounted")

		return
	}

	for i := 0; i < shared.MAX_ADFS; i++ {
		adfPathname := emulator.GetAdf(i)

		if adfPathname == "" {
			continue
		}

		if !strings.HasPrefix(adfPathname, _mountpoint) {
			continue
		}

		oldVolume := emulator.GetFloppySoundVolumeDisk(i)
		emulator.SetFloppySoundVolumeDisk(i, 0)

		if !detachAdf(i, adfPathname) {
			emulator.SetFloppySoundVolumeDisk(i, oldVolume)
		}
	}

	unmountMedium(path, _mountpoint, syscall.MNT_DETACH)
}

func detachDHMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	_mountpoint, exists := mounted[path]

	if !exists || _mountpoint == "" {
		log.Println(path, label, "not mounted")

		return
	}

	for i := 0; i < shared.MAX_HDFS; i++ {
		hdfPathname := emulator.GetHdf(i)

		if hdfPathname == "" {
			continue
		}

		if !strings.HasPrefix(hdfPathname, _mountpoint) {
			continue
		}

		detachHdf(i, hdfPathname)
	}

	unmountMedium(path, _mountpoint, syscall.MNT_DETACH)
}

func detachCDMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	_mountpoint, exists := mounted[path]

	if !exists || _mountpoint == "" {
		log.Println(path, label, "not mounted")

		return
	}

	for i := 0; i < shared.MAX_CDS; i++ {
		hdfPathname := emulator.GetIso(i)

		if hdfPathname == "" {
			continue
		}

		if !strings.HasPrefix(hdfPathname, _mountpoint) {
			continue
		}

		detachIso(i, hdfPathname)
	}

	unmountMedium(path, _mountpoint, syscall.MNT_DETACH)
}

func detachMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	if shared.AP4_MEDIUM_DF_REG_EX.MatchString(label) {
		detachDFMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
	} else if shared.AP4_MEDIUM_DH_REG_EX.MatchString(label) {
		detachDHMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
	} else if shared.AP4_MEDIUM_CD_REG_EX.MatchString(label) {
		detachCDMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
	}
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

	attachMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
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

	detachMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
}

func discoverDriveDevices() {
	log.Println("Getting information about physicall drives")

	if err := driveDevicesDiscovery.Refresh(); err != nil {
		log.Println("Perhaps you need to run this program as root")
		log.Fatalln(err)
	}
}

func printFloppyDevices() {
	floppyDevices := driveDevicesDiscovery.GetFloppies()

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

	log.Printf("%v v%v\n", shared.AMIPI400_UNIXNAME, shared.AMIPI400_VERSION)
	log.Printf("Executable directory %v\n", exeDir)
	log.Printf("Log filename %v\n", logFilename)

	amigaDiskDevicesDiscovery.SetAttachedAmigaDiskDeviceCallback(attachedAmigaDiskDeviceCallback)
	amigaDiskDevicesDiscovery.SetDetachedAmigaDiskDeviceCallback(detachedAmigaDiskDeviceCallback)
	amigaDiskDevicesDiscovery.SetMountpoint(shared.FILE_SYSTEM_MOUNT)
	allKeyboardsControl.SetKeyEventCallback(keyEventCallback)
	commander.SetTmpIniPathname(shared.AMIBERRY_EMULATOR_TMP_INI_PATHNAME)
	emulator.SetExecutablePathname(shared.AMIBERRY_EXE_PATHNAME)
	emulator.SetConfigPathname(shared.AMIPI400_AMIBERRY_CONFIG_PATHNAME)
	emulator.SetAmiberryCommander(&commander)
	blockDevices.AddAttachedCallback(attachedBlockDeviceCallback)
	blockDevices.AddDetachedCallback(detachedBlockDeviceCallback)

	discoverDriveDevices()
	printFloppyDevices()
	printCDROMDevices()

	amigaDiskDevicesDiscovery.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	amigaDiskDevicesDiscovery.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	allKeyboardsControl.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	allKeyboardsControl.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	emulator.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	emulator.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	commander.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	commander.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	blockDevices.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	blockDevices.SetDebugMode(shared.RUNNERS_DEBUG_MODE)

	amigaDiskDevicesDiscovery.Start(&amigaDiskDevicesDiscovery)
	allKeyboardsControl.Start(&allKeyboardsControl)
	emulator.Start(&emulator)
	commander.Start(&commander)
	blockDevices.Start(&blockDevices)

	defer amigaDiskDevicesDiscovery.Stop(&amigaDiskDevicesDiscovery)
	defer allKeyboardsControl.Stop(&allKeyboardsControl)
	defer emulator.Stop(&emulator)
	defer commander.Stop(&commander)
	defer blockDevices.Stop(&blockDevices)

	runnersBlocker.AddRunner(&amigaDiskDevicesDiscovery)
	runnersBlocker.AddRunner(&allKeyboardsControl)
	runnersBlocker.AddRunner(&emulator)
	runnersBlocker.AddRunner(&commander)
	runnersBlocker.AddRunner(&blockDevices)

	runnersBlocker.BlockUntilRunning()
}
