package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
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
	"gopkg.in/ini.v1"
)

var runnersBlocker components.RunnersBlocker
var allKeyboardsControl components.AllKeyboardsControl
var amigaDiskDevicesDiscovery components_amipi400.AmigaDiskDevicesDiscovery
var emulator components_amipi400.AmiberryEmulator
var driveDevicesDiscovery components.DriveDevicesDiscovery
var commander components_amipi400.AmiberryCommander
var blockDevices components.BlockDevices
var mounted = make(map[string]string)           // [devicePathname]mountpoint
var mediumConfig = make(map[string]*ini.File)   // [devicePathname]*ini.File
var mountpointFiles = make(map[string][]string) // [mountpoint]romFiles
var dfIndexMountpoint = make(map[int]string)    // [driveIndex]mountpoint
var hdIndexMountpoint = make(map[int]string)    // [driveIndex]mountpoint
var cdIndexMountpoint = make(map[int]string)    // [driveIndex]mountpoint

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
		if emulator.GetHd(index) == "" {
			return index
		}
	}

	return shared.DRIVE_INDEX_UNSPECIFIED
}

func getHdfSlot(pathname string) int {
	for index := 0; index < shared.MAX_HDFS; index++ {
		if emulator.GetHd(index) == pathname {
			return index
		}
	}

	return shared.DRIVE_INDEX_UNSPECIFIED
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
	if emulator.GetHd(index) != "" {
		log.Printf("HDF already attached at DH%v, eject it first\n", index)

		return false
	}

	log.Printf("Attaching %v to DH%v (boot priority %v)\n", pathname, index, bootPriority)

	emulator.AttachHdf(index, bootPriority, pathname)

	utils.UnixUtilsInstance.Sync()
	emulator.HardReset()

	return true
}

