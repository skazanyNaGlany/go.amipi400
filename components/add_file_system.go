package components

import (
	"errors"
	"log"
	"path/filepath"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"github.com/winfsp/cgofuse/fuse"
	"golang.org/x/exp/slices"
)

type ADDFileSystem struct {
	fuse.FileSystemBase

	running           bool
	mountDir          string
	mediums           []interfaces.Medium
	preReadCallbacks  []interfaces.PreReadCallback
	postReadCallbacks []interfaces.PostReadCallback
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

func (addfs *ADDFileSystem) RemoveMediumByDevicePathname(devicePathname string) (interfaces.Medium, error) {
	for i, medium := range addfs.mediums {
		if medium.GetDevicePathname() == devicePathname {
			addfs.mediums = slices.Delete(addfs.mediums, i, i+1)

			if err := medium.Close(); err != nil {
				return medium, err
			}

			return medium, nil
		}
	}

	return nil, nil
}

// Find the medium by public file-system pathname
// like /__dev__sda.adf , /__dev__sdb.adf etc.
func (addfs *ADDFileSystem) FindMediumByPublicFSPathname(publicFSPathname string) interfaces.Medium {
	fullWithMountPathname := filepath.Join(addfs.mountDir, publicFSPathname)

	for _, medium := range addfs.mediums {
		if medium.GetPublicPathname() == fullWithMountPathname {
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

func (addfs *ADDFileSystem) Open(path string, flags int) (errc int, fh uint64) {
	if medium := addfs.FindMediumByPublicFSPathname(path); medium != nil {
		return 0, 0
	}

	return -fuse.ENOENT, ^uint64(0)
}

// Block device cannot be truncated, so just return here
func (addfs *ADDFileSystem) Truncate(path string, size int64, fh uint64) int {
	return 0
}

func (addfs *ADDFileSystem) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	if path == "/" {
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	}

	if medium := addfs.FindMediumByPublicFSPathname(path); medium != nil {
		return medium.Getattr(path, stat, fh)
	}

	return -fuse.ENOENT
}

func (addfs *ADDFileSystem) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {
	fill(".", nil, 0)
	fill("..", nil, 0)

	fullMountPath := filepath.Join(addfs.mountDir, path)

	for _, medium := range addfs.mediums {
		publicPathname := medium.GetPublicPathname()
		dirName := filepath.Dir(publicPathname)

		if dirName == fullMountPath {
			fill(medium.GetPublicName(), nil, 0)
		}
	}

	return 0
}

func (addfs *ADDFileSystem) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	if medium := addfs.FindMediumByPublicFSPathname(path); medium != nil {
		for _, callback := range addfs.preReadCallbacks {
			callback(medium, path, buff, ofst, fh)
		}

		startTime := time.Now().Unix()
		result := medium.Read(path, buff, ofst, fh)
		totalTime := time.Now().Unix() - startTime

		for _, callback := range addfs.postReadCallbacks {
			callback(medium, path, buff, ofst, fh, result, totalTime)
		}

		return result
	}

	return -fuse.ENOENT
}

func (addfs *ADDFileSystem) Write(path string, buff []byte, ofst int64, fh uint64) int {
	if medium := addfs.FindMediumByPublicFSPathname(path); medium != nil {
		return medium.Write(path, buff, ofst, fh)
	}

	return -fuse.ENOSYS
}

func (addfs *ADDFileSystem) AddPreReadCallback(preReadCallback interfaces.PreReadCallback) {
	addfs.preReadCallbacks = append(addfs.preReadCallbacks, preReadCallback)
}

func (addfs *ADDFileSystem) AddPostReadCallback(postReadCallback interfaces.PostReadCallback) {
	addfs.postReadCallbacks = append(addfs.postReadCallbacks, postReadCallback)
}
