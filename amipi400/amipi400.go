package main

import (
	"log"

	"github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/components/utils"
	"github.com/skazanyNaGlany/go.amipi400/consts"
)

var runnersBlocker components.RunnersBlocker
var keyboardControls []*components.KeyboardControl

func keyEventCallback(sender any, key string, pressed bool) {
}

func initKeyboardControls() {
	kc := components.KeyboardControl{}
	devices := kc.FindAllKeyboardDevices()

	for _, idevice := range devices {
		_kc := &components.KeyboardControl{}

		_kc.SetKeyboardDevice(idevice)

		_kc.SetVerboseMode(consts.RUNNERS_VERBOSE_MODE)
		_kc.SetDebugMode(consts.RUNNERS_DEBUG_MODE)

		_kc.AddKeyEventCallback(keyEventCallback)

		keyboardControls = append(keyboardControls, _kc)
	}
}

func startKeyboardControls() {
	for _, kc := range keyboardControls {
		kc.Start(kc)
	}
}

func stopKeyboardControls() {
	for _, kc := range keyboardControls {
		kc.Stop(kc)
	}
}

func addKeyboardControlsRunners() {
	for _, kc := range keyboardControls {
		runnersBlocker.AddRunner(kc)
	}
}

func clearPressedKeys() {
	for _, kc := range keyboardControls {
		kc.ClearPressedKeys()
	}
}

func isKeysPressed(keys []string) bool {
	for _, kc := range keyboardControls {
		if kc.IsKeysPressed(keys) {
			return true
		}
	}

	return false
}

func main() {
	utils.GoUtilsInstance.CheckPlatform()
	utils.UnixUtilsInstance.CheckForRoot()

	exeDir := utils.GoUtilsInstance.MustCwdToExeOrScript()
	logFilename := utils.GoUtilsInstance.MustDuplicateLog(exeDir)

	log.Printf("%v v%v\n", consts.AMIPI400_UNIXNAME, consts.AMIPI400_VERSION)
	log.Printf("Executable directory %v\n", exeDir)
	log.Printf("Log filename %v\n", logFilename)

	initKeyboardControls()

	startKeyboardControls()

	defer stopKeyboardControls()

	addKeyboardControlsRunners()

	runnersBlocker.BlockUntilRunning()
}
