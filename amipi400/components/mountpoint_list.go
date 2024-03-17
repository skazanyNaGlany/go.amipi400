package components

import "slices"

type MountpointList struct {
	Mountpoints []*Mountpoint
}

func (ml *MountpointList) Clear() {
	ml.Mountpoints = make([]*Mountpoint, 0)
}

func (ml *MountpointList) FindMountpoint(mountpoint *Mountpoint) int {
	for i, iMp := range ml.Mountpoints {
		if iMp == mountpoint {
			return i
		}
	}

	return -1
}

func (ml *MountpointList) AddMountpoint(mountpoint *Mountpoint) {
	if ml.FindMountpoint(mountpoint) != -1 {
		return
	}

	ml.Mountpoints = append(ml.Mountpoints, mountpoint)
}

func (ml *MountpointList) RemoveMountpoint(mountpoint *Mountpoint) {
	if index := ml.FindMountpoint(mountpoint); index != -1 {
		ml.Mountpoints = slices.Delete(ml.Mountpoints, index, index+1)
	}
}

func (ml *MountpointList) GetMountpointByDevicePathname(
	devicePathname string,
) *Mountpoint {
	for _, iMp := range ml.Mountpoints {
		if iMp.DevicePathname == devicePathname {
			return iMp
		}
	}

	return nil
}

func (ml *MountpointList) GetMountpointByLabel(label string) *Mountpoint {
	for _, iMp := range ml.Mountpoints {
		if iMp.Label == label {
			return iMp
		}
	}

	return nil
}

func (ml *MountpointList) GetMountpoint(mountpoint string) *Mountpoint {
	for _, iMp := range ml.Mountpoints {
		if iMp.Mountpoint == mountpoint {
			return iMp
		}
	}

	return nil
}

func (ml *MountpointList) HasMountpointByLabel(label string) bool {
	return ml.GetMountpointByLabel(label) != nil
}

func (ml *MountpointList) HasMountpoint(mountpoint string) bool {
	return ml.GetMountpoint(mountpoint) != nil
}

func (ml *MountpointList) GetMountpointByDFIndex(index int) *Mountpoint {
	for _, iMp := range ml.Mountpoints {
		if iMp.DFIndex == index {
			return iMp
		}
	}

	return nil
}

func (ml *MountpointList) GetMountpointByDHIndex(index int) *Mountpoint {
	for _, iMp := range ml.Mountpoints {
		if iMp.DHIndex == index {
			return iMp
		}
	}

	return nil
}

func (ml *MountpointList) GetMountpointByCDIndex(index int) *Mountpoint {
	for _, iMp := range ml.Mountpoints {
		if iMp.CDIndex == index {
			return iMp
		}
	}

	return nil
}

func NewMountpointList() *MountpointList {
	ml := MountpointList{}
	ml.Mountpoints = make([]*Mountpoint, 0)

	return &ml
}
