package components

import (
	"log"
	"strings"

	"github.com/MarinX/keylogger"
	"github.com/skazanyNaGlany/go.amipi400/shared"
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
	devices := akc.findAllKeyboardDevices()

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

func (akc *AllKeyboardsControl) ClearPressedKeys() {
	for _, kc := range akc.keyboardControls {
		kc.ClearPressedKeys()
	}
}

func (akc *AllKeyboardsControl) GetReleasedKeysAgo(ms int64) map[string]int64 {
	all := make(map[string]int64)

	for _, kc := range akc.keyboardControls {
		releasedKeys := kc.GetReleasedKeysAgo(ms)

		for ikey, timestamp := range releasedKeys {
			all[ikey] = timestamp
		}
	}

	return all
}

func (akc *AllKeyboardsControl) ClearReleasedKeys() {
	for _, kc := range akc.keyboardControls {
		kc.ClearReleasedKeys()
	}
}

func (akc *AllKeyboardsControl) ClearKeysSequence() {
	for _, kc := range akc.keyboardControls {
		kc.ClearKeysSequence()
	}
}

func (akc *AllKeyboardsControl) GetKeysSequence() []KeySequence {
	all := make([]KeySequence, 0)

	for _, kc := range akc.keyboardControls {
		all = append(all, kc.GetKeysSequence()...)
	}

	if len(all) > shared.MAX_KEYS_SEQUENCE {
		return all[0:shared.MAX_KEYS_SEQUENCE]
	}

	return all
}

func (akc *AllKeyboardsControl) GetReleasedKeysSequenceAsString() []string {
	released := make([]string, 0)

	for _, ks := range akc.GetKeysSequence() {
		if ks.Pressed {
			continue
		}

		released = append(released, ks.Key)
	}

	return released
}

func (akc *AllKeyboardsControl) IsKeysPressed(keys []string) bool {
	for _, kc := range akc.keyboardControls {
		if kc.IsKeysPressed(keys) {
			return true
		}
	}

	return false
}

func (akc *AllKeyboardsControl) IsKeysReleasedAgo(keys []string, ms int64) bool {
	for _, kc := range akc.keyboardControls {
		if kc.IsKeysReleasedAgo(keys, ms) {
			return true
		}
	}

	return false
}

func (akc *AllKeyboardsControl) findAllKeyboardDevices() []string {
	devices := keylogger.FindAllKeyboardDevices()

	for _, pathname := range utils.FileUtilsInstance.GetDirFiles("/dev/input/by-path", false) {
		if !strings.HasSuffix(pathname, "-event-kbd") {
			continue
		}

		pathname, _ = realpath.Realpath(pathname)

		devices = append(devices, pathname)
	}

	return funk.UniqString(devices)
}

func (akc *AllKeyboardsControl) SetKeyEventCallback(
	callback interfaces.KeyEventCallback,
) {
	akc.keyEventCallback = callback
}

func (akc *AllKeyboardsControl) ClearAll() {
	akc.ClearPressedKeys()
	akc.ClearReleasedKeys()
	akc.ClearKeysSequence()
}

func (akc *AllKeyboardsControl) WriteOnce(key string) error {
	return akc.keyboardControls[0].WriteOnce(key)
}
