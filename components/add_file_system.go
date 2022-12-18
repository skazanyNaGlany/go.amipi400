package components

import (
	"errors"
	"log"

	"github.com/winfsp/cgofuse/fuse"
)

type ADDFileSystem struct {
	fuse.FileSystemBase

	running  bool
	mountDir string
}

func (addfs *ADDFileSystem) start() {
	options := []string{"-d", "-o", "allow_other"}

	host := fuse.NewFileSystemHost(addfs)

	if !host.Mount(addfs.mountDir, options) {
		addfs.running = false

		return
	}
}

func (addfs *ADDFileSystem) SetMountDir(mountDir string) {
	addfs.mountDir = mountDir
}

func (addfs *ADDFileSystem) Start() error {
	if addfs.mountDir == "" {
		return errors.New("ADDFileSystem.mountDir not set")
	}

	log.Printf("Starting ADDFileSystem %p\n", addfs)

	addfs.running = true

	go addfs.start()

	return nil
}

func (addfs *ADDFileSystem) Stop() error {
	log.Printf("Stopping ADDFileSystem %p\n", addfs)

	addfs.running = false

	return nil
}

func (addfs *ADDFileSystem) IsRunning() bool {
	return addfs.running
}
