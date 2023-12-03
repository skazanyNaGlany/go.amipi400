package components

import (
	"os/exec"
	"time"

	shared_components "github.com/skazanyNaGlany/go.amipi400/shared/components"
)

type MonitorControl struct {
	shared_components.RunnerBase
	turnOffForSecs int
}

func (mc *MonitorControl) loop() {
	for mc.IsRunning() {
		if mc.turnOffForSecs <= 0 {
			time.Sleep(time.Millisecond * 10)
		}

		if mc.turnOffForSecs > 0 {
			mc.turnOff()
		}
	}

	mc.SetRunning(false)
}

func (mc *MonitorControl) turnOff() {
	for {
		if mc.turnOffForSecs <= 0 {
			break
		}

		mc.TurnOffMonitor()

		time.Sleep(time.Second * 1)

		mc.turnOffForSecs--
	}

	mc.TurnOnMonitor()
}

func (mc *MonitorControl) TurnOffMonitor() {
	exec.Command(
		"xset",
		"dpms",
		"force",
		"off").CombinedOutput()
}

func (mc *MonitorControl) TurnOnMonitor() {
	exec.Command(
		"xset",
		"dpms",
		"force",
		"on").CombinedOutput()
}

func (mc *MonitorControl) TurnOffForSecs(seconds int) {
	mc.turnOffForSecs = seconds
}

func (mc *MonitorControl) Run() {
	mc.loop()
}
