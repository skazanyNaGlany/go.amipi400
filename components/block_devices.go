package components

import (
	"log"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/interfaces"
)

type BlockDevices struct {
	running          bool
	attachedHandlers []interfaces.AttachedBlockDevice
	detachedHandlers []interfaces.DetachedBlockDevice
}

func (bd *BlockDevices) loop() {
	bd.running = true

	for bd.running {
		time.Sleep(time.Millisecond * 10)
	}

	bd.running = false
}

func (bd *BlockDevices) AddAttachedHandler(handler interfaces.AttachedBlockDevice) {
	bd.attachedHandlers = append(bd.attachedHandlers, handler)
}

func (bd *BlockDevices) AddDetachedHandler(handler interfaces.DetachedBlockDevice) {
	bd.detachedHandlers = append(bd.detachedHandlers, handler)
}

func (bd *BlockDevices) Start() error {
	log.Printf("Starting BlockDevices %p\n", bd)

	go bd.loop()

	return nil
}

func (bd *BlockDevices) Stop() error {
	log.Printf("Stopping BlockDevices %v\n", bd)

	bd.running = false

	return nil
}

func (bd *BlockDevices) IsRunning() bool {
	return true
}
