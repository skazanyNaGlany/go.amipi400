package interfaces

import "github.com/winfsp/cgofuse/fuse"

type MediumDriver interface {
	Probe(
		basePath, name string,
		size uint64,
		_type, mountpoint, label, path, fsType, ptType string,
		readOnly bool) (Medium, error)
	Getattr(medium Medium, path string, stat *fuse.Stat_t, fh uint64) (errc int)
	Read(medium Medium, path string, buff []byte, ofst int64, fh uint64) (n int)
	Write(medium Medium, path string, buff []byte, ofst int64, fh uint64) int
	CloseMedium(medium Medium) error
}
