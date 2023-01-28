package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	components_amipi400 "github.com/skazanyNaGlany/go.amipi400/amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/components/utils"
	"github.com/skazanyNaGlany/go.amipi400/consts"
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
	baseName = strings.Replace(baseName, consts.FLOPPY_ADF_FULL_EXTENSION, "", 1)

	index := funk.IndexOfString(floppyDevices, baseName)

	if index < 0 {
		return 0
	}

	if index >= consts.MAX_ADFS {
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
	baseName = strings.Replace(baseName, consts.CD_ISO_FULL_EXTENSION, "", 1)

	index := funk.IndexOfString(cdromDevices, baseName)

	if index < 0 {
		return 0
	}

	if index >= consts.MAX_CDS {
		return 0
	}

	return index
}

func getHdfFreeSlot() int {
	for index := 0; index < consts.MAX_HDFS; index++ {
		if emulator.GetHdf(index) == "" {
			return index
		}
	}

	return -1
}

func getHdfSlot(pathname string) int {
	for index := 0; index < consts.MAX_HDFS; index++ {
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

func attachAmigaDiskDeviceAdf(pathname string) {
	index := adfPathnameToDFIndex(pathname)

	attachAdf(index, pathname)
}

func attachAmigaDiskDeviceIso(pathname string) {
	index := isoPathnameToCDIndex(pathname)
	strIndex := fmt.Sprint(index)

	if emulator.GetIso(index) != "" {
		log.Println("ISO already attached at CD" + strIndex + ", eject it first")

		return
	}

	log.Println("Attaching", pathname, "to CD"+strIndex)

	emulator.AttachCd(index, pathname)
}

func detachAmigaDiskDeviceAdf(pathname string) {
	index := adfPathnameToDFIndex(pathname)

	detachAdf(index, pathname)
}

func detachAmigaDiskDeviceIso(pathname string) {
	index := isoPathnameToCDIndex(pathname)
	strIndex := fmt.Sprint(index)

	currentIsoPathname := emulator.GetIso(index)

	if currentIsoPathname == "" {
		log.Println("ISO not attached to CD" + strIndex + ", cannot eject")

		return
	}

	if currentIsoPathname != pathname {
		log.Println(pathname + " not attached to CD" + strIndex + ", cannot eject")

		return
	}

	log.Println("Detaching", pathname, "from CD"+strIndex)

	emulator.DetachCd(index)
}

func attachAmigaDiskDeviceHdf(pathname string) {
	slotIndex := getHdfFreeSlot()

	if slotIndex == -1 {
		log.Println("Cannot find free HDF slot, eject other HDF")

		return
	}

	strIndex := fmt.Sprint(slotIndex)

	log.Println("Attaching", pathname, "to DH"+strIndex)

	emulator.AttachHdf(slotIndex, pathname)
	emulator.HardReset()
}

func detachAmigaDiskDeviceHdf(pathname string) {
	slotIndex := getHdfSlot(pathname)

	if slotIndex == -1 {
		log.Println("HDF", pathname, "not attached")

		return
	}

	strIndex := fmt.Sprint(slotIndex)

	log.Println("Detaching", pathname, "from DH"+strIndex)

	emulator.DetachHdf(slotIndex)
	emulator.HardReset()
}

func attachedAmigaDiskDeviceCallback(pathname string) {
	isAdf := strings.HasSuffix(pathname, consts.FLOPPY_ADF_FULL_EXTENSION)

	if isAdf {
		attachAmigaDiskDeviceAdf(pathname)

		return
	}

	isHdf := strings.HasSuffix(pathname, consts.HD_HDF_FULL_EXTENSION)

	if isHdf {
		attachAmigaDiskDeviceHdf(pathname)

		return
	}

	isIso := strings.HasSuffix(pathname, consts.CD_ISO_FULL_EXTENSION)

	if isIso {
		attachAmigaDiskDeviceIso(pathname)

		return
	}

	log.Fatalln(pathname, "not supported")
}

func detachedAmigaDiskDeviceCallback(pathname string) {
	isAdf := strings.HasSuffix(pathname, consts.FLOPPY_ADF_FULL_EXTENSION)

	if isAdf {
		detachAmigaDiskDeviceAdf(pathname)

		return
	}

	isHdf := strings.HasSuffix(pathname, consts.HD_HDF_FULL_EXTENSION)

	if isHdf {
		detachAmigaDiskDeviceHdf(pathname)

		return
	}

	isIso := strings.HasSuffix(pathname, consts.CD_ISO_FULL_EXTENSION)

	if isIso {
		detachAmigaDiskDeviceIso(pathname)

		return
	}

	log.Fatalln(pathname, "not supported")
}

func keyEventCallback(sender any, key string, pressed bool) {
	if allKeyboardsControl.IsKeysPressed(consts.SOFT_RESET_KEYS) {
		allKeyboardsControl.ClearPressedKeys()

		emulator.SoftReset()
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

func mediumLabelToIndex(label string) int {
	char := label[6]

	if char == 'X' {
		return -1
	}

	index, err := strconv.ParseInt(string(char), 10, 32)

	if err != nil {
		return -1
	}

	return int(index)
}

func attachDFMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	// mount the medium if not mounted
	if mountpoint == "" {
		log.Println(path, label, "running fsck")

		output, err := utils.UnixUtilsInstance.RunFsck(path)

		if err != nil {
			// fail or not, try to mount it anyway
			log.Println(err)
		}

		log.Println("Fsck output:")
		utils.GoUtilsInstance.LogPrintLines(output)

		target := filepath.Join(consts.AP4_ROOT_MOUNTPOINT, label)

		log.Println(path, label, "mounting as", target)

		if err := os.MkdirAll(target, 0777); err != nil {
			log.Println(err)

			return
		}

		if err := syscall.Mount(path, target, fsType, syscall.MS_SYNC, ""); err != nil {
			log.Println(err)

			return
		}

		mounted[path] = target

		mountpoint = target
	}

	// find first .adf file and attach it to the emulator
	firstAdfpathname := getDirectoryFirstFile(mountpoint, consts.FLOPPY_ADF_FULL_EXTENSION)

	if firstAdfpathname == "" {
		log.Println(path, label, "contains no", consts.FLOPPY_ADF_EXTENSION, "files")

		return
	}

	index := mediumLabelToIndex(label)

	if index == -1 {
		log.Println(path, label, "cannot get index for medium")

		return
	}

	if attachAdf(index, firstAdfpathname) {
		// TODO set floppy sound
	}
}

func attachMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	if consts.AP4_MEDIUM_DF_REG_EX.MatchString(label) {
		attachDFMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
	}
}

func detachDFMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	_mountpoint, exists := mounted[path]

	if !exists {
		log.Println(path, label, "not mounted")

		return
	}

	for i := 0; i < consts.MAX_ADFS; i++ {
		adfPathname := emulator.GetAdf(i)

		if adfPathname == "" {
			continue
		}

		if strings.HasPrefix(adfPathname, _mountpoint) {
			if detachAdf(i, adfPathname) {
				// TODO set floppy sound
			}
		}
	}
}

func detachMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	if consts.AP4_MEDIUM_DF_REG_EX.MatchString(label) {
		detachDFMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
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

	log.Printf("%v v%v\n", consts.AMIPI400_UNIXNAME, consts.AMIPI400_VERSION)
	log.Printf("Executable directory %v\n", exeDir)
	log.Printf("Log filename %v\n", logFilename)

	amigaDiskDevicesDiscovery.SetAttachedAmigaDiskDeviceCallback(attachedAmigaDiskDeviceCallback)
	amigaDiskDevicesDiscovery.SetDetachedAmigaDiskDeviceCallback(detachedAmigaDiskDeviceCallback)
	amigaDiskDevicesDiscovery.SetMountpoint(consts.FILE_SYSTEM_MOUNT)
	allKeyboardsControl.SetKeyEventCallback(keyEventCallback)
	commander.SetTmpIniPathname(consts.AMIBERRY_EMULATOR_TMP_INI_PATHNAME)
	emulator.SetExecutablePathname(consts.AMIBERRY_EXE_PATHNAME)
	emulator.SetConfigPathname(consts.AMIPI400_AMIBERRY_CONFIG_PATHNAME)
	emulator.SetAmiberryCommander(&commander)
	blockDevices.AddAttachedCallback(attachedBlockDeviceCallback)
	blockDevices.AddDetachedCallback(detachedBlockDeviceCallback)

	discoverDriveDevices()
	printFloppyDevices()
	printCDROMDevices()

	amigaDiskDevicesDiscovery.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	amigaDiskDevicesDiscovery.SetDebugMode(consts.RUNNERS_DEBUG_MODE)
	allKeyboardsControl.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	allKeyboardsControl.SetDebugMode(consts.RUNNERS_DEBUG_MODE)
	emulator.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	emulator.SetDebugMode(consts.RUNNERS_DEBUG_MODE)
	commander.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	commander.SetDebugMode(consts.RUNNERS_DEBUG_MODE)
	blockDevices.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
	blockDevices.SetDebugMode(consts.RUNNERS_DEBUG_MODE)

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
