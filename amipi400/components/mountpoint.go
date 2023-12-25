package components

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
)

type Mountpoint struct {
	DevicePathname string
	Mountpoint     string
	Label          string
	FsType         string
	Files          []string
	DFIndex        int
	DHIndex        int
	CDIndex        int
	DHBootPriority int
	Config         *MountpointConfig
}

func (m *Mountpoint) Mount() error {
	// do not use syscall.MS_SYNCHRONOUS as "flags" parameter
	// or "flush" as "data" parameter since it will slow down
	// the emulator when accessing / reading / writing rom
	// files
	return syscall.Mount(
		m.DevicePathname,
		m.Mountpoint,
		m.FsType,
		0,
		"")
}

func (m *Mountpoint) Unmount() error {
	return syscall.Unmount(m.Mountpoint, syscall.MNT_FORCE)
}

func (m *Mountpoint) Fix() (string, error) {
	return utils.UnixUtilsInstance.RunFsck(m.DevicePathname)
}

func (m *Mountpoint) HasFiles() bool {
	return len(m.Files) > 0
}

func (m *Mountpoint) LoadFiles(extensions []string) {
	files := utils.FileUtilsInstance.GetDirFiles(m.Mountpoint, false, extensions...)

	sort.Strings(files)

	m.Files = files
}

func (m *Mountpoint) LoadConfig() error {
	err := m.Config.Load()

	if err != nil {
		m.setupDefaultFileNoConfig()

		return err
	}

	return m.setupDefaultFilePath()
}

func (m *Mountpoint) setupDefaultFileNoConfig() {
	if !m.HasFiles() {
		return
	}

	m.Config.AmiPi400.DefaultFile = m.Files[0]
}

func (m *Mountpoint) setupDefaultFilePath() error {
	if m.Config.AmiPi400.DefaultFile == shared.MEDIUM_CONFIG_DEFAULT_FILE_NONE {
		return nil
	}

	fullPathname := filepath.Join(m.Mountpoint, m.Config.AmiPi400.DefaultFile)

	stat, err := os.Stat(fullPathname)

	if err != nil {
		return err
	}

	if stat.IsDir() {
		return fmt.Errorf("%v must be a file, not a directory", fullPathname)
	}

	m.Config.AmiPi400.DefaultFile = fullPathname

	return nil
}

func NewMountpoint(devicePathname string, mountpoint string, label string, fsType string) *Mountpoint {
	mp := Mountpoint{
		DevicePathname: devicePathname,
		Mountpoint:     mountpoint,
		Label:          label,
		FsType:         fsType}

	mp.Files = make([]string, 0)
	mp.DFIndex = shared.DRIVE_INDEX_UNSPECIFIED
	mp.DHIndex = shared.DRIVE_INDEX_UNSPECIFIED
	mp.CDIndex = shared.DRIVE_INDEX_UNSPECIFIED
	mp.Config = NewMountpointConfig(filepath.Join(mountpoint, shared.MEDIUM_CONFIG_INI_NAME))

	return &mp
}