func attachHdDir(index int, bootPriority int, pathname string) bool {
	if emulator.GetHd(index) != "" {
		log.Printf("HDF already attached at DH%v, eject it first\n", index)

		return false
	}

	log.Printf("Attaching %v to DH%v (boot priority %v)\n", pathname, index, bootPriority)

	emulator.AttachHdDir(index, bootPriority, pathname)

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

func detachHd(index int, pathname string) bool {
	strIndex := fmt.Sprint(index)

	currentHdfPathname := emulator.GetHd(index)

	if currentHdfPathname == "" {
		log.Println("HDF not attached to DH" + strIndex + ", cannot eject")

		return false
	}

	if currentHdfPathname != pathname {
		log.Println(pathname + " not attached to DH" + strIndex + ", cannot eject")

		return false
	}

	log.Println("Detaching", pathname, "from DH"+strIndex)

	emulator.DetachHd(index)

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

	if index == shared.DRIVE_INDEX_UNSPECIFIED {
		log.Println("Cannot find free HDF slot, eject other HDF")

		return
	}

	attachHdf(index, 0, pathname)
}

func detachAmigaDiskDeviceHdf(pathname string) {
	index := getHdfSlot(pathname)

	if index == shared.DRIVE_INDEX_UNSPECIFIED {
		log.Println("HDF", pathname, "not attached")

		return
	}

	detachHd(index, pathname)
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
			// this is not a mistake, the user must hold CTRL-ALT-ALTGR less seconds
			// than shared.HARD_RESET_KEYS_MIN_MS
			if currentTimestamp-pressedTimestamp < shared.HARD_RESET_KEYS_MIN_MS {
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

func isReleasedKey(key string) bool {
	releasedKeys := allKeyboardsControl.GetReleasedKeys()

	_, exists := releasedKeys[key]

	return exists
}

func getReleasedKeysSequence() []string {
	released := make([]string, 0)

	for _, ks := range allKeyboardsControl.GetKeysSequence() {
		if ks.Pressed {
			continue
		}

		released = append(released, ks.Key)
	}

	return released
}

func getKeyboardCommand() string {
	releasedSequence := getReleasedKeysSequence()
	lenReleasedSequence := len(releasedSequence)

	if lenReleasedSequence < 4 {
		return ""
	}

	if releasedSequence[0] != shared.KEY_TAB ||
		releasedSequence[1] != shared.KEY_TAB ||
		releasedSequence[lenReleasedSequence-1] != shared.KEY_TAB ||
		releasedSequence[lenReleasedSequence-2] != shared.KEY_TAB {
		return ""
	}

	releasedSequence = releasedSequence[2:]
	lenReleasedSequence = len(releasedSequence)

	releasedSequence = releasedSequence[0 : lenReleasedSequence-2]

	// replace some key codes into real keys
	for i, key := range releasedSequence {
		if key == shared.KEY_SPACE {
			releasedSequence[i] = " "
		} else {
			releasedSequence[i] = key
		}
	}

	return strings.Join(releasedSequence, "")
}

func clearAllKeyboardsControl() {
	allKeyboardsControl.ClearPressedKeys()
	allKeyboardsControl.ClearReleasedKeys()
	allKeyboardsControl.ClearKeysSequence()
}

func processKeyboardCommand(keyboardCommand string) {
	if dfEjectRule := utils.RegExInstance.FindNamedMatches(
		shared.DF_EJECT_FROM_SOURCE_INDEX_RE,
		keyboardCommand); len(dfEjectRule) > 0 {
		// example: df0
		dfEjectFromSourceIndex(dfEjectRule["source_index"])
	} else if dfSourceTargetRule := utils.RegExInstance.FindNamedMatches(
		shared.DF_INSERT_FROM_SOURCE_TO_TARGET_INDEX_RE,
		keyboardCommand); len(dfSourceTargetRule) > 0 {
		// example: df0kwater disk 2df1
		dfInsertFromSourceIndexToTargetIndex(
			dfSourceTargetRule["filename_part"],
			dfSourceTargetRule["source_index"],
			dfSourceTargetRule["target_index"])
	} else if dfSourceRule := utils.RegExInstance.FindNamedMatches(
		shared.DF_INSERT_FROM_SOURCE_INDEX_RE,
		keyboardCommand); len(dfSourceRule) > 0 {
		// example: df0traps
		dfInsertFromSourceIndexToTargetIndex(
			dfSourceRule["filename_part"],
			dfSourceRule["source_index"],
			shared.DRIVE_INDEX_UNSPECIFIED_STR)
	}
}

// DF medium - detach rom from source DF<index>, find other rom by name pattern (filename part)
// and attach it to target DF<index>
//
// target_index as -1 means (unspecified), so it will be source_index
func dfInsertFromSourceIndexToTargetIndex(filenamePart, sourceIndex, targetIndex string) {
	if targetIndex == "N" {
		dfInsertFromSourceIndexToManyIndex(filenamePart, sourceIndex)
		return
	}

	filenamePart = strings.TrimSpace(filenamePart)
	sourceIndexInt, _ := utils.StringUtilsInstance.StringToInt(sourceIndex, 10, 16)
	targetIndexInt, _ := utils.StringUtilsInstance.StringToInt(targetIndex, 10, 16)

	if filenamePart == "" {
		return
	}

	if targetIndexInt == shared.DRIVE_INDEX_UNSPECIFIED {
		targetIndexInt = sourceIndexInt
	}

	if sourceIndexInt > shared.MAX_ADFS-1 || targetIndexInt > shared.MAX_ADFS-1 {
		return
	}

	mountpoint, mountpointExists := dfIndexMountpoint[sourceIndexInt]

	if !mountpointExists {
		return
	}

	targetIndexAdf := emulator.GetAdf(targetIndexInt)

	if targetIndexAdf != "" {
		targetIndexOldVolume := emulator.GetFloppySoundVolumeDisk(targetIndexInt)
		emulator.SetFloppySoundVolumeDisk(targetIndexInt, 0)

		if !detachAdf(targetIndexInt, targetIndexAdf) {
			emulator.SetFloppySoundVolumeDisk(targetIndexInt, targetIndexOldVolume)
			return
		}
	}

	foundAdfPathname := findSimilarROMFile(mountpoint, filenamePart)

	if foundAdfPathname == "" {
		return
	}

	if isAdfAttached(foundAdfPathname) {
		return
	}

	targetIndexOldVolume := emulator.GetFloppySoundVolumeDisk(targetIndexInt)
	emulator.SetFloppySoundVolumeDisk(targetIndexInt, shared.FLOPPY_DISK_IN_DRIVE_SOUND_VOLUME)

	if !attachAdf(targetIndexInt, foundAdfPathname) {
		emulator.SetFloppySoundVolumeDisk(targetIndexInt, targetIndexOldVolume)
	}
}

// DF medium - detach rom from source DF<index>, find other rom by name pattern (filename part)
// and attach it to target DF<index>
//
// target_index as -1 means (unspecified), so it will be source_index
func dfEjectFromSourceIndex(sourceIndex string) {
	if sourceIndex == "N" {
		dfEjectFromSourceIndexAll(sourceIndex)
		return
	}

	sourceIndexInt, _ := utils.StringUtilsInstance.StringToInt(sourceIndex, 10, 16)

	if sourceIndexInt > shared.MAX_ADFS-1 {
		return
	}

	sourceIndexAdf := emulator.GetAdf(sourceIndexInt)

	if sourceIndexAdf == "" {
		// ADF not attached at index
		return
	}

	targetIndexOldVolume := emulator.GetFloppySoundVolumeDisk(sourceIndexInt)
	emulator.SetFloppySoundVolumeDisk(sourceIndexInt, 0)

	if !detachAdf(sourceIndexInt, sourceIndexAdf) {
		emulator.SetFloppySoundVolumeDisk(sourceIndexInt, targetIndexOldVolume)
		return
	}
}

func dfEjectFromSourceIndexAll(sourceIndex string) {
	// TODO
	panic("unimplemented")
}

func isAdfAttached(adfPathname string) bool {
	for i := 0; i < shared.MAX_ADFS; i++ {
		if emulator.GetAdf(i) == adfPathname {
			return true
		}
	}

	return false
}

func findSimilarROMFile(mountpoint string, filenamePattern string) string {
	filenamePattern = utils.StringUtilsInstance.StringUnify(filenamePattern)
	filenamePattern = strings.ReplaceAll(filenamePattern, " ", ".*")
	filenamePattern = ".*" + filenamePattern + ".*"
	filenamePattern = strings.ToUpper(filenamePattern)

	filenamePatternRegEx := regexp.MustCompile(filenamePattern)

	for _, iRomPathname := range mountpointFiles[mountpoint] {
		iRomBasename := path.Base(iRomPathname)
		iRomBasename = adfBasenameCleanDiskOf(iRomBasename)

		iRomBasenameUnified := utils.StringUtilsInstance.StringUnify(iRomBasename)
		iRomBasenameUnified = strings.ToUpper(iRomBasenameUnified)

		if filenamePatternRegEx.MatchString(iRomBasenameUnified) {
			return iRomPathname
		}
	}

	return ""
}

func dfInsertFromSourceIndexToManyIndex(filenamePart, sourceIndex string) {
	// TODO
	panic("unimplemented")
}

func keyEventCallback(sender any, key string, pressed bool) {
	if isSoftResetKeys() {
		clearAllKeyboardsControl()

		utils.UnixUtilsInstance.Sync()
		emulator.SoftReset()
	} else if isHardResetKeys() {
		clearAllKeyboardsControl()

		// TODO detach and unmount all mediums
		utils.UnixUtilsInstance.Sync()
		emulator.HardReset()
	} else if isToggleAutoHeightKeys() {
		clearAllKeyboardsControl()

		emulator.ToggleAutoHeight()
	} else if isReleasedKey(shared.KEY_ESC) {
		clearAllKeyboardsControl()
	} else {
		if keyboardCommand := getKeyboardCommand(); keyboardCommand != "" {
			clearAllKeyboardsControl()
			processKeyboardCommand(keyboardCommand)
		}
	}
}

func getMountpointFirstFile(mounpoint string, extension string) string {
	for _, pathname := range mountpointFiles[mounpoint] {
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
		index = shared.DRIVE_INDEX_UNSPECIFIED
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

func mountpointIsMounted(mountpoint string) bool {
	for _, iMountpoint := range mounted {
		if iMountpoint == mountpoint {
			return true
		}
	}

	return false
}

func fixMountMedium(devicePathname, label, fsType string, dfIndex int, hdIndex int, cdIndex int) (string, error) {
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

	if mountpointIsMounted(target) {
		return "", fmt.Errorf("%v already mounted, unmount first", target)
	}

	if err := os.MkdirAll(target, 0777); err != nil {
		return "", err
	}

	if err := syscall.Mount(
		devicePathname,
		target,
		fsType,
		syscall.MS_SYNCHRONOUS,
		"flush"); err != nil {
		return "", err
	}

	mounted[devicePathname] = target

	if dfIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		dfIndexMountpoint[dfIndex] = target
	}

	if hdIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		hdIndexMountpoint[hdIndex] = target
	}

	if cdIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		cdIndexMountpoint[cdIndex] = target
	}

	return target, nil
}

func unmountMedium(devicePathname string, mountpoint string, flags int, dfIndex int, hdIndex int, cdIndex int) {
	log.Println("Unmount", mountpoint)

	syscall.Unmount(mountpoint, flags)
	delete(mounted, devicePathname)

	if dfIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		delete(dfIndexMountpoint, dfIndex)
	}

	if hdIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		delete(hdIndexMountpoint, hdIndex)
	}

	if cdIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		delete(cdIndexMountpoint, cdIndex)
	}
}

