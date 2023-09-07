package components

import (
	"log"
	"strings"

	"github.com/MarinX/keylogger"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
	"github.com/skazanyNaGlany/go.amipi400/shared/interfaces"
	"github.com/thoas/go-funk"
	"github.com/yookoala/realpath"
)

type AllKeyboardsControl struct {
	RunnerBase

	keyboardControls []*KeyboardControl
	keyEventCallback interfaces.KeyEventCallback
}

func (akc *AllKeyboardsControl) Run() {
	devices := akc.FindAllKeyboardDevices()

	for _, idevice := range devices {
		_kc := &KeyboardControl{}

		_kc.SetKeyboardDevice(idevice)

		_kc.SetVerboseMode(akc.IsVerboseMode())
		_kc.SetDebugMode(akc.IsDebugMode())

		if akc.keyEventCallback != nil {
			_kc.AddKeyEventCallback(akc.keyEventCallback)
		}

		_kc.Start(_kc)

		akc.keyboardControls = append(akc.keyboardControls, _kc)
	}
}

func (akc *AllKeyboardsControl) Stop(_runner interfaces.Runner) error {
	if akc.verboseMode {
		log.Printf("Stopping %T %p\n", _runner, &_runner)
	}

	for _, kc := range akc.keyboardControls {
		kc.Stop(kc)
	}

	akc.running = false

	return nil
}

func (akc *AllKeyboardsControl) IsRunning() bool {
	if !akc.running {
		return false
	}

	for _, kc := range akc.keyboardControls {
		if !kc.IsRunning() {
			return false
		}
	}

	return true
}

func (akc *AllKeyboardsControl) GetPressedKeys() map[string]int {
	all := make(map[string]int)

	for _, kc := range akc.keyboardControls {
		pressedKeys := kc.GetPressedKeys()

		for ikey, timestamp := range pressedKeys {
			all[ikey] = timestamp
		}
	}

	return all
}

func (akc *AllKeyboardsControl) ClearPressedKeys() {
	for _, kc := range akc.keyboardControls {
		kc.ClearPressedKeys()
	}
}

func (akc *AllKeyboardsControl) SetPressedKey(key string) bool {
	for _, kc := range akc.keyboardControls {
		kc.SetPressedKey(key)
	}

	return false
}

func (akc *AllKeyboardsControl) ClearPressedKey(key string) bool {
	for _, kc := range akc.keyboardControls {
		kc.ClearPressedKey(key)
	}

	return false
}

func (akc *AllKeyboardsControl) IsKeyPressed(key string) bool {
	for _, kc := range akc.keyboardControls {
		if kc.IsKeyPressed(key) {
			return true
		}
	}

	return false
}

func (akc *AllKeyboardsControl) IsKeysPressed(keys []string) bool {
	for _, kc := range akc.keyboardControls {
		if kc.IsKeysPressed(keys) {
			return true
		}
	}

	return false
}

func (akc *AllKeyboardsControl) FindAllKeyboardDevices() []string {
	devices := keylogger.FindAllKeyboardDevices()

	for _, pathname := range utils.FileUtilsInstance.GetDirFiles("/dev/input/by-path") {
		if !strings.HasSuffix(pathname, "-event-kbd") {
			continue
		}

		pathname, _ = realpath.Realpath(pathname)

		devices = append(devices, pathname)
	}

	return funk.UniqString(devices)
}

func (akc *AllKeyboardsControl) SetKeyEventCallback(callback interfaces.KeyEventCallback) {
	akc.keyEventCallback = callback
}
