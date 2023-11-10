package components

import (
	"os"
	"strconv"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
)

type LEDControl struct {
	RunnerBase
	blinkPowerLedSecs int
}

func (lc *LEDControl) loop() {
	for lc.running {
		if lc.blinkPowerLedSecs <= 0 {
			time.Sleep(time.Millisecond * 10)
		}

		if lc.blinkPowerLedSecs > 0 {
			lc.blinkPowerLed()
		}
	}

	lc.running = false
}

func (lc *LEDControl) blinkPowerLed() {
	step := 0

	for lc.blinkPowerLedSecs > 0 {
		if step%2 == 0 {
			lc.disablePowerLed()
		} else {
			lc.enablePowerLed()
		}

		time.Sleep(time.Second * 1)

		lc.blinkPowerLedSecs--
		step++
	}
}

func (lc *LEDControl) setPowerLedBrightness(brightness int) {
	brightnessStr := strconv.FormatInt(int64(brightness), 10)

	utils.FileUtilsInstance.FileWriteBytes(
		shared.LED0_BRIGHTNESS_PATHNAME,
		0,
		[]byte(brightnessStr),
		os.O_WRONLY,
		0,
		nil)
}

func (lc *LEDControl) disablePowerLed() {
	lc.setPowerLedBrightness(0)
}

func (lc *LEDControl) enablePowerLed() {
	lc.setPowerLedBrightness(100)
}

func (lc *LEDControl) BlinkPowerLEDSecs(seconds int) {
	lc.blinkPowerLedSecs = seconds
}

func (lc *LEDControl) Run() {
	lc.loop()
}