func loadMediumConfig(devicePathname string, mountpoint string) error {
	mediumConfig[devicePathname] = nil

	cfg, err := ini.Load(
		filepath.Join(mountpoint, shared.MEDIUM_CONFIG_INI_NAME))

	if err != nil {
		log.Println(devicePathname, mountpoint, "medium config does not exists")
		return err
	}

	mediumConfig[devicePathname] = cfg

	return nil
}

func getMediumDefaultFile(devicePathname string) string {
	cfg, exists := mediumConfig[devicePathname]

	if !exists || cfg == nil {
		return ""
	}

	if !cfg.HasSection(shared.MEDIUM_CONFIG_DEFAULT_SECTION) {
		return ""
	}

	if !cfg.Section(shared.MEDIUM_CONFIG_DEFAULT_SECTION).HasKey(
		shared.MEDIUM_CONFIG_DEFAULT_FILE) {
		return ""
	}

	filename := cfg.Section(shared.MEDIUM_CONFIG_DEFAULT_SECTION).Key(
		shared.MEDIUM_CONFIG_DEFAULT_FILE).String()

	fullPathname := filepath.Join(mounted[devicePathname], filename)

	stat, err := os.Stat(fullPathname)

	if err != nil || stat.IsDir() {
		return ""
	}

	return fullPathname
}

func attachDFMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	var err error
	var firstAdfPathname string

	index, _, err := parseMediumLabel(label, shared.AP4_MEDIUM_DF_RE)

	if err != nil || index == shared.DRIVE_INDEX_UNSPECIFIED {
		log.Println(path, label, "cannot get index for medium: ", err)

		return
	}

	if emulator.GetAdf(index) != "" {
		log.Printf("ADF already attached at DF%v, eject it first\n", index)

		return
	}

	if mountpoint != "" {
		unmountMedium(
			path,
			mountpoint,
			syscall.MNT_DETACH,
			index,
			shared.DRIVE_INDEX_UNSPECIFIED,
			shared.DRIVE_INDEX_UNSPECIFIED)
		mountpoint = ""
	}

	mountpoint, err = fixMountMedium(
		path,
		label,
		fsType,
		index,
		shared.DRIVE_INDEX_UNSPECIFIED,
		shared.DRIVE_INDEX_UNSPECIFIED)

	if err != nil {
		log.Println(path, label, err)

		return
	}

	loadMountpointFiles(mountpoint, []string{shared.FLOPPY_ADF_FULL_EXTENSION})
	loadMediumConfig(path, mountpoint)

	firstAdfPathname = getMediumDefaultFile(path)

	if firstAdfPathname == "" {
		// find first .adf file and attach it to the emulator
		firstAdfPathname = getMountpointFirstFile(mountpoint, shared.FLOPPY_ADF_FULL_EXTENSION)
	}

	if firstAdfPathname == "" {
		log.Println(path, label, "contains no", shared.FLOPPY_ADF_EXTENSION, "files")

		unmountMedium(
			path,
			mountpoint,
			syscall.MNT_DETACH,
			index,
			shared.DRIVE_INDEX_UNSPECIFIED,
			shared.DRIVE_INDEX_UNSPECIFIED)

		return
	}

	oldVolume := emulator.GetFloppySoundVolumeDisk(index)
	emulator.SetFloppySoundVolumeDisk(index, shared.FLOPPY_DISK_IN_DRIVE_SOUND_VOLUME)

	// TODO unmount when attachAdf return false and the medium
	// was not mounted prevoiusly mountpoint == "", so it means
	// the file was not attached
	if !attachAdf(index, firstAdfPathname) {
		unmountMedium(
			path,
			mountpoint,
			syscall.MNT_DETACH,
			index,
			shared.DRIVE_INDEX_UNSPECIFIED,
			shared.DRIVE_INDEX_UNSPECIFIED)

		emulator.SetFloppySoundVolumeDisk(index, oldVolume)
	}
}

