package components

import (
	"os"
	"strconv"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
)

type NumLockLEDControl struct {
	RunnerBase
	blinkNumLockLedSecs int
}

func (nllc *NumLockLEDControl) loop() {
	for nllc.running {
		if nllc.blinkNumLockLedSecs <= 0 {
			time.Sleep(time.Millisecond * 10)
		}

		if nllc.blinkNumLockLedSecs > 0 {
			nllc.blinkNumLockLed()
		}

		nllc.disableNumLockLed()
	}

	nllc.running = false
}

func (nllc *NumLockLEDControl) blinkNumLockLed() {
	step := 0

	for nllc.blinkNumLockLedSecs > 0 {
		if step%2 == 0 {
			nllc.disableNumLockLed()
		} else {
			nllc.enableNumLockLed()
		}

		time.Sleep(time.Second * 1)

		nllc.blinkNumLockLedSecs--
		step++
	}
}

func (nllc *NumLockLEDControl) setNumLockLedState(state bool) {
	stateInt := 0

	if state {
		stateInt = 1
	}

	stateStr := strconv.FormatInt(int64(stateInt), 10)

	utils.FileUtilsInstance.FileWriteBytes(
		shared.NUM_LOCK_LED0_BRIGHTNESS_PATHNAME,
		0,
		[]byte(stateStr),
		os.O_WRONLY,
		0,
		nil)
}

func (nllc *NumLockLEDControl) disableNumLockLed() {
	nllc.setNumLockLedState(false)
}

func (nllc *NumLockLEDControl) enableNumLockLed() {
	nllc.setNumLockLedState(true)
}

func (nllc *NumLockLEDControl) BlinkNumLockLEDSecs(seconds int) {
	nllc.blinkNumLockLedSecs = seconds
}

func (nllc *NumLockLEDControl) Run() {
	nllc.loop()
}
