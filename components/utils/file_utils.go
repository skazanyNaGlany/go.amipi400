package utils

import (
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type FileUtils struct{}

var FileUtilsInstance FileUtils

func (fu *FileUtils) GetDirOldestFile(dir string) (os.FileInfo, error) {
	var oldestFile os.FileInfo
	var err error

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return oldestFile, err
	}

	oldestTime := time.Now()
	for _, file := range files {
		if file.Mode().IsRegular() && file.ModTime().Before(oldestTime) {
			oldestFile = file
			oldestTime = file.ModTime()
		}
	}

	if oldestFile == nil {
		err = os.ErrNotExist
	}

	return oldestFile, err
}

func (fu *FileUtils) GetDirSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return err
	})

	return size, err
}

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
	size int64,
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

	if size == -1 {
		stat, err := handle.Stat()

		if err != nil {
			return nil, 0, err
		}

		size = stat.Size()
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

	// needed to set file permissions properly
	defer syscall.Umask(syscall.Umask(0))

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
