package interfaces

import "github.com/winfsp/cgofuse/fuse"

type MediumDriver interface {
	Probe(pathname string, size uint64, _type string, readOnly bool) (Medium, error)
	Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int)
	Read(path string, buff []byte, ofst int64, fh uint64) (n int)
	Write(path string, buff []byte, ofst int64, fh uint64) int
}
