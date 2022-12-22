package components

import (
	"log"
	"os"
	"strconv"
	"time"
)

const led0brightnessPathname = "/sys/class/leds/led0/brightness"

type LEDControl struct {
	running           bool
	blinkPowerLedSecs int
}

func (vc *LEDControl) loop() {
	for vc.running {
		if vc.blinkPowerLedSecs <= 0 {
			time.Sleep(time.Millisecond * 10)
		}

		if vc.blinkPowerLedSecs > 0 {
			vc.blinkPowerLed()
		}
	}

	vc.running = false
}

func (vc *LEDControl) blinkPowerLed() {
	for vc.blinkPowerLedSecs > 0 {
		vc.disablePowerLed()

		time.Sleep(time.Second * 1)

		vc.enablePowerLed()

		time.Sleep(time.Second * 1)

		vc.blinkPowerLedSecs--
	}
}

func (vc *LEDControl) setPowerLedBrightness(brightness int) {
	brightnessStr := strconv.FormatInt(int64(brightness), 10)

	FileUtilsInstance.FileWriteBytes(
		led0brightnessPathname,
		0,
		[]byte(brightnessStr),
		os.O_WRONLY,
		0,
		nil)
}

func (vc *LEDControl) disablePowerLed() {
	vc.setPowerLedBrightness(0)
}

func (vc *LEDControl) enablePowerLed() {
	vc.setPowerLedBrightness(100)
}

func (vc *LEDControl) BlinkPowerLEDSecs(seconds int) {
	vc.blinkPowerLedSecs = seconds
}

func (vc *LEDControl) Start() error {
	log.Printf("Starting LEDControl %p\n", vc)

	vc.running = true

	go vc.loop()

	return nil
}

func (vc *LEDControl) Stop() error {
	log.Printf("Stopping LEDControl %p\n", vc)

	vc.running = false

	return nil
}

func (vc *LEDControl) IsRunning() bool {
	return vc.running
}
