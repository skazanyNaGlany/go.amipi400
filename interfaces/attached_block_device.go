package interfaces

type AttachedBlockDevice func(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool)
