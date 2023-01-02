package components

import (
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FileUtils struct{}

var FileUtilsInstance FileUtils

func (fu *FileUtils) GetDirFiles(dir string) []string {
	result := make([]string, 0)
	files, _ := ioutil.ReadDir(dir)

	for _, filename := range files {
		result = append(result, filepath.Join(dir, filename.Name()))
	}

	return result
}

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

func (fu *FileUtils) FileWriteBytes(
	name string,
	offset int64,
	buff []byte,
	flag int,
	perm fs.FileMode,
	useHandle *os.File) (int, error) {
	var handle *os.File
	var err error
	var n int

	if useHandle == nil {
		if flag == 0 {
			flag = os.O_WRONLY
		}

		handle, err = os.OpenFile(name, flag, perm)

		if err != nil {
			return 0, err
		}

		defer handle.Close()
	} else {
		handle = useHandle
	}

	if _, err = handle.Seek(offset, io.SeekStart); err != nil {
		return 0, err
	}

	n, err = handle.Write(buff)

	if err != nil {
		return 0, err
	}

	return n, nil
}