func loadMountpointFiles(mountpoint string, extensions []string) {
	files := utils.FileUtilsInstance.GetDirFiles(mountpoint, false, extensions...)

	sort.Strings(files)

	mountpointFiles[mountpoint] = files
}

func attachDHMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	var err error

	index, bootPriority, err := parseMediumLabel(label, shared.AP4_MEDIUM_DH_RE)

	if err != nil || index == shared.DRIVE_INDEX_UNSPECIFIED {
		log.Println(path, label, "cannot get index for medium: ", err)

		return
	}

	if emulator.GetHd(index) != "" {
		log.Printf("HDF already attached at DH%v, eject it first\n", index)

		return
	}

	// TODO unmount all
	if mountpoint != "" {
		unmountMedium(
			path,
			mountpoint,
			syscall.MNT_DETACH,
			shared.DRIVE_INDEX_UNSPECIFIED,
			index,
			shared.DRIVE_INDEX_UNSPECIFIED)
		mountpoint = ""
	}

	mountpoint, err = fixMountMedium(
		path,
		label,
		fsType,
		shared.DRIVE_INDEX_UNSPECIFIED,
		index,
		shared.DRIVE_INDEX_UNSPECIFIED)

	if err != nil {
		log.Println(path, label, err)

		return
	}

	// TODO unmount when attachHdf return false and the medium
	// was not mounted prevoiusly mountpoint == "", so it means
	// the file was not attached
	if !attachHdDir(index, bootPriority, mountpoint) {
		unmountMedium(
			path,
			mountpoint,
			syscall.MNT_DETACH,
			shared.DRIVE_INDEX_UNSPECIFIED,
			index,
			shared.DRIVE_INDEX_UNSPECIFIED)
	}
}

func attachHFMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	var err error
	var firstHdfPathname string

	index, bootPriority, err := parseMediumLabel(label, shared.AP4_MEDIUM_HF_RE)

	if err != nil || index == shared.DRIVE_INDEX_UNSPECIFIED {
		log.Println(path, label, "cannot get index for medium: ", err)

		return
	}

	if emulator.GetHd(index) != "" {
		log.Printf("HDF already attached at DH%v, eject it first\n", index)

		return
	}

	// TODO unmount all
	if mountpoint != "" {
		unmountMedium(
			path,
			mountpoint,
			syscall.MNT_DETACH,
			shared.DRIVE_INDEX_UNSPECIFIED,
			index,
			shared.DRIVE_INDEX_UNSPECIFIED)
		mountpoint = ""
	}

	mountpoint, err = fixMountMedium(
		path,
		label,
		fsType,
		shared.DRIVE_INDEX_UNSPECIFIED,
		index,
		shared.DRIVE_INDEX_UNSPECIFIED)

	if err != nil {
		log.Println(path, label, err)

		return
	}

	loadMountpointFiles(mountpoint, []string{shared.HD_HDF_FULL_EXTENSION})
	loadMediumConfig(path, mountpoint)

	firstHdfPathname = getMediumDefaultFile(path)

	if firstHdfPathname == "" {
		// find first .hdf file and attach it to the emulator
		firstHdfPathname = getMountpointFirstFile(mountpoint, shared.HD_HDF_FULL_EXTENSION)
	}

	if firstHdfPathname == "" {
		log.Println(path, label, "contains no", shared.HD_HDF_EXTENSION, "files")

		unmountMedium(
			path,
			mountpoint,
			syscall.MNT_DETACH,
			shared.DRIVE_INDEX_UNSPECIFIED,
			index,
			shared.DRIVE_INDEX_UNSPECIFIED)

		return
	}

	// TODO unmount when attachHdf return false and the medium
	// was not mounted prevoiusly mountpoint == "", so it means
	// the file was not attached
	if !attachHdf(index, bootPriority, firstHdfPathname) {
		unmountMedium(
			path,
			mountpoint,
			syscall.MNT_DETACH,
			shared.DRIVE_INDEX_UNSPECIFIED,
			index,
			shared.DRIVE_INDEX_UNSPECIFIED)
	}
}

func attachCDMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	var err error
	var firstIsoPathname string

	index, _, err := parseMediumLabel(label, shared.AP4_MEDIUM_CD_RE)

	if err != nil || index == shared.DRIVE_INDEX_UNSPECIFIED {
		log.Println(path, label, "cannot get index for medium: ", err)

		return
	}

	if emulator.GetIso(index) != "" {
		log.Printf("ISO already attached at CD%v, eject it first\n", index)

		return
	}

	if mountpoint != "" {
		unmountMedium(
			path,
			mountpoint,
			syscall.MNT_DETACH,
			shared.DRIVE_INDEX_UNSPECIFIED,
			shared.DRIVE_INDEX_UNSPECIFIED,
			index)
		mountpoint = ""
	}

	mountpoint, err = fixMountMedium(
		path,
		label,
		fsType,
		shared.DRIVE_INDEX_UNSPECIFIED,
		shared.DRIVE_INDEX_UNSPECIFIED,
		index)

	if err != nil {
		log.Println(path, label, err)

		return
	}

	loadMountpointFiles(mountpoint, []string{shared.CD_ISO_FULL_EXTENSION})
	loadMediumConfig(path, mountpoint)

	firstIsoPathname = getMediumDefaultFile(path)

	if firstIsoPathname == "" {
		// find first .iso file and attach it to the emulator
		firstIsoPathname = getMountpointFirstFile(mountpoint, shared.CD_ISO_FULL_EXTENSION)
	}

	if firstIsoPathname == "" {
		log.Println(path, label, "contains no", shared.CD_ISO_EXTENSION, "files")

		unmountMedium(
			path,
			mountpoint,
			syscall.MNT_DETACH,
			shared.DRIVE_INDEX_UNSPECIFIED,
			shared.DRIVE_INDEX_UNSPECIFIED,
			index)

		return
	}

	// TODO unmount when attachIso return false and the medium
	// was not mounted prevoiusly mountpoint == "", so it means
	// the file was not attached
	if !attachIso(index, firstIsoPathname) {
		unmountMedium(
			path,
			mountpoint,
			syscall.MNT_DETACH,
			shared.DRIVE_INDEX_UNSPECIFIED,
			shared.DRIVE_INDEX_UNSPECIFIED,
			index)
	}
}

func attachMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	if shared.AP4_MEDIUM_DF_RE.MatchString(label) {
		attachDFMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
	} else if shared.AP4_MEDIUM_DH_RE.MatchString(label) {
		attachDHMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
	} else if shared.AP4_MEDIUM_HF_RE.MatchString(label) {
		attachHFMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
	} else if shared.AP4_MEDIUM_CD_RE.MatchString(label) {
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

	index, _, err := parseMediumLabel(label, shared.AP4_MEDIUM_DF_RE)

	if err != nil || index == shared.DRIVE_INDEX_UNSPECIFIED {
		log.Println(path, label, "cannot get index for medium: ", err)

		return
	}

	for i := 0; i < shared.MAX_ADFS; i++ {
		adfPathname := emulator.GetAdf(i)

		if adfPathname == "" {
			continue
		}

		// TODO check with _mountpoint + '/'
		if !strings.HasPrefix(adfPathname, _mountpoint) {
			continue
		}

		oldVolume := emulator.GetFloppySoundVolumeDisk(i)
		emulator.SetFloppySoundVolumeDisk(i, 0)

		if !detachAdf(i, adfPathname) {
			emulator.SetFloppySoundVolumeDisk(i, oldVolume)
		}
	}

	unmountMedium(
		path,
		_mountpoint,
		syscall.MNT_DETACH,
		index,
		shared.DRIVE_INDEX_UNSPECIFIED,
		shared.DRIVE_INDEX_UNSPECIFIED)
}

