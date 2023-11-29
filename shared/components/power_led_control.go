package components

import (
	"os"
	"strconv"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
)

type PowerLEDControl struct {
	RunnerBase
	blinkPowerLedSecs int
}

func (plc *PowerLEDControl) loop() {
	for plc.running {
		if plc.blinkPowerLedSecs <= 0 {
			time.Sleep(time.Millisecond * 10)
		}

		if plc.blinkPowerLedSecs > 0 {
			plc.blinkPowerLed()
		}

		plc.enablePowerLed()
	}

	plc.running = false
}

func (plc *PowerLEDControl) blinkPowerLed() {
	step := 0

	for plc.blinkPowerLedSecs > 0 {
		if step%2 == 0 {
			plc.disablePowerLed()
		} else {
			plc.enablePowerLed()
		}

		time.Sleep(time.Second * 1)

		plc.blinkPowerLedSecs--
		step++
	}
}

func (plc *PowerLEDControl) setPowerLedBrightness(brightness int) {
	brightnessStr := strconv.FormatInt(int64(brightness), 10)

	utils.FileUtilsInstance.FileWriteBytes(
		shared.POWER_LED0_BRIGHTNESS_PATHNAME,
		0,
		[]byte(brightnessStr),
		os.O_WRONLY,
		0,
		nil)
}

func (plc *PowerLEDControl) disablePowerLed() {
	plc.setPowerLedBrightness(0)
}

func (plc *PowerLEDControl) enablePowerLed() {
	plc.setPowerLedBrightness(100)
}

func (plc *PowerLEDControl) BlinkPowerLEDSecs(seconds int) {
	plc.blinkPowerLedSecs = seconds
}

func (plc *PowerLEDControl) Run() {
	plc.loop()
}
