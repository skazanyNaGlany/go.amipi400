package components

import (
	"log"
	"time"

	"github.com/MarinX/keylogger"
	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/interfaces"
	"github.com/thoas/go-funk"
)

// key codes that does not exists in
// the github.com/MarinX/keylogger
var extendedKeyCodeMap = map[uint16]string{
	125: "KEY_LEFTMETA",
}

type KeySequence struct {
	Key       string
	Timestamp int
	Pressed   bool
}

type KeyboardControl struct {
	RunnerBase

	keyboard          *keylogger.KeyLogger
	pressedKeys       map[string]int
	keyEventCallbacks []interfaces.KeyEventCallback
	keyboardDevice    string
	keysSequence      []KeySequence
}

func (kc *KeyboardControl) init() bool {
	var err error

	kc.pressedKeys = make(map[string]int)
	kc.keysSequence = make([]KeySequence, 0)
	kc.keyboard, err = keylogger.New(kc.keyboardDevice)

	if err != nil {
		if kc.debugMode {
			log.Println(err)
		}

		kc.running = false
		return false
	}

	return true
}

func (kc *KeyboardControl) copyPressedKeys() map[string]int {
	_copy := make(map[string]int)

	for k, c := range kc.pressedKeys {
		_copy[k] = c
	}

	return _copy
}

func (kc *KeyboardControl) loop() {
	defer kc.keyboard.Close()

	for kc.running {
		time.Sleep(time.Millisecond * 10)

		events := kc.keyboard.Read()

		for ievent := range events {
			if ievent.Type == keylogger.EvKey {
				keyStr := ievent.KeyString()

				if keyStr == "" {
					keyStr = kc.guessUnknownKeyName(ievent.Code)
				}

				if ievent.KeyPress() {
					pressedKeysCopy := kc.copyPressedKeys()

					if kc.SetPressedKey(keyStr) {
						kc.callKeyEventCallbacks(
							keyStr,
							true,
							pressedKeysCopy,
							kc.pressedKeys)
					}
				}

				if ievent.KeyRelease() {
					pressedKeysCopy := kc.copyPressedKeys()

					if kc.ClearPressedKey(keyStr) {
						kc.callKeyEventCallbacks(
							keyStr,
							false,
							pressedKeysCopy,
							kc.pressedKeys)
					}
				}
			}
		}
	}

	kc.running = false
}

func (kc *KeyboardControl) guessUnknownKeyName(code uint16) string {
	keyStr, exists := extendedKeyCodeMap[code]

	if exists {
		return keyStr
	}

	return ""
}

func (kc *KeyboardControl) Run() {
	if !kc.init() {
		return
	}

	kc.loop()
}

func (kc *KeyboardControl) Stop(_runner interfaces.Runner) error {
	if kc.verboseMode {
		log.Printf("Stopping %T %p\n", _runner, &_runner)
	}

	kc.running = false

	if kc.keyboard != nil {
		if err := kc.keyboard.Close(); err != nil {
			if kc.debugMode {
				log.Println(err)
			}
		}
	}

	return nil
}

func (kc *KeyboardControl) GetPressedKeys() map[string]int {
	return kc.pressedKeys
}

func (kc *KeyboardControl) ClearPressedKeys() {
	kc.pressedKeys = make(map[string]int)
}

func (kc *KeyboardControl) AddKeySequence(key string, timestamp int, pressed bool) {
	kc.keysSequence = append(kc.keysSequence, KeySequence{
		Key:       key,
		Timestamp: timestamp,
		Pressed:   pressed})

	if len(kc.keysSequence) >= shared.MAX_KEYS_SEQUENCE {
		kc.keysSequence = kc.keysSequence[1:]
	}
}

func (kc *KeyboardControl) ClearKeysSequence() {
	kc.keysSequence = make([]KeySequence, 0)
}

func (kc *KeyboardControl) GetKeysSequence() []KeySequence {
	return kc.keysSequence
}

func (kc *KeyboardControl) SetPressedKey(key string) bool {
	if _, isPressed := kc.pressedKeys[key]; !isPressed {
		timestamp := time.Now().Nanosecond()

		kc.pressedKeys[key] = timestamp
		kc.AddKeySequence(key, timestamp, true)

		return true
	}

	return false
}

func (kc *KeyboardControl) ClearPressedKey(key string) bool {
	if _, isPressed := kc.pressedKeys[key]; isPressed {
		delete(kc.pressedKeys, key)
		kc.AddKeySequence(key, time.Now().Nanosecond(), false)

		return true
	}

	return false
}

func (kc *KeyboardControl) IsKeyPressed(key string) bool {
	_, isPressed := kc.pressedKeys[key]

	return isPressed
}

func (kc *KeyboardControl) IsKeysPressed(keys []string) bool {
	count := 0

	for ikey := range kc.pressedKeys {
		if funk.ContainsString(keys, ikey) {
			count++
		}
	}

	return count == len(keys)
}

func (kc *KeyboardControl) AddKeyEventCallback(callback interfaces.KeyEventCallback) {
	kc.keyEventCallbacks = append(kc.keyEventCallbacks, callback)
}

func (kc *KeyboardControl) callKeyEventCallbacks(
	key string,
	pressed bool,
	prevPressedKeys map[string]int,
	newPressedKeys map[string]int) {
	for _, callback := range kc.keyEventCallbacks {
		callback(kc, key, pressed, prevPressedKeys, newPressedKeys)
	}
}

func (kc *KeyboardControl) SetKeyboardDevice(keyboardDevice string) {
	kc.keyboardDevice = keyboardDevice
}
