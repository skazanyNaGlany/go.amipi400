package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
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
var powerLEDControl components.PowerLEDControl
var numLockLEDControl components.NumLockLEDControl
var allKeyboardsControl components.AllKeyboardsControl
var amigaDiskDevicesDiscovery components_amipi400.AmigaDiskDevicesDiscovery
var emulator components_amipi400.AmiberryEmulator
var driveDevicesDiscovery components.DriveDevicesDiscovery
var commander components_amipi400.AmiberryCommander
var blockDevices components.BlockDevices
var mountpoints = components_amipi400.NewMountpointList()
var mainConfig = components_amipi400.NewMainConfig(shared.MAIN_CONFIG_INI_PATHNAME)
var initializing = true

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

	volume := 0

	if !amigaDiskDevicesDiscovery.HasFile(pathname) {
		volume = shared.FLOPPY_DISK_IN_DRIVE_SOUND_VOLUME
	}

	log.Println("Attaching", pathname, "to DF"+strIndex)

	emulator.AttachAdf(index, pathname, volume, 0)

	return true
}

func attachIso(index int, pathname string) bool {
	// TODO add support for CUE and NRG files
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

	emulator.DetachAdf(index, 0, 0)

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

	attachAdf(index, pathname)
}

func attachAmigaDiskDeviceIso(pathname string) {
	index := isoPathnameToCDIndex(pathname)

	attachIso(index, pathname)
}

func detachAmigaDiskDeviceAdf(pathname string) {
	index := adfPathnameToDFIndex(pathname)

	detachAdf(index, pathname)
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

	onHDOperationStart()
	defer onHDOperationDone()

	attachHdf(index, shared.DH_BOOT_PRIORITY_DEFAULT, pathname)
}

func detachAmigaDiskDeviceHdf(pathname string) {
	index := getHdfSlot(pathname)

	if index == shared.DRIVE_INDEX_UNSPECIFIED {
		log.Println("HDF", pathname, "not attached")

		return
	}

	onHDOperationStart()
	defer onHDOperationDone()

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

func servicesIdleCallback(sender any) {
	if !initializing {
		return
	}

	if !amigaDiskDevicesDiscovery.IsIdle() {
		return
	}

	if !blockDevices.IsIdle() {
		return
	}

	initializing = false

	log.Println("All services idle, running emulator")

	emulator.SetRerunEmulator(true)
	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)
}

func isSoftResetKeys() bool {
	return allKeyboardsControl.IsKeysReleasedAgo(shared.SOFT_RESET_KEYS, shared.SOFT_RESET_KEYS_MIN_MS)
}

func isHardResetKeys() bool {
	return allKeyboardsControl.IsKeysReleasedAgo(shared.HARD_RESET_KEYS, shared.HARD_RESET_KEYS_MIN_MS)
}

func isToggleZoomKeys() bool {
	return allKeyboardsControl.IsKeysReleasedAgo(shared.TOGGLE_ZOOM_KEYS, shared.TOGGLE_ZOOM_KEYS_MIN_MS)
}

func isShutdownKeys() bool {
	return allKeyboardsControl.IsKeysReleasedAgo(shared.SHUTDOWN_KEYS, shared.SHUTDOWN_KEYS_MIN_MS)
}

func getKeyboardCommand() string {
	releasedSequence := allKeyboardsControl.GetReleasedKeysSequenceAsString()
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
	allKeyboardsControl.ClearAll()
}

func processKeyboardCommand(keyboardCommand string) {
	if dfEjectRule := utils.RegExInstance.FindNamedMatches(
		shared.DF_EJECT_FROM_SOURCE_INDEX_RE,
		keyboardCommand); len(dfEjectRule) > 0 {
		// example: df0
		// example: dfn
		dfEjectFromSourceIndex(dfEjectRule["source_index"])
	} else if dfSourceByDiskRule := utils.RegExInstance.FindNamedMatches(
		shared.DF_INSERT_FROM_SOURCE_INDEX_BY_DISK_NO_RE,
		keyboardCommand); len(dfSourceByDiskRule) > 0 {
		// example: df02
		dfInsertFromSourceIndexToTargetIndexByDiskNo(
			dfSourceByDiskRule["disk_no"],
			dfSourceByDiskRule["source_index"],
			shared.DRIVE_INDEX_UNSPECIFIED_STR)
	} else if dfSourceTargetByDiskRule := utils.RegExInstance.FindNamedMatches(
		shared.DF_INSERT_FROM_SOURCE_TO_TARGET_INDEX_BY_DISK_NO_RE,
		keyboardCommand); len(dfSourceTargetByDiskRule) > 0 {
		// example: df02df1
		dfInsertFromSourceIndexToTargetIndexByDiskNo(
			dfSourceTargetByDiskRule["disk_no"],
			dfSourceTargetByDiskRule["source_index"],
			dfSourceTargetByDiskRule["target_index"])
	} else if dfSourceTargetRule := utils.RegExInstance.FindNamedMatches(
		shared.DF_INSERT_FROM_SOURCE_TO_TARGET_INDEX_RE,
		keyboardCommand); len(dfSourceTargetRule) > 0 {
		// example: df0kwaterdf1
		// example: df0kwaterdfn
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
	} else if cdEjectRule := utils.RegExInstance.FindNamedMatches(
		shared.CD_EJECT_FROM_SOURCE_INDEX_RE,
		keyboardCommand); len(cdEjectRule) > 0 {
		// example: cd0
		cdEjectFromSourceIndex(cdEjectRule["source_index"])
	} else if cdSourceRule := utils.RegExInstance.FindNamedMatches(
		shared.CD_INSERT_FROM_SOURCE_INDEX_RE,
		keyboardCommand); len(cdSourceRule) > 0 {
		// example: cd0workbenchiso
		cdInsertFromSourceIndexToTargetIndex(
			cdSourceRule["filename_part"],
			cdSourceRule["source_index"],
			shared.DRIVE_INDEX_UNSPECIFIED_STR)
	} else if hfEjectRule := utils.RegExInstance.FindNamedMatches(
		shared.HF_EJECT_FROM_SOURCE_INDEX_RE,
		keyboardCommand); len(hfEjectRule) > 0 {
		// example: hf0
		hfEjectFromSourceIndex(hfEjectRule["source_index"])
	} else if hfSourceRule := utils.RegExInstance.FindNamedMatches(
		shared.HF_INSERT_FROM_SOURCE_INDEX_RE,
		keyboardCommand); len(hfSourceRule) > 0 {
		// example: hf0workbenchhdf
		hfInsertFromSourceIndexToTargetIndex(
			hfSourceRule["filename_part"],
			hfSourceRule["source_index"],
			shared.DRIVE_INDEX_UNSPECIFIED_STR)
	} else if shared.UNMOUNT_ALL_RE.MatchString(keyboardCommand) {
		// example: u
		unmountAll(false)
	} else if dfUnmountRule := utils.RegExInstance.FindNamedMatches(
		shared.DF_UNMOUNT_FROM_SOURCE_INDEX_RE,
		keyboardCommand); len(dfUnmountRule) > 0 {
		// example: udf0
		// example: udfn
		dfUnmountFromSourceIndex(dfUnmountRule["source_index"])
	} else if cdUnmountRule := utils.RegExInstance.FindNamedMatches(
		shared.CD_UNMOUNT_FROM_SOURCE_INDEX_RE,
		keyboardCommand); len(cdUnmountRule) > 0 {
		// example: ucd0
		// example: ucdn
		cdUnmountFromSourceIndex(cdUnmountRule["source_index"])
	} else if hfUnmountRule := utils.RegExInstance.FindNamedMatches(
		shared.HF_UNMOUNT_FROM_SOURCE_INDEX_RE,
		keyboardCommand); len(hfUnmountRule) > 0 {
		// example: uhf0
		// example: uhfn
		hfUnmountFromSourceIndex(hfUnmountRule["source_index"])
	} else if dhUnmountRule := utils.RegExInstance.FindNamedMatches(
		shared.DH_UNMOUNT_FROM_SOURCE_INDEX_RE,
		keyboardCommand); len(dhUnmountRule) > 0 {
		// example: udh0
		// example: udhn
		dhUnmountFromSourceIndex(dhUnmountRule["source_index"])
	} else if lowLevelCopyRule := utils.RegExInstance.FindNamedMatches(
		shared.LOW_LEVEL_COPY_RE,
		keyboardCommand); len(lowLevelCopyRule) > 0 {
		// example: cdf0dh1
		lowLevelCopy(
			lowLevelCopyRule["source_low_level_device"],
			lowLevelCopyRule["source_index"],
			lowLevelCopyRule["target_low_level_device"],
			lowLevelCopyRule["target_index"])
	}
}

