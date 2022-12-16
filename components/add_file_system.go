package components

import (
	"log"
)

type ADDFileSystem struct {
	running bool
}

func (bd *ADDFileSystem) Start() error {
	log.Printf("Starting ADDFileSystem %p\n", bd)

	return nil
}

func (bd *ADDFileSystem) Stop() error {
	log.Printf("Stopping ADDFileSystem %p\n", bd)

	bd.running = false

	return nil
}

func (bd *ADDFileSystem) IsRunning() bool {
	return bd.running
}
