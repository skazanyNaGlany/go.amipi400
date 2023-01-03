package components

import (
	"os"
	"strconv"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/components/utils"
	"github.com/skazanyNaGlany/go.amipi400/consts"
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
	for lc.blinkPowerLedSecs > 0 {
		lc.disablePowerLed()

		time.Sleep(time.Second * 1)

		lc.enablePowerLed()

		time.Sleep(time.Second * 1)

		lc.blinkPowerLedSecs--
	}
}

func (lc *LEDControl) setPowerLedBrightness(brightness int) {
	brightnessStr := strconv.FormatInt(int64(brightness), 10)

	utils.FileUtilsInstance.FileWriteBytes(
		consts.LED0_BRIGHTNESS_PATHNAME,
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