func fillIndexes(indexStr string, maxIndex int) []int {
	indexes := make([]int, 0)

	if indexStr == "N" {
		for index := 0; index < maxIndex; index++ {
			indexes = append(indexes, index)
		}
	} else {
		indexInt, _ := utils.StringUtilsInstance.StringToInt(indexStr, 10, 16)

		if indexInt > shared.MAX_ADFS-1 {
			return indexes
		}

		indexes = append(indexes, indexInt)
	}

	return indexes
}

func dfUnmountFromSourceIndex(sourceIndex string) {
	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)

	sourceIndexes := fillIndexes(sourceIndex, shared.MAX_ADFS)
	countFailed := 0

	for _, index := range sourceIndexes {
		mountpoint := mountpoints.GetMountpointByDFIndex(index)

		if mountpoint == nil {
			continue
		}

		detachDFMountpointROMs(mountpoint)

		if !unmountMountpoint(mountpoint, true) {
			countFailed++
		}
	}

	if countFailed > 0 {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
	}
}

func cdUnmountFromSourceIndex(sourceIndex string) {
	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)

	sourceIndexes := fillIndexes(sourceIndex, shared.MAX_CDS)
	countFailed := 0

	for _, index := range sourceIndexes {
		mountpoint := mountpoints.GetMountpointByCDIndex(index)

		if mountpoint == nil {
			continue
		}

		detachCDMountpointROMs(mountpoint)

		if !unmountMountpoint(mountpoint, true) {
			countFailed++
		}
	}

	if countFailed > 0 {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
	}
}

func hfUnmountFromSourceIndex(sourceIndex string) {
	dhUnmountFromSourceIndex(sourceIndex)
}

func dhUnmountFromSourceIndex(sourceIndex string) {
	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)

	onHDOperationStart()
	defer onHDOperationDone()

	sourceIndexes := fillIndexes(sourceIndex, shared.MAX_HDFS)
	countFailed := 0

	for _, index := range sourceIndexes {
		mountpoint := mountpoints.GetMountpointByDHIndex(index)

		if mountpoint == nil {
			continue
		}

		detachDHMountpointROMs(mountpoint)

		if !unmountMountpoint(mountpoint, true) {
			countFailed++
		}
	}

	if countFailed > 0 {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
	}
}

