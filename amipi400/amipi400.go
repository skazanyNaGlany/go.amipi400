package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

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

func attachAmigaDiskDeviceAdf(pathname string) {
	index := adfPathnameToDFIndex(pathname)
	strIndex := fmt.Sprint(index)

	if emulator.GetAdf(index) != "" {
		log.Println("ADF already attached at DF" + strIndex + ", eject it first")

		return
	}

	log.Println("Attaching", pathname, "to DF"+strIndex)

	emulator.AttachAdf(index, pathname)
}

func detachAmigaDiskDeviceAdf(pathname string) {
	index := adfPathnameToDFIndex(pathname)
	strIndex := fmt.Sprint(index)

	currentAdfPathname := emulator.GetAdf(index)

	if currentAdfPathname == "" {
		log.Println("ADF not attached to DF" + strIndex + ", cannot eject")

		return
	}

	if currentAdfPathname != pathname {
		log.Println(pathname + " not attached to DF" + strIndex + ", cannot eject")

		return
	}

	log.Println("Detaching", pathname, "from DF"+strIndex)

	emulator.DetachAdf(index)
}

func attachedAmigaDiskDeviceCallback(pathname string) {
	isAdf := strings.HasSuffix(pathname, consts.FLOPPY_ADF_FULL_EXTENSION)

	if isAdf {
		attachAmigaDiskDeviceAdf(pathname)

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

	log.Fatalln(pathname, "not supported")
}

func keyEventCallback(sender any, key string, pressed bool) {
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

	for _, devicePathname := range cdroms {
		log.Println("\t" + devicePathname)
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

	amigaDiskDevicesDiscovery.Start(&amigaDiskDevicesDiscovery)
	allKeyboardsControl.Start(&allKeyboardsControl)
	emulator.Start(&emulator)
	commander.Start(&commander)

	defer amigaDiskDevicesDiscovery.Stop(&amigaDiskDevicesDiscovery)
	defer allKeyboardsControl.Stop(&allKeyboardsControl)
	defer emulator.Stop(&emulator)
	defer commander.Stop(&commander)

	runnersBlocker.AddRunner(&amigaDiskDevicesDiscovery)
	runnersBlocker.AddRunner(&allKeyboardsControl)
	runnersBlocker.AddRunner(&emulator)
	runnersBlocker.AddRunner(&commander)

	runnersBlocker.BlockUntilRunning()
}
