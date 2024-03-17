package components

import (
	"log"
	"path/filepath"

	"github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/interfaces"
	shared_components "github.com/skazanyNaGlany/go.amipi400/shared/components"
	"github.com/winfsp/cgofuse/fuse"
	"golang.org/x/exp/slices"
)

type ADDFileSystem struct {
	shared_components.RunnerBase
	fuse.FileSystemBase

	mountDir string
	mediums  []interfaces.Medium
}

func (addfs *ADDFileSystem) start() {
	if addfs.mountDir == "" {
		addfs.SetRunning(false)

		return
	}

	options := []string{"-o", "allow_other", "-o", "direct_io"}

	host := fuse.NewFileSystemHost(addfs)

	if !host.Mount(addfs.mountDir, options) {
		addfs.SetRunning(false)

		return
	}
}

func (addfs *ADDFileSystem) SetMountDir(mountDir string) {
	addfs.mountDir = mountDir
}

func (addfs *ADDFileSystem) AddMedium(medium interfaces.Medium) {
	addfs.mediums = append(addfs.mediums, medium)
}

func (addfs *ADDFileSystem) RemoveMediumByDevicePathname(
	devicePathname string,
) (interfaces.Medium, error) {
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
func (addfs *ADDFileSystem) FindMediumByPublicFSPathname(
	publicFSPathname string,
) interfaces.Medium {
	fullWithMountPathname := filepath.Join(addfs.mountDir, publicFSPathname)

	for _, medium := range addfs.mediums {
		if medium.GetPublicPathname() == fullWithMountPathname {
			return medium
		}
	}

	return nil
}

// File-system related methods:
// Open
// Truncate
// Getattr
// Readdir
// Read
// Write

func (addfs *ADDFileSystem) Open(path string, flags int) (errc int, fh uint64) {
	if medium := addfs.FindMediumByPublicFSPathname(path); medium != nil {
		return medium.Open(path, flags)
	}

	return -fuse.ENOENT, ^uint64(0)
}

// Block device cannot be truncated, so just return here
func (addfs *ADDFileSystem) Truncate(path string, size int64, fh uint64) int {
	return 0
}

func (addfs *ADDFileSystem) Getattr(
	path string,
	stat *fuse.Stat_t,
	fh uint64,
) (errc int) {
	if path == "/" {
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	}

	if medium := addfs.FindMediumByPublicFSPathname(path); medium != nil {
		result, err := medium.Getattr(path, stat, fh)

		if err != nil {
			if addfs.IsDebugMode() {
				log.Printf("%v: %v\n", path, err)

				return -fuse.EIO
			}
		}

		if result < 0 {
			log.Printf("%v: %v\n", path, result)

			return result
		}

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

func (addfs *ADDFileSystem) Read(
	path string,
	buff []byte,
	ofst int64,
	fh uint64,
) (n int) {
	if medium := addfs.FindMediumByPublicFSPathname(path); medium != nil {
		n, err := medium.Read(path, buff, ofst, fh)

		if err != nil {
			if addfs.IsDebugMode() {
				log.Printf("%v: %v\n", path, err)

				return -fuse.EIO
			}
		}

		if n < 0 {
			log.Printf("%v: %v\n", path, n)

			return n
		}

		return n
	}

	return -fuse.ENOENT
}

func (addfs *ADDFileSystem) Write(path string, buff []byte, ofst int64, fh uint64) int {
	if medium := addfs.FindMediumByPublicFSPathname(path); medium != nil {
		n, err := medium.Write(path, buff, ofst, fh)

		if err != nil {
			if addfs.IsDebugMode() {
				log.Printf("%v: %v\n", path, err)
			}

			if n < 0 {
				return n
			}

			return -fuse.EIO
		}

		if n < 0 {
			log.Printf("%v: %v\n", path, n)

			return n
		}

		return n
	}

	return -fuse.ENOSYS
}

func (addfs *ADDFileSystem) Run() {
	addfs.start()
}