func lowLevelCopy(
	sourceLowLevelDevice string,
	sourceIndex string,
	targetLowLevelDevice string,
	targetIndex string) {
	var sourceMountpoint *components_amipi400.Mountpoint
	var targetMountpoint *components_amipi400.Mountpoint
	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)

	// TODO move to consts.go
	supportedDevices := []string{shared.LOW_LEVEL_DEVICE_FLOPPY, shared.LOW_LEVEL_DEVICE_HARD_DISK}

	sourceIndexInt, _ := utils.StringUtilsInstance.StringToInt(sourceIndex, 10, 16)
	targetIndexInt, _ := utils.StringUtilsInstance.StringToInt(targetIndex, 10, 16)

	if !funk.ContainsString(supportedDevices, sourceLowLevelDevice) {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if !funk.ContainsString(supportedDevices, targetLowLevelDevice) {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	// valiate source/target indexes

	// source
	if sourceLowLevelDevice == shared.LOW_LEVEL_DEVICE_FLOPPY {
		if sourceIndexInt > shared.MAX_ADFS-1 {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	} else if sourceLowLevelDevice == shared.LOW_LEVEL_DEVICE_HARD_DISK {
		if sourceIndexInt > shared.MAX_HDFS-1 {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	// target
	if targetLowLevelDevice == shared.LOW_LEVEL_DEVICE_FLOPPY {
		if targetIndexInt > shared.MAX_ADFS-1 {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	} else if sourceLowLevelDevice == shared.LOW_LEVEL_DEVICE_HARD_DISK {
		if targetIndexInt > shared.MAX_HDFS-1 {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	if sourceLowLevelDevice == targetLowLevelDevice {
		if sourceIndexInt == targetIndexInt {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	// validation almost done, next:
	// get source/target pathnames
	// for DH - check if the user wants to copy a directory, if does error
	// detach
	// pause the emulator
	// copy
	// resume the emulator
	// attach again

	if sourceLowLevelDevice == shared.LOW_LEVEL_DEVICE_HARD_DISK || targetLowLevelDevice == shared.LOW_LEVEL_DEVICE_HARD_DISK {
		onHDOperationStart()
		defer onHDOperationDone()
	}

	sourcePathname := ""
	targetPathname := ""

	// get and detach both

	// source
	if sourceLowLevelDevice == shared.LOW_LEVEL_DEVICE_FLOPPY {
		sourcePathname = emulator.GetAdf(sourceIndexInt)

		if sourcePathname == "" {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		if !detachAdf(sourceIndexInt, sourcePathname) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		sourceMountpoint = mountpoints.GetMountpointByDFIndex(sourceIndexInt)

		if sourceMountpoint == nil {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	} else if sourceLowLevelDevice == shared.LOW_LEVEL_DEVICE_HARD_DISK {
		sourcePathname = emulator.GetHd(sourceIndexInt)

		if sourcePathname == "" {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		sourceIsHdf := strings.HasSuffix(sourcePathname, shared.HD_HDF_FULL_EXTENSION)

		if !sourceIsHdf {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		if !detachHd(sourceIndexInt, sourcePathname) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		sourceMountpoint = mountpoints.GetMountpointByDHIndex(sourceIndexInt)

		if sourceMountpoint == nil {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	// target
	if targetLowLevelDevice == shared.LOW_LEVEL_DEVICE_FLOPPY {
		targetPathname = emulator.GetAdf(targetIndexInt)

		if targetPathname == "" {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		if !detachAdf(targetIndexInt, targetPathname) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		targetMountpoint = mountpoints.GetMountpointByDFIndex(targetIndexInt)

		if targetMountpoint == nil {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	} else if targetLowLevelDevice == shared.LOW_LEVEL_DEVICE_HARD_DISK {
		targetPathname = emulator.GetHd(targetIndexInt)

		if targetPathname == "" {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		targetIsHdf := strings.HasSuffix(targetPathname, shared.HD_HDF_FULL_EXTENSION)

		if !targetIsHdf {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		if !detachHd(targetIndexInt, targetPathname) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		targetMountpoint = mountpoints.GetMountpointByDHIndex(targetIndexInt)

		if targetMountpoint == nil {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	log.Println("Low-level copying", sourcePathname, "to", targetPathname)

	utils.UnixUtilsInstance.Sync()

	// attach both
	// source
	if sourceLowLevelDevice == shared.LOW_LEVEL_DEVICE_FLOPPY {
		if !attachAdf(sourceIndexInt, sourcePathname) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	} else if sourceLowLevelDevice == shared.LOW_LEVEL_DEVICE_HARD_DISK {
		if !attachHdf(sourceIndexInt, sourceMountpoint.DHBootPriority, sourcePathname) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	// target
	if targetLowLevelDevice == shared.LOW_LEVEL_DEVICE_FLOPPY {
		if !attachAdf(targetIndexInt, targetPathname) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	} else if targetLowLevelDevice == shared.LOW_LEVEL_DEVICE_HARD_DISK {
		if !attachHdf(targetIndexInt, targetMountpoint.DHBootPriority, targetPathname) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	utils.UnixUtilsInstance.Sync()
}

func dfInsertFromSourceIndexToTargetIndexByDiskNo(diskNo, sourceIndex, targetIndex string) {
	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)

	diskNoInt, _ := utils.StringUtilsInstance.StringToInt(diskNo, 10, 16)
	sourceIndexInt, _ := utils.StringUtilsInstance.StringToInt(sourceIndex, 10, 16)
	targetIndexInt, _ := utils.StringUtilsInstance.StringToInt(targetIndex, 10, 16)

	if targetIndexInt == shared.DRIVE_INDEX_UNSPECIFIED {
		targetIndexInt = sourceIndexInt
	}

	if sourceIndexInt > shared.MAX_ADFS-1 || targetIndexInt > shared.MAX_ADFS-1 {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	mountpoint := mountpoints.GetMountpointByDFIndex(sourceIndexInt)

	if mountpoint == nil {
		// allow to manage only these drives mounted by amipi400.go
		// so skip these from amiga_disk_devices.go
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	sourceIndexAdf := emulator.GetAdf(sourceIndexInt)
	targetIndexAdf := emulator.GetAdf(targetIndexInt)

	if targetIndexAdf != "" {
		if amigaDiskDevicesDiscovery.HasFile(targetIndexAdf) {
			// ADF attached by amiga_disk_devices.go
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		if !detachAdf(targetIndexInt, targetIndexAdf) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	if sourceIndexAdf == "" {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	foundAdfPathnames := findSimilarROMFiles(mountpoint, sourceIndexAdf)
	lenFoundAdfPathnames := len(foundAdfPathnames)
	toInsertPathname := ""

	if lenFoundAdfPathnames == 0 {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	requiredDiskNoOfMax := fmt.Sprintf(shared.ADF_DISK_NO_OF_MAX, diskNoInt, lenFoundAdfPathnames)

	for _, pathname := range foundAdfPathnames {
		if strings.Contains(pathname, requiredDiskNoOfMax) {
			toInsertPathname = pathname
			break
		}
	}

	if toInsertPathname == "" {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if attachedIndex := isAdfAttached(toInsertPathname); attachedIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		if !detachAdf(attachedIndex, toInsertPathname) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	if !attachAdf(targetIndexInt, toInsertPathname) {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}
}

func dfInsertFromSourceIndexToTargetIndex(filenamePart, sourceIndex, targetIndex string) {
	if targetIndex == "N" {
		dfInsertFromSourceIndexToManyIndex(filenamePart, sourceIndex)
		return
	}

	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)

	filenamePart = strings.TrimSpace(filenamePart)
	sourceIndexInt, _ := utils.StringUtilsInstance.StringToInt(sourceIndex, 10, 16)
	targetIndexInt, _ := utils.StringUtilsInstance.StringToInt(targetIndex, 10, 16)

	if filenamePart == "" {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if targetIndexInt == shared.DRIVE_INDEX_UNSPECIFIED {
		targetIndexInt = sourceIndexInt
	}

	if sourceIndexInt > shared.MAX_ADFS-1 || targetIndexInt > shared.MAX_ADFS-1 {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	mountpoint := mountpoints.GetMountpointByDFIndex(sourceIndexInt)

	if mountpoint == nil {
		// allow to manage only these drives mounted by amipi400.go
		// so skip these from amiga_disk_devices.go
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	targetIndexAdf := emulator.GetAdf(targetIndexInt)

	if targetIndexAdf != "" {
		if amigaDiskDevicesDiscovery.HasFile(targetIndexAdf) {
			// ADF attached by amiga_disk_devices.go
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		if !detachAdf(targetIndexInt, targetIndexAdf) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	foundAdfPathname := findSimilarROMFile(mountpoint, filenamePart)

	if foundAdfPathname == "" {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if attachedIndex := isAdfAttached(foundAdfPathname); attachedIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		if !detachAdf(attachedIndex, foundAdfPathname) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	if !attachAdf(targetIndexInt, foundAdfPathname) {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}
}

func hfInsertFromSourceIndexToTargetIndex(filenamePart, sourceIndex, targetIndex string) {
	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)

	filenamePart = strings.TrimSpace(filenamePart)
	sourceIndexInt, _ := utils.StringUtilsInstance.StringToInt(sourceIndex, 10, 16)
	targetIndexInt, _ := utils.StringUtilsInstance.StringToInt(targetIndex, 10, 16)

	if filenamePart == "" {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if targetIndexInt == shared.DRIVE_INDEX_UNSPECIFIED {
		targetIndexInt = sourceIndexInt
	}

	if sourceIndexInt > shared.MAX_HDFS-1 || targetIndexInt > shared.MAX_HDFS-1 {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	mountpoint := mountpoints.GetMountpointByDHIndex(sourceIndexInt)

	if mountpoint == nil {
		// allow to manage only these drives mounted by amipi400.go
		// so skip these from amiga_disk_devices.go
		return
	}

	matched := utils.RegExInstance.FindNamedMatches(
		shared.AP4_MEDIUM_HF_RE,
		mountpoint.Label)
	isHdLabel := len(matched) != 0

	if !isHdLabel {
		// source medium is not HF (perhaps DH), cannot use it
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	onHDOperationStart()
	defer onHDOperationDone()

	targetIndexHdf := emulator.GetHd(targetIndexInt)

	if targetIndexHdf != "" {
		if amigaDiskDevicesDiscovery.HasFile(targetIndexHdf) {
			// HDF attached by amiga_disk_devices.go
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		if !detachHd(targetIndexInt, targetIndexHdf) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	foundHdfPathname := findSimilarROMFile(mountpoint, filenamePart)

	if foundHdfPathname == "" {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if attachedIndex := isHdfAttached(foundHdfPathname); attachedIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		if !detachHd(attachedIndex, foundHdfPathname) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	if !attachHdf(
		targetIndexInt,
		mountpoint.DHBootPriority,
		foundHdfPathname) {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}
}

func cdInsertFromSourceIndexToTargetIndex(filenamePart, sourceIndex, targetIndex string) {
	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)

	// TODO test me - check if ISO attached by this method works in the emulator
	// eg. by using emulating CD32 (use cd32.uae.template config)
	filenamePart = strings.TrimSpace(filenamePart)
	sourceIndexInt, _ := utils.StringUtilsInstance.StringToInt(sourceIndex, 10, 16)
	targetIndexInt, _ := utils.StringUtilsInstance.StringToInt(targetIndex, 10, 16)

	if filenamePart == "" {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if targetIndexInt == shared.DRIVE_INDEX_UNSPECIFIED {
		targetIndexInt = sourceIndexInt
	}

	if sourceIndexInt > shared.MAX_CDS-1 || targetIndexInt > shared.MAX_CDS-1 {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	mountpoint := mountpoints.GetMountpointByCDIndex(sourceIndexInt)

	targetIndexIso := emulator.GetIso(targetIndexInt)

	if targetIndexIso != "" {
		if amigaDiskDevicesDiscovery.HasFile(targetIndexIso) {
			// ISO attached by amiga_disk_devices.go
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		if !detachIso(targetIndexInt, targetIndexIso) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	foundIsoPathname := findSimilarROMFile(mountpoint, filenamePart)

	if foundIsoPathname == "" {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if attachedIndex := isIsoAttached(foundIsoPathname); attachedIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		if !detachIso(attachedIndex, foundIsoPathname) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}

	if !attachIso(targetIndexInt, foundIsoPathname) {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}
}

func dfEjectFromSourceIndex(sourceIndex string) {
	if sourceIndex == "N" {
		dfEjectFromSourceIndexAll()
		return
	}

	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)

	sourceIndexInt, _ := utils.StringUtilsInstance.StringToInt(sourceIndex, 10, 16)

	if sourceIndexInt > shared.MAX_ADFS-1 {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	sourceIndexAdf := emulator.GetAdf(sourceIndexInt)

	if sourceIndexAdf == "" {
		// ADF not attached at index
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if amigaDiskDevicesDiscovery.HasFile(sourceIndexAdf) {
		// ADF attached by amiga_disk_devices.go
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if !detachAdf(sourceIndexInt, sourceIndexAdf) {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}
}

func cdEjectFromSourceIndex(sourceIndex string) {
	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)

	sourceIndexInt, _ := utils.StringUtilsInstance.StringToInt(sourceIndex, 10, 16)

	if sourceIndexInt > shared.MAX_CDS-1 {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	sourceIndexIso := emulator.GetIso(sourceIndexInt)

	if sourceIndexIso == "" {
		// ISO not attached at index
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if amigaDiskDevicesDiscovery.HasFile(sourceIndexIso) {
		// ISO attached by amiga_disk_devices.go
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if !detachIso(sourceIndexInt, sourceIndexIso) {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}
}

func hfEjectFromSourceIndex(sourceIndex string) {
	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)

	sourceIndexInt, _ := utils.StringUtilsInstance.StringToInt(sourceIndex, 10, 16)

	if sourceIndexInt > shared.MAX_HDFS-1 {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	sourceIndexHdf := emulator.GetHd(sourceIndexInt)

	if sourceIndexHdf == "" {
		// HDF not attached at index
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if amigaDiskDevicesDiscovery.HasFile(sourceIndexHdf) {
		// HDF attached by amiga_disk_devices.go
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if strings.HasSuffix(sourceIndexHdf, "/") {
		// DH is not HDF file but directory, cannot detach
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	onHDOperationStart()
	defer onHDOperationDone()

	if !detachHd(sourceIndexInt, sourceIndexHdf) {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}
}

func dfEjectFromSourceIndexAll() {
	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)

	for index := 0; index < shared.MAX_ADFS; index++ {
		adfPathname := emulator.GetAdf(index)

		if adfPathname == "" {
			continue
		}

		if amigaDiskDevicesDiscovery.HasFile(adfPathname) {
			// ADF attached by amiga_disk_devices.go
			continue
		}

		if !detachAdf(index, adfPathname) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}
}

func isAdfAttached(adfPathname string) int {
	for i := 0; i < shared.MAX_ADFS; i++ {
		if emulator.GetAdf(i) == adfPathname {
			return i
		}
	}

	return shared.DRIVE_INDEX_UNSPECIFIED
}

func isIsoAttached(isoPathname string) int {
	for i := 0; i < shared.MAX_CDS; i++ {
		if emulator.GetIso(i) == isoPathname {
			return i
		}
	}

	return shared.DRIVE_INDEX_UNSPECIFIED
}

func isHdfAttached(hdfPathname string) int {
	for i := 0; i < shared.MAX_CDS; i++ {
		if emulator.GetHd(i) == hdfPathname {
			return i
		}
	}

	return shared.DRIVE_INDEX_UNSPECIFIED
}

func findSimilarROMFiles(mountpoint *components_amipi400.Mountpoint, pathname string) []string {
	basename := path.Base(pathname)
	extension := path.Ext(basename)

	diskNoStrPart := utils.RegExInstance.FindNamedMatches(shared.ADF_DISK_NO_OF_MAX_RE, basename)

	if len(diskNoStrPart) == 0 {
		// there is no (Disk NO of MAX) in the ADF name
		return []string{pathname}
	}

	noDiskSignFilename := strings.Replace(basename, diskNoStrPart["disk_no_of_max"], "", 1)
	noExtFilename := strings.TrimSuffix(noDiskSignFilename, extension)

	similar := make([]string, 0)

	for _, iPathname := range mountpoint.Files {
		iPathnameBasename := path.Base(iPathname)

		if !strings.HasPrefix(iPathnameBasename, noExtFilename) {
			continue
		}

		if !shared.ADF_DISK_NO_OF_MAX_RE.MatchString(iPathnameBasename) {
			continue
		}

		if !strings.HasSuffix(iPathnameBasename, extension) {
			continue
		}

		if !funk.ContainsString(similar, iPathnameBasename) {
			similar = append(similar, iPathname)
		}
	}

	return similar
}

func findSimilarROMFile(mountpoint *components_amipi400.Mountpoint, filenamePattern string) string {
	if !mountpoint.HasFiles() {
		return ""
	}

	for _, iRomPathname := range mountpoint.Files {
		iRomBasenme := path.Base(iRomPathname)

		iRomUnified := utils.StringUtilsInstance.StringUnify(iRomBasenme)
		iRomUnified = strings.ToUpper(iRomUnified)

		if strings.Contains(iRomUnified, filenamePattern) {
			return iRomPathname
		}

		iRomParts := strings.Split(iRomUnified, " ")
		filenamePatternCopy := filenamePattern

		for _, iRomPart := range iRomParts {
			if strings.Index(filenamePatternCopy, iRomPart) == 0 {
				filenamePatternCopy = strings.Replace(filenamePatternCopy, iRomPart, "", 1)
				filenamePatternCopy = strings.TrimSpace(filenamePatternCopy)

				if filenamePatternCopy == "" {
					return iRomPathname
				}
			}
		}
	}

	return ""
}

func dfInsertFromSourceIndexToManyIndex(filenamePart, sourceIndex string) {
	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)

	filenamePart = strings.TrimSpace(filenamePart)
	sourceIndexInt, _ := utils.StringUtilsInstance.StringToInt(sourceIndex, 10, 16)

	if filenamePart == "" {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	if sourceIndexInt > shared.MAX_ADFS-1 {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	mountpoint := mountpoints.GetMountpointByDFIndex(sourceIndexInt)

	if mountpoint == nil {
		// allow to manage only these drives mounted by amipi400.go
		// so skip these from amiga_disk_devices.go
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	// find first ADF by pattern typed by the user
	foundAdfPathname := findSimilarROMFile(mountpoint, filenamePart)

	if foundAdfPathname == "" {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	// find similar ADFs by the first ADF and attach
	// them to the emulator
	foundAdfPathnames := findSimilarROMFiles(mountpoint, foundAdfPathname)
	lenFoundAdfPathnames := len(foundAdfPathnames)

	if lenFoundAdfPathnames == 0 {
		numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		return
	}

	for targetIndexInt, pathname := range foundAdfPathnames {
		if targetIndexInt+1 > shared.MAX_ADFS {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}

		targetIndexAdf := emulator.GetAdf(targetIndexInt)

		if targetIndexAdf != "" {
			if amigaDiskDevicesDiscovery.HasFile(targetIndexAdf) {
				// ADF attached by amiga_disk_devices.go
				continue
			}

			if !detachAdf(targetIndexInt, targetIndexAdf) {
				numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
				return
			}
		}

		if !attachAdf(targetIndexInt, pathname) {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
			return
		}
	}
}

func keyEventCallback(sender any, key string, pressed bool) {
	if isSoftResetKeys() {
		clearAllKeyboardsControl()

		utils.UnixUtilsInstance.Sync()
		emulator.SoftReset()
	} else if isHardResetKeys() {
		clearAllKeyboardsControl()

		utils.UnixUtilsInstance.Sync()
		emulator.HardReset()
	} else if isToggleZoomKeys() {
		clearAllKeyboardsControl()

		emulator.ToggleZoom()

		saveZoomConfigSetting()
	} else if isShutdownKeys() {
		clearAllKeyboardsControl()

		utils.UnixUtilsInstance.Sync()
		utils.UnixUtilsInstance.Shutdown()
	} else if allKeyboardsControl.IsKeysReleasedAgo(
		shared.CLEAR_BUFFER_KEYS,
		shared.CLEAR_COMMAND_BUFFER_MIN_MS) {
		clearAllKeyboardsControl()
	} else if diskNo := isReplaceDFByIndexShortcut(); diskNo != shared.DISK_INDEX_UNSPECIFIED {
		clearAllKeyboardsControl()

		dfInsertFromSourceIndexToTargetIndexByDiskNo(
			fmt.Sprint(diskNo),
			"0",
			shared.DRIVE_INDEX_UNSPECIFIED_STR)
	} else {
		if keyboardCommand := getKeyboardCommand(); keyboardCommand != "" {
			clearAllKeyboardsControl()
			processKeyboardCommand(keyboardCommand)
		} else {
			emulateNumPad()
		}
	}
}

func emulateNumPad() {
	if allKeyboardsControl.IsKeysReleasedAgo(
		shared.NUMPAD_EMULATE_ENTER_KEYS,
		shared.NUMPAD_EMULATE_MIN_MS) {
		// emulate ENTER on the numpad
		allKeyboardsControl.WriteOnce(shared.KEY_R_ENTER)
	}
}

func saveZoomConfigSetting() {
	mainConfig.AmiPi400.Zoom = emulator.IsZoom()

	if err := mainConfig.Save(); err != nil {
		log.Println(err)
	}
}

func isReplaceDFByIndexShortcut() int {
	var err error

	releasedKeys := allKeyboardsControl.GetReleasedKeys()

	currentTimestamp := time.Now().UnixMilli()
	goodCount := 0
	diskNo := int64(shared.DISK_INDEX_UNSPECIFIED)

	for key, pressedTimestamp := range releasedKeys {
		pressedTimestampChange := currentTimestamp - pressedTimestamp

		if pressedTimestampChange < 0 || pressedTimestampChange > 1000 {
			continue
		}

		if key == shared.KEY_LEFTMETA {
			goodCount++
		} else {
			diskNo, err = strconv.ParseInt(key, 10, 16)

			if err == nil {
				goodCount++
			}
		}
	}

	if goodCount != 2 {
		return shared.DISK_INDEX_UNSPECIFIED
	}

	return int(diskNo)
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

func setupAddMountpoint(
	devicePathname string,
	label string,
	fsType string,
	dfIndex int,
	dhIndex int,
	cdIndex int,
	dhBootPriority int,
	extensions []string) (*components_amipi400.Mountpoint, error) {
	mountpointStr := filepath.Join(shared.AP4_ROOT_MOUNTPOINT, label)

	if mountpoints.HasMountpoint(mountpointStr) {
		return nil, fmt.Errorf("%v already mounted, unmount first", mountpointStr)
	}

	if err := os.MkdirAll(mountpointStr, 0777); err != nil {
		return nil, fmt.Errorf("cannot create directory for mountpoint %v (%v, %v)", mountpointStr, devicePathname, err)
	}

	mountpoint := components_amipi400.NewMountpoint(devicePathname, mountpointStr, label, fsType)

	log.Printf("Fix %v\n", mountpoint.DevicePathname)
	mountpoint.Fix()

	utils.UnixUtilsInstance.Sync()

	log.Printf("Mount %v as %v (%v)\n", mountpoint.DevicePathname, mountpoint.Mountpoint, mountpoint.FsType)

	if err := mountpoint.Mount(); err != nil {
		return nil, fmt.Errorf("cannot mount mountpoint %v (%v, %v)", mountpointStr, devicePathname, err)
	}

	if extensions != nil {
		mountpoint.LoadFiles(extensions)
	}

	if err := mountpoint.LoadConfig(); err != nil {
		log.Printf("Cannot load medium config for %v: %v\n", mountpoint.Mountpoint, err)
	}

	if extensions != nil && !mountpoint.HasFiles() && mountpoint.Config.AmiPi400.DefaultFile != shared.MEDIUM_CONFIG_DEFAULT_FILE_NONE {
		unmountMountpoint(mountpoint, false)

		return nil, fmt.Errorf("%v contains no %v files", label, strings.Join(extensions, ","))
	}

	if dfIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		mountpoint.DFIndex = dfIndex
	}

	if dhIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		mountpoint.DHIndex = dhIndex
		mountpoint.DHBootPriority = dhBootPriority
	}

	if cdIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		mountpoint.CDIndex = cdIndex
	}

	mountpoints.AddMountpoint(mountpoint)

	return mountpoint, nil
}

func unmountMountpoint(mountpoint *components_amipi400.Mountpoint, remove bool) bool {
	var err error

	log.Printf("Unmount %v from %v (%v)\n", mountpoint.DevicePathname, mountpoint.Mountpoint, mountpoint.FsType)

	// try to unmount a mountpoint for 16 times
	for i := 0; i < 8; i++ {
		utils.UnixUtilsInstance.Sync()

		err = mountpoint.Unmount()

		if err == nil {
			break
		}

		time.Sleep(time.Second * 1)
	}

	if err != nil {
		log.Printf("Cannot unmount %v: %v\n", mountpoint.Mountpoint, err)

		return false
	}

	if remove {
		mountpoints.RemoveMountpoint(mountpoint)
	}

	return true
}

func attachDFMediumDiskImage(
	name string,
	size uint64,
	_type, mountpointStr, label, path, fsType, ptType string,
	readOnly bool) {

	index, _, err := parseMediumLabel(label, shared.AP4_MEDIUM_DF_RE)

	if err != nil || index == shared.DRIVE_INDEX_UNSPECIFIED {
		log.Println(path, label, "cannot get index for medium: ", err)
		return
	}

	if emulator.GetAdf(index) != "" {
		log.Printf("ADF already attached at DF%v, eject it first\n", index)
		return
	}

	if mountpointStr != "" {
		unmountMountpoint(
			components_amipi400.NewMountpoint(path, mountpointStr, "", ""),
			false)
	}

	mountpoint, err := setupAddMountpoint(
		path,
		label,
		fsType,
		index,
		shared.DRIVE_INDEX_UNSPECIFIED,
		shared.DRIVE_INDEX_UNSPECIFIED,
		0,
		[]string{shared.FLOPPY_ADF_FULL_EXTENSION})

	if err != nil {
		log.Println(err)
		return
	}

	if mountpoint.Config.AmiPi400.DefaultFile == shared.MEDIUM_CONFIG_DEFAULT_FILE_NONE {
		return
	}

	if !attachAdf(index, mountpoint.Config.AmiPi400.DefaultFile) {
		unmountMountpoint(mountpoint, true)
		return
	}
}

func attachDHMediumDiskImage(
	name string,
	size uint64,
	_type, mountpointStr, label, path, fsType, ptType string,
	readOnly bool) {
	index, bootPriority, err := parseMediumLabel(label, shared.AP4_MEDIUM_DH_RE)

	if err != nil || index == shared.DRIVE_INDEX_UNSPECIFIED {
		log.Println(path, label, "cannot get index for medium: ", err)
		return
	}

	if emulator.GetHd(index) != "" {
		log.Printf("HDF already attached at DH%v, eject it first\n", index)
		return
	}

	if mountpointStr != "" {
		unmountMountpoint(
			components_amipi400.NewMountpoint(path, mountpointStr, "", ""),
			false)
	}

	mountpoint, err := setupAddMountpoint(
		path,
		label,
		fsType,
		shared.DRIVE_INDEX_UNSPECIFIED,
		index,
		shared.DRIVE_INDEX_UNSPECIFIED,
		bootPriority,
		nil)

	if err != nil {
		log.Println(err)
		return
	}

	onHDOperationStart()
	defer onHDOperationDone()

	if !attachHdDir(index, bootPriority, mountpoint.Mountpoint) {
		unmountMountpoint(mountpoint, true)
		return
	}
}

func attachHFMediumDiskImage(
	name string,
	size uint64,
	_type, mountpointStr, label, path, fsType, ptType string,
	readOnly bool) {
	index, bootPriority, err := parseMediumLabel(label, shared.AP4_MEDIUM_HF_RE)

	if err != nil || index == shared.DRIVE_INDEX_UNSPECIFIED {
		log.Println(path, label, "cannot get index for medium: ", err)
		return
	}

	if emulator.GetHd(index) != "" {
		log.Printf("HDF already attached at DH%v, eject it first\n", index)
		return
	}

	if mountpointStr != "" {
		unmountMountpoint(
			components_amipi400.NewMountpoint(path, mountpointStr, "", ""), false)
	}

	mountpoint, err := setupAddMountpoint(
		path,
		label,
		fsType,
		shared.DRIVE_INDEX_UNSPECIFIED,
		index,
		shared.DRIVE_INDEX_UNSPECIFIED,
		bootPriority,
		[]string{shared.HD_HDF_FULL_EXTENSION})

	if err != nil {
		log.Println(err)
		return
	}

	if mountpoint.Config.AmiPi400.DefaultFile == shared.MEDIUM_CONFIG_DEFAULT_FILE_NONE {
		return
	}

	onHDOperationStart()
	defer onHDOperationDone()

	if !attachHdf(index, bootPriority, mountpoint.Config.AmiPi400.DefaultFile) {
		unmountMountpoint(mountpoint, true)
		return
	}
}

func attachCDMediumDiskImage(
	name string,
	size uint64,
	_type, mountpointStr, label, path, fsType, ptType string,
	readOnly bool) {
	index, _, err := parseMediumLabel(label, shared.AP4_MEDIUM_CD_RE)

	if err != nil || index == shared.DRIVE_INDEX_UNSPECIFIED {
		log.Println(path, label, "cannot get index for medium: ", err)
		return
	}

	if emulator.GetIso(index) != "" {
		log.Printf("ISO already attached at CD%v, eject it first\n", index)
		return
	}

	if mountpointStr != "" {
		unmountMountpoint(
			components_amipi400.NewMountpoint(path, mountpointStr, "", ""),
			false)
	}

	mountpoint, err := setupAddMountpoint(
		path,
		label,
		fsType,
		shared.DRIVE_INDEX_UNSPECIFIED,
		shared.DRIVE_INDEX_UNSPECIFIED,
		index,
		shared.DH_BOOT_PRIORITY_UNSPECIFIED,
		[]string{shared.CD_ISO_FULL_EXTENSION})

	if err != nil {
		log.Println(err)
		return
	}

	if mountpoint.Config.AmiPi400.DefaultFile == shared.MEDIUM_CONFIG_DEFAULT_FILE_NONE {
		return
	}

	if !attachIso(index, mountpoint.Config.AmiPi400.DefaultFile) {
		unmountMountpoint(mountpoint, true)
		return
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
	_type, mountpointStr, label, path, fsType, ptType string,
	readOnly bool) {
	mountpoint := mountpoints.GetMountpointByDevicePathname(path)

	if mountpoint == nil {
		log.Println(path, label, "not mounted")
		return
	}

	if mountpoint.DFIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		adfPathname := emulator.GetAdf(mountpoint.DFIndex)

		if adfPathname != "" {
			detachAdf(mountpoint.DFIndex, adfPathname)
		}
	}

	unmountMountpoint(mountpoint, true)
}

// One function for both HDF files and directories
func detachHDMediumDiskImage(
	name string,
	size uint64,
	_type, mountpointStr, label, path, fsType, ptType string,
	readOnly bool) {
	mountpoint := mountpoints.GetMountpointByDevicePathname(path)

	if mountpoint == nil {
		log.Println(path, label, "not mounted")
		return
	}

	onHDOperationStart()
	defer onHDOperationDone()

	if mountpoint.DHIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		hdfPathname := emulator.GetHd(mountpoint.DHIndex)

		if hdfPathname != "" {
			detachHd(mountpoint.DHIndex, hdfPathname)
		}
	}

	unmountMountpoint(mountpoint, true)
}

func detachCDMediumDiskImage(
	name string,
	size uint64,
	_type, mountpointStr, label, path, fsType, ptType string,
	readOnly bool) {
	mountpoint := mountpoints.GetMountpointByDevicePathname(path)

	if mountpoint == nil {
		log.Println(path, label, "not mounted")
		return
	}

	if mountpoint.CDIndex != shared.DRIVE_INDEX_UNSPECIFIED {
		hdfPathname := emulator.GetIso(mountpoint.CDIndex)

		if hdfPathname != "" {
			detachIso(mountpoint.CDIndex, hdfPathname)
		}
	}

	unmountMountpoint(mountpoint, true)
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

func detachDFMountpointROMs(mountpoint *components_amipi400.Mountpoint) {
	for index := 0; index < shared.MAX_ADFS; index++ {
		pathname := emulator.GetAdf(index)

		if pathname == "" {
			continue
		}

		if funk.ContainsString(mountpoint.Files, pathname) {
			detachAdf(index, pathname)
		}
	}
}

func detachDHMountpointROMs(mountpoint *components_amipi400.Mountpoint) {
	for index := 0; index < shared.MAX_HDFS; index++ {
		pathname := emulator.GetHd(index)

		if pathname == "" {
			continue
		}

		// can be a file
		if funk.ContainsString(mountpoint.Files, pathname) {
			detachHd(index, pathname)
		}

		// or a directory
		if mountpoint.Mountpoint == pathname {
			detachHd(index, pathname)
		}
	}
}

func detachCDMountpointROMs(mountpoint *components_amipi400.Mountpoint) {
	for index := 0; index < shared.MAX_CDS; index++ {
		pathname := emulator.GetIso(index)

		if pathname == "" {
			continue
		}

		if funk.ContainsString(mountpoint.Files, pathname) {
			detachIso(index, pathname)
		}
	}
}

func detachMountpointROMs(mountpoint *components_amipi400.Mountpoint) {
	detachDFMountpointROMs(mountpoint)
	detachDHMountpointROMs(mountpoint)
	detachCDMountpointROMs(mountpoint)
}

func onHDOperationStart() {
	if initializing {
		return
	}

	emulator.SetRerunEmulator(false)
}

func onHDOperationDone() {
	if initializing {
		return
	}

	emulator.SetRerunEmulator(true)
}

func unmountAll(fromSignal bool) {
	if !fromSignal {
		onHDOperationStart()
		defer onHDOperationDone()

		powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
		defer powerLEDControl.BlinkPowerLEDSecs(shared.CMD_SUCCESS_BLINK_POWER_SECS)
	}

	emulator.HardReset()

	log.Println("Unmounting all mountpoints...")

	utils.UnixUtilsInstance.Sync()

	// try to unmount all mountpoints for 8 times
	for i := 0; i < 8; i++ {
		for _, mountpoint := range mountpoints.Mountpoints {
			detachMountpointROMs(mountpoint)
			unmountMountpoint(mountpoint, true)
		}

		if len(mountpoints.Mountpoints) == 0 {
			break
		}

		log.Println("Retrying...")

		time.Sleep(time.Second * 1)
	}

	if len(mountpoints.Mountpoints) > 0 {
		if !fromSignal {
			numLockLEDControl.BlinkNumLockLEDSecs(shared.CMD_FAILURE_BLINK_NUM_LOCK_SECS)
		}
	}

	log.Println("Done unmounting mountpoints")
}

func stopServices() {
	amigaDiskDevicesDiscovery.Stop(&amigaDiskDevicesDiscovery)
	allKeyboardsControl.Stop(&allKeyboardsControl)
	emulator.Stop(&emulator)
	commander.Stop(&commander)
	blockDevices.Stop(&blockDevices)
	powerLEDControl.Stop(&powerLEDControl)
	numLockLEDControl.Stop(&numLockLEDControl)
}

func gracefulShutdown() {
	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan

	unmountAll(true)
	stopServices()
}

func main() {
	utils.GoUtilsInstance.CheckPlatform()
	utils.UnixUtilsInstance.CheckForRoot()
	utils.SysUtilsInstance.CheckForExecutables(
		shared.AMIPI400_NEEDED_EXECUTABLES)

	mainConfig.Load()

	// this will save config file
	// with default values if not exists
	mainConfig.Save()

	exeDir := utils.GoUtilsInstance.MustCwdToExeOrScript()
	logFilename := utils.GoUtilsInstance.MustDuplicateLog(exeDir)

	log.Printf("%v v%v\n", shared.AMIPI400_UNIXNAME, shared.AMIPI400_VERSION)
	log.Printf("Executable directory %v\n", exeDir)
	log.Printf("Log filename %v\n", logFilename)

	log.Println("Waiting for all services to became idle...")

	// start blinking of the power LED and wait for amigaDiskDevicesDiscovery
	// and blockDevices to became idle, before running the emulator
	// servicesIdleCallback will disable blinking of the power LED and
	// unlock the emulator by SetRerunEmulator(true)
	powerLEDControl.EnablePowerLed()
	powerLEDControl.BlinkPowerLEDSecs(shared.CMD_PENDING_BLINK_POWER_SECS)
	emulator.SetRerunEmulator(false)

	amigaDiskDevicesDiscovery.SetAttachedAmigaDiskDeviceCallback(attachedAmigaDiskDeviceCallback)
	amigaDiskDevicesDiscovery.SetDetachedAmigaDiskDeviceCallback(detachedAmigaDiskDeviceCallback)
	amigaDiskDevicesDiscovery.SetMountpoint(shared.FILE_SYSTEM_MOUNT)
	amigaDiskDevicesDiscovery.SetIdleCallback(servicesIdleCallback)
	allKeyboardsControl.SetKeyEventCallback(keyEventCallback)
	commander.SetTmpIniPathname(shared.AMIBERRY_EMULATOR_TMP_INI_PATHNAME)
	emulator.SetExecutablePathname(shared.AMIBERRY_EXE_PATHNAME)
	emulator.SetConfigPathname(shared.AMIPI400_AMIBERRY_CONFIG_PATHNAME)
	emulator.SetAmiberryCommander(&commander)
	emulator.SetZoom(mainConfig.AmiPi400.Zoom)
	blockDevices.AddAttachedCallback(attachedBlockDeviceCallback)
	blockDevices.AddDetachedCallback(detachedBlockDeviceCallback)
	blockDevices.SetIdleCallback(servicesIdleCallback)

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
	powerLEDControl.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	powerLEDControl.SetDebugMode(shared.RUNNERS_DEBUG_MODE)
	numLockLEDControl.SetVerboseMode(shared.RUNNERS_VERBOSE_MODE)
	numLockLEDControl.SetDebugMode(shared.RUNNERS_DEBUG_MODE)

	amigaDiskDevicesDiscovery.Start(&amigaDiskDevicesDiscovery)
	allKeyboardsControl.Start(&allKeyboardsControl)
	emulator.Start(&emulator)
	commander.Start(&commander)
	blockDevices.Start(&blockDevices)
	powerLEDControl.Start(&powerLEDControl)
	numLockLEDControl.Start(&numLockLEDControl)

	runnersBlocker.AddRunner(&amigaDiskDevicesDiscovery)
	runnersBlocker.AddRunner(&allKeyboardsControl)
	runnersBlocker.AddRunner(&emulator)
	runnersBlocker.AddRunner(&commander)
	runnersBlocker.AddRunner(&blockDevices)
	runnersBlocker.AddRunner(&powerLEDControl)
	runnersBlocker.AddRunner(&numLockLEDControl)

	go gracefulShutdown()

	runnersBlocker.BlockUntilRunning()
}
