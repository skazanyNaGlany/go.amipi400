package components

import (
	"log"
	"time"

	"github.com/MarinX/keylogger"
	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"github.com/thoas/go-funk"
	"golang.org/x/exp/slices"
)

type KeyboardControl struct {
	RunnerBase

	keyboard          *keylogger.KeyLogger
	pressedKeys       []string
	keyEventCallbacks []interfaces.KeyEventCallback
}

func (kc *KeyboardControl) init() bool {
	var err error

	kc.pressedKeys = make([]string, 0)

	keyboardDevice := keylogger.FindKeyboardDevice()

	if keyboardDevice == "" {
		if kc.debugMode {
			log.Println("No keyboard found")

			kc.running = false
			return false
		}
	}

	kc.keyboard, err = keylogger.New(keyboardDevice)

	if err != nil {
		if kc.debugMode {
			log.Println(err)

			kc.running = false
			return false
		}
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

func (kc *KeyboardControl) GetPressedKeys() []string {
	return kc.pressedKeys
}

func (kc *KeyboardControl) ClearPressedKeys() {
	kc.pressedKeys = make([]string, 0)
}

func (kc *KeyboardControl) SetPressedKey(key string) bool {
	if !funk.ContainsString(kc.pressedKeys, key) {
		kc.pressedKeys = append(kc.pressedKeys, key)

		return true
	}

	return false
}

func (kc *KeyboardControl) ClearPressedKey(key string) bool {
	if index := funk.IndexOfString(kc.pressedKeys, key); index != -1 {
		kc.pressedKeys = slices.Delete(kc.pressedKeys, index, index+1)

		return true
	}

	return false
}

func (kc *KeyboardControl) IsKeyPressed(key string) bool {
	return funk.ContainsString(kc.pressedKeys, key)
}

func (kc *KeyboardControl) AddKeyEventCallback(callback interfaces.KeyEventCallback) {
	kc.keyEventCallbacks = append(kc.keyEventCallbacks, callback)
}

func (kc *KeyboardControl) callKeyEventCallbacks(key string, pressed bool) {
	for _, callback := range kc.keyEventCallbacks {
		callback(key, pressed)
	}
}
