package components

import (
	"errors"
	"log"
	"path/filepath"

	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"github.com/winfsp/cgofuse/fuse"
	"golang.org/x/exp/slices"
)

type ADDFileSystem struct {
	fuse.FileSystemBase

	running  bool
	mountDir string
	mediums  []interfaces.Medium
}

func (addfs *ADDFileSystem) start() {
	// options := []string{"-d", "-o", "allow_other"}
	options := []string{"-o", "allow_other"}

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

func (addfs *ADDFileSystem) AddMedium(medium interfaces.Medium) {
	addfs.mediums = append(addfs.mediums, medium)
}

func (addfs *ADDFileSystem) RemoveMediumByDevicePathname(devicePathname string) interfaces.Medium {
	for i, medium := range addfs.mediums {
		if medium.GetDevicePathname() == devicePathname {
			addfs.mediums = slices.Delete(addfs.mediums, i, i+1)

			return medium
		}
	}

	return nil
}

// File-system related methods:
// Truncate
// Getattr
// Readdir
// Read
// Write

// Truncate changes the size of a file.
func (addfs *ADDFileSystem) Truncate(path string, size int64, fh uint64) int {
	return 0
}

// func (addfs *ADDFileSystem) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
// 	switch path {
// 	case "/":
// 		stat.Mode = fuse.S_IFDIR | 0555
// 		return 0
// 	// case "/" + filename:
// 	// 	stat.Mode = fuse.S_IFREG | 0444
// 	// 	stat.Size = int64(len(contents))
// 	// 	return 0
// 	default:
// 		return -fuse.ENOENT
// 	}
// }

func (addfs *ADDFileSystem) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	if path == "/" {
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	}

	return -fuse.ENOENT
}

func (addfs *ADDFileSystem) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {
	fill(".", nil, 0)
	fill("..", nil, 0)

	fullPath := filepath.Join(addfs.mountDir, path)

	for _, medium := range addfs.mediums {
		publicPathname := medium.GetPublicPathname()
		dirName := filepath.Dir(publicPathname)

		if dirName == fullPath {
			fill(medium.GetPublicName(), nil, 0)
		}
	}

	return 0
}

func (addfs *ADDFileSystem) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	return -fuse.ENOENT
}

func (addfs *ADDFileSystem) Write(path string, buff []byte, ofst int64, fh uint64) int {
	return -fuse.ENOSYS
}