// One function for both HDF files and directories
func detachHDMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	_mountpoint, exists := mounted[path]

	if !exists || _mountpoint == "" {
		log.Println(path, label, "not mounted")

		return
	}

	index, _, err := parseMediumLabel(label, shared.AP4_MEDIUM_DH_RE)

	if err != nil || index == shared.DRIVE_INDEX_UNSPECIFIED {
		index, _, err = parseMediumLabel(label, shared.AP4_MEDIUM_HF_RE)

		if err != nil || index == shared.DRIVE_INDEX_UNSPECIFIED {
			log.Println(path, label, "cannot get index for medium: ", err)

			return
		}
	}

	for i := 0; i < shared.MAX_HDFS; i++ {
		hdfPathname := emulator.GetHd(i)

		if hdfPathname == "" {
			continue
		}

		// TODO check with _mountpoint + '/'
		if !strings.HasPrefix(hdfPathname, _mountpoint) {
			continue
		}

		detachHd(i, hdfPathname)
	}

	unmountMedium(
		path,
		_mountpoint,
		syscall.MNT_DETACH,
		shared.DRIVE_INDEX_UNSPECIFIED,
		index,
		shared.DRIVE_INDEX_UNSPECIFIED)
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

	index, _, err := parseMediumLabel(label, shared.AP4_MEDIUM_CD_RE)

	if err != nil || index == shared.DRIVE_INDEX_UNSPECIFIED {
		log.Println(path, label, "cannot get index for medium: ", err)

		return
	}

	for i := 0; i < shared.MAX_CDS; i++ {
		hdfPathname := emulator.GetIso(i)

		if hdfPathname == "" {
			continue
		}

		// TODO check with _mountpoint + '/'
		if !strings.HasPrefix(hdfPathname, _mountpoint) {
			continue
		}

		detachIso(i, hdfPathname)
	}

	unmountMedium(
		path,
		_mountpoint,
		syscall.MNT_DETACH,
		shared.DRIVE_INDEX_UNSPECIFIED,
		shared.DRIVE_INDEX_UNSPECIFIED,
		index)
}

func detachMediumDiskImage(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	if shared.AP4_MEDIUM_DF_RE.MatchString(label) {
		detachDFMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
	} else if shared.AP4_MEDIUM_DH_RE.MatchString(label) {
		detachHDMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
	} else if shared.AP4_MEDIUM_HF_RE.MatchString(label) {
		detachHDMediumDiskImage(name, size, _type, mountpoint, label, path, fsType, ptType, readOnly)
	} else if shared.AP4_MEDIUM_CD_RE.MatchString(label) {
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

func unmountAll() {
	utils.UnixUtilsInstance.Sync()

	for devicePathname, mountpoint := range mounted {
		unmountMedium(
			devicePathname,
			mountpoint,
			syscall.MNT_DETACH,
			shared.DRIVE_INDEX_UNSPECIFIED,
			shared.DRIVE_INDEX_UNSPECIFIED,
			shared.DRIVE_INDEX_UNSPECIFIED)
	}

	dfIndexMountpoint = make(map[int]string)
	hdIndexMountpoint = make(map[int]string)
	cdIndexMountpoint = make(map[int]string)
}

func stopServices() {
	amigaDiskDevicesDiscovery.Stop(&amigaDiskDevicesDiscovery)
	allKeyboardsControl.Stop(&allKeyboardsControl)
	emulator.Stop(&emulator)
	commander.Stop(&commander)
	blockDevices.Stop(&blockDevices)
}

func gracefulShutdown() {
	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan

	unmountAll()
	stopServices()
}

func adfBasenameCleanDiskOf(basename string) string {
	return shared.ADF_REMOVE_OF_NO_RE.ReplaceAllString(basename, "($1)")
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

	runnersBlocker.AddRunner(&amigaDiskDevicesDiscovery)
	runnersBlocker.AddRunner(&allKeyboardsControl)
	runnersBlocker.AddRunner(&emulator)
	runnersBlocker.AddRunner(&commander)
	runnersBlocker.AddRunner(&blockDevices)

	go gracefulShutdown()

	runnersBlocker.BlockUntilRunning()
}
