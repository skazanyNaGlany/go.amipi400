package interfaces

type AttachedBlockDevice func(
	name string,
	size int,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) (bool, error)
