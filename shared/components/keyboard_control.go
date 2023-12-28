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
	125: shared.KEY_LEFTMETA,
}

type KeySequence struct {
	Key       string
	Timestamp int64
	Pressed   bool
}

type KeyboardControl struct {
	RunnerBase

	keyboard          *keylogger.KeyLogger
	pressedKeys       map[string]int64
	releasedKeys      map[string]int64
	keyEventCallbacks []interfaces.KeyEventCallback
	keyboardDevice    string
	keysSequence      []KeySequence
}

func (kc *KeyboardControl) init() bool {
	var err error

	kc.pressedKeys = make(map[string]int64)
	kc.releasedKeys = make(map[string]int64)
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
					if kc.SetPressedKey(keyStr) {
						kc.callKeyEventCallbacks(keyStr, true)
					}
				}

				if ievent.KeyRelease() {
					if kc.ClearPressedKey(keyStr) {
						kc.callKeyEventCallbacks(keyStr, false)
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

func (kc *KeyboardControl) GetPressedKeys() map[string]int64 {
	return kc.pressedKeys
}

func (kc *KeyboardControl) ClearPressedKeys() {
	kc.pressedKeys = make(map[string]int64)
}

func (kc *KeyboardControl) GetReleasedKeys() map[string]int64 {
	return kc.releasedKeys
}

func (kc *KeyboardControl) GetReleasedKeysAgo(ms int64) map[string]int64 {
	currentTimestamp := time.Now().UnixMilli()
	keys := make(map[string]int64)

	for key, timestamp := range kc.releasedKeys {
		if currentTimestamp-timestamp <= ms {
			keys[key] = timestamp
		}
	}

	return keys
}

func (kc *KeyboardControl) ClearReleasedKeys() {
	kc.releasedKeys = make(map[string]int64)
}

func (kc *KeyboardControl) AddKeySequence(key string, timestamp int64, pressed bool) {
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
		timestamp := time.Now().UnixMilli()

		kc.pressedKeys[key] = timestamp
		delete(kc.releasedKeys, key)

		kc.AddKeySequence(key, timestamp, true)

		return true
	}

	return false
}

func (kc *KeyboardControl) ClearPressedKey(key string) bool {
	if _, isPressed := kc.pressedKeys[key]; isPressed {
		oldTimestamp := kc.pressedKeys[key]

		delete(kc.pressedKeys, key)
		kc.releasedKeys[key] = oldTimestamp

		kc.AddKeySequence(key, time.Now().UnixMilli(), false)

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

func (kc *KeyboardControl) IsKeysPressedAgo(keys []string, ms int64) bool {
	pressedKeys := kc.GetPressedKeys()
	currentTimestamp := time.Now().UnixMilli()
	goodCount := 0

	for key, pressedTimestamp := range pressedKeys {
		if funk.ContainsString(keys, key) {
			if currentTimestamp-pressedTimestamp <= ms {
				goodCount++
			}
		}
	}

	return goodCount == len(keys)
}

func (kc *KeyboardControl) IsKeysReleased(keys []string) bool {
	count := 0

	for ikey := range kc.releasedKeys {
		if funk.ContainsString(keys, ikey) {
			count++
		}
	}

	return count == len(keys)
}

func (kc *KeyboardControl) IsKeysReleasedAgo(keys []string, ms int64) bool {
	releasedKeys := kc.GetReleasedKeys()
	currentTimestamp := time.Now().UnixMilli()
	goodCount := 0

	for key, releasedTimestamp := range releasedKeys {
		if funk.ContainsString(keys, key) {
			if currentTimestamp-releasedTimestamp <= ms {
				goodCount++
			}
		}
	}

	return goodCount == len(keys)
}

func (kc *KeyboardControl) AddKeyEventCallback(callback interfaces.KeyEventCallback) {
	kc.keyEventCallbacks = append(kc.keyEventCallbacks, callback)
}

func (kc *KeyboardControl) callKeyEventCallbacks(key string, pressed bool) {
	for _, callback := range kc.keyEventCallbacks {
		callback(kc, key, pressed)
	}
}

func (kc *KeyboardControl) SetKeyboardDevice(keyboardDevice string) {
	kc.keyboardDevice = keyboardDevice
}

func (kc *KeyboardControl) WriteOnce(key string) error {
	return kc.keyboard.WriteOnce(key)
}
