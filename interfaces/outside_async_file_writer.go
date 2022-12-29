package interfaces

import (
	"io/fs"
	"os"
)

type OutsideAsyncFileWriterCallback func(name string, offset int64, buff []byte, flag int, perm fs.FileMode, useHandle *os.File)
