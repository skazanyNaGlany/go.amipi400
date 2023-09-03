package interfaces

import (
	"io/fs"
	"os"
)

type FileWriteBytesCallback func(name string, offset int64, buff []byte, flag int, perm fs.FileMode, useHandle *os.File, n int, err error)
