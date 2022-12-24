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

	keyboard    *keylogger.KeyLogger
	pressedKeys []string
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
					if !funk.ContainsString(kc.pressedKeys, keyStr) {
						kc.pressedKeys = append(kc.pressedKeys, keyStr)
					}
				}

				if ievent.KeyRelease() {
					kc.ClearPressedKey(keyStr)
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

func (kc *KeyboardControl) ClearPressedKey(key string) {
	if index := funk.IndexOfString(kc.pressedKeys, key); index != -1 {
		kc.pressedKeys = slices.Delete(kc.pressedKeys, index, index+1)
	}
}

func (kc *KeyboardControl) IsKeyPressed(key string) bool {
	return funk.ContainsString(kc.pressedKeys, key)
}
