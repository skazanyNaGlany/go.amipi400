package components

import (
	"os"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
	"gopkg.in/ini.v1"
)

type Mountpoint struct {
	DevicePathname string
	Mountpoint     string
	Label          string
	FsType         string
	Files          []string
	DefaultFile    string
	DFIndex        int
	DHIndex        int
	CDIndex        int
	DHBootPriority int
	Config         *ini.File
}

func (m *Mountpoint) Mount() error {
	return syscall.Mount(
		m.DevicePathname,
		m.Mountpoint,
		m.FsType,
		syscall.MS_SYNCHRONOUS,
		"flush")
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
	var err error

	m.Config, err = ini.Load(
		filepath.Join(m.Mountpoint, shared.MEDIUM_CONFIG_INI_NAME))

	if err != nil {
		return err
	}

	return nil

}

func (m *Mountpoint) LoadDefaultFile() {
	if m.HasFiles() {
		// by default first file in the mountpoint
		// is default file
		m.DefaultFile = m.Files[0]
	}

	if m.Config == nil {
		return
	}

	if !m.Config.HasSection(shared.MEDIUM_CONFIG_DEFAULT_SECTION) {
		return
	}

	if !m.Config.Section(shared.MEDIUM_CONFIG_DEFAULT_SECTION).HasKey(shared.MEDIUM_CONFIG_DEFAULT_FILE) {
		return
	}

	filename := m.Config.Section(shared.MEDIUM_CONFIG_DEFAULT_SECTION).Key(shared.MEDIUM_CONFIG_DEFAULT_FILE).String()

	if filename == shared.MEDIUM_CONFIG_DEFAULT_FILE_NONE {
		m.DefaultFile = shared.MEDIUM_CONFIG_DEFAULT_FILE_NONE
		return
	}

	fullPathname := filepath.Join(m.Mountpoint, filename)

	stat, err := os.Stat(fullPathname)

	if err != nil || stat.IsDir() {
		return
	}

	m.DefaultFile = fullPathname
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

	return &mp
}
