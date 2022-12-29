package interfaces

import (
	"os"

	"github.com/winfsp/cgofuse/fuse"
)

type MediumDriver interface {
	Probe(
		basePath, name string,
		size uint64,
		_type, mountpoint, label, path, fsType, ptType string,
		readOnly, force bool) (Medium, error)
	Getattr(medium Medium, path string, stat *fuse.Stat_t, fh uint64) (int, error)
	OpenMediumHandle(medium Medium, readAhead ...int) (*os.File, error)
	Read(medium Medium, path string, buff []byte, ofst int64, fh uint64) (int, error)
	Write(medium Medium, path string, buff []byte, ofst int64, fh uint64) (int, error)
	CloseMedium(medium Medium) error
	SetVerboseMode(verboseMode bool)
	SetDebugMode(debugMode bool)
	GetVerboseMode() bool
	GetDebugMode() bool
	SetOutsideAsyncFileWriterCallback(callback OutsideAsyncFileWriterCallback)
}
