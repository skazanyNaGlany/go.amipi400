package components

import (
	"io"
	"io/fs"
	"os"
)

type FileUtils struct{}

var fileUtils FileUtils

func (fu *FileUtils) FileReadBytes(
	name string,
	offset int64,
	size uint64,
	flag int,
	perm fs.FileMode,
	useHandle *os.File) ([]byte, int, error) {
	var handle *os.File
	var err error
	var n int

	if useHandle == nil {
		if flag == 0 {
			flag = os.O_RDONLY
		}

		handle, err = os.OpenFile(name, flag, perm)

		if err != nil {
			return nil, 0, err
		}

		defer handle.Close()
	} else {
		handle = useHandle
	}

	if _, err = handle.Seek(offset, io.SeekStart); err != nil {
		return nil, 0, err
	}

	data := make([]byte, size)

	n, err = handle.Read(data)

	if err != nil {
		return nil, 0, err
	}

	return data, n, nil
}
