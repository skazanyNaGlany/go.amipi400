package components

import (
	"log"
	"time"

	"github.com/itchyny/volume-go"
)

type VolumeControl struct {
	running     bool
	muteForSecs int
}

func (vc *VolumeControl) loop() {
	for vc.running {
		if vc.muteForSecs <= 0 {
			time.Sleep(time.Millisecond * 10)
		}

		if vc.muteForSecs > 0 {
			vc.mute()
		}
	}

	vc.running = false
}

func (vc *VolumeControl) mute() {
	if err := volume.Mute(); err != nil {
		log.Println(err)

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
		log.Println(err)
	}
}

func (vc *VolumeControl) MuteForSecs(seconds int) {
	vc.muteForSecs = seconds
}

func (vc *VolumeControl) Start() error {
	log.Printf("Starting VolumeControl %p\n", vc)

	vc.running = true

	go vc.loop()

	return nil
}

func (vc *VolumeControl) Stop() error {
	log.Printf("Stopping VolumeControl %p\n", vc)

	vc.running = false

	return nil
}

func (vc *VolumeControl) IsRunning() bool {
	return vc.running
}
