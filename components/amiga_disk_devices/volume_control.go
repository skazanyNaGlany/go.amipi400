package amigadiskdevices

import (
	"log"
	"time"

	"github.com/itchyny/volume-go"
	"github.com/skazanyNaGlany/go.amipi400/components"
)

type VolumeControl struct {
	components.RunnerBase
	muteForSecs int
}

func (vc *VolumeControl) loop() {
	for vc.IsRunning() {
		if vc.muteForSecs <= 0 {
			time.Sleep(time.Millisecond * 10)
		}

		if vc.muteForSecs > 0 {
			vc.mute()
		}
	}

	vc.SetRunning(false)
}

func (vc *VolumeControl) mute() {
	if err := volume.Mute(); err != nil {
		if vc.GetDebugMode() {
			log.Println(err)
		}

		vc.muteForSecs = 0

		return
	}

	for {
		if vc.muteForSecs <= 0 {
			break
		}

		time.Sleep(time.Second * 1)

		vc.muteForSecs--
	}

	if err := volume.Unmute(); err != nil {
		if vc.GetDebugMode() {
			log.Println(err)
		}
	}
}

func (vc *VolumeControl) MuteForSecs(seconds int) {
	vc.muteForSecs = seconds
}

func (vc *VolumeControl) Run() {
	vc.loop()
}
