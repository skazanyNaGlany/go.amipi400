package interfaces

import "os"

type FileReadBytesDirectCallback func(path string, block []byte, n int, ofst int64, handle *os.File, err error)
