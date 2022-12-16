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

func (fs *ADDFileSystem) start() {
	host := fuse.NewFileSystemHost(fs)

	if !host.Mount(fs.mountDir, []string{}) {
		fs.running = false

		return
	}
}

func (fs *ADDFileSystem) SetMountDir(mountDir string) {
	fs.mountDir = mountDir
}

func (fs *ADDFileSystem) Start() error {
	if fs.mountDir == "" {
		return errors.New("ADDFileSystem.mountDir not set")
	}

	log.Printf("Starting ADDFileSystem %p\n", fs)

	fs.running = true

	go fs.start()

	return nil
}

func (fs *ADDFileSystem) Stop() error {
	log.Printf("Stopping ADDFileSystem %p\n", fs)

	fs.running = false

	return nil
}

func (bd *ADDFileSystem) IsRunning() bool {
	return bd.running
}
