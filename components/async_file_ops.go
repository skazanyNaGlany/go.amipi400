package components

import (
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/ncw/directio"
	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"golang.org/x/exp/slices"
)

type AsyncFileOps struct {
	RunnerBase
	operations []map[string]any // TODO convert to channel
}

func (afo *AsyncFileOps) loop() {
	for afo.running {
		time.Sleep(time.Millisecond * 10)

		afo.execute()
	}

	afo.running = false
}

func (afo *AsyncFileOps) execute() {
	handles := make(map[string]*os.File)

	for len(afo.operations) > 0 {
		ioperation := afo.operations[0]
		afo.operations = slices.Delete(afo.operations, 0, 0+1)

		afo.executeOperation(ioperation, handles)
	}

	for name := range handles {
		ihandle := handles[name]

		ihandle.Close()

		delete(handles, name)
	}
}

func (afo *AsyncFileOps) openDirectIOHandle(name string, flag int, perm fs.FileMode, handles map[string]*os.File) (*os.File, error) {
	handle, exists := handles[name]

	if exists {
		return handle, nil
	}

	handle, err := directio.OpenFile(name, flag, perm)

	if err != nil {
		return nil, err
	}

	handles[name] = handle

	return handle, nil
}

func (afo *AsyncFileOps) openHandle(name string, flag int, perm fs.FileMode, handles map[string]*os.File) (*os.File, error) {
	handle, exists := handles[name]

	if exists {
		return handle, nil
	}

	handle, err := os.OpenFile(name, flag, perm)

	if err != nil {
		return nil, err
	}

	handles[name] = handle

	return handle, nil
}

func (afo *AsyncFileOps) executeDirectReadOperation(ioperation map[string]any, handles map[string]*os.File) {
	var err error
	var n int

	name := ioperation["name"].(string)
	offset := ioperation["offset"].(int64)
	flag := ioperation["flag"].(int)
	perm := ioperation["perm"].(fs.FileMode)
	useHandle := ioperation["useHandle"].(*os.File)
	callback := ioperation["callback"].(interfaces.FileReadBytesDirectCallback)

	if flag == 0 {
		flag = os.O_RDWR
	}

	if useHandle == nil {
		useHandle, err = afo.openDirectIOHandle(name, flag, perm, handles)

		if err != nil {
			if callback != nil {
				callback(name, nil, n, offset, useHandle, err)
			}

			return
		}
	}

	if _, err = useHandle.Seek(offset, io.SeekStart); err != nil {
		if callback != nil {
			callback(name, nil, n, offset, useHandle, err)
		}

		return
	}

	block := directio.AlignedBlock(directio.BlockSize)

	n, err = io.ReadFull(useHandle, block)

	if err != nil {
		if callback != nil {
			callback(name, block, n, offset, useHandle, err)
		}

		return
	}

	if callback != nil {
		callback(name, block, n, offset, useHandle, nil)
	}
}

func (afo *AsyncFileOps) executeWriteOperation(ioperation map[string]any, handles map[string]*os.File) {
	var err error
	var n int

	name := ioperation["name"].(string)
	offset := ioperation["offset"].(int64)
	buff := ioperation["buff"].([]byte)
	flag := ioperation["flag"].(int)
	perm := ioperation["perm"].(fs.FileMode)
	useHandle := ioperation["useHandle"].(*os.File)
	callback := ioperation["callback"].(interfaces.FileWriteBytesCallback)

	if flag == 0 {
		flag = os.O_RDWR
	}

	if useHandle == nil {
		useHandle, err = afo.openHandle(name, flag, perm, handles)

		if err != nil {
			if callback != nil {
				callback(name, offset, buff, flag, perm, useHandle, 0, err)
			}

			return
		}
	}

	n, err = FileUtilsInstance.FileWriteBytes(name, offset, buff, flag, perm, useHandle)

	if callback != nil {
		callback(name, offset, buff, flag, perm, useHandle, n, err)
	}
}

func (afo *AsyncFileOps) executeOperation(ioperation map[string]any, handles map[string]*os.File) {
	// TODO move types to consts.go
	if ioperation["type"] == "direct_read" {
		afo.executeDirectReadOperation(ioperation, handles)
	} else if ioperation["type"] == "write" {
		afo.executeWriteOperation(ioperation, handles)
	}
}

func (afo *AsyncFileOps) FileReadBytesDirect(
	name string,
	offset int64,
	flag int,
	perm fs.FileMode,
	useHandle *os.File,
	max int,
	callback interfaces.FileReadBytesDirectCallback) {
	if max > 0 {
		count := afo.getCountOpsForName(name)

		if count >= max {
			return
		}
	}

	op := make(map[string]any)

	op["name"] = name
	op["offset"] = offset
	op["flag"] = flag
	op["perm"] = perm
	op["useHandle"] = useHandle
	op["type"] = "direct_read"
	op["callback"] = callback

	afo.operations = append(afo.operations, op)
}

func (afo *AsyncFileOps) FileWriteBytes(
	name string,
	offset int64,
	buff []byte,
	flag int,
	perm fs.FileMode,
	useHandle *os.File,
	max int,
	callback interfaces.FileWriteBytesCallback) {
	if max > 0 {
		count := afo.getCountOpsForName(name)

		if count >= max {
			return
		}
	}

	// make a copy of the buffer to
	// avoid race condition issues
	buffCopy := make([]byte, len(buff))
	copy(buffCopy, buff)

	op := make(map[string]any)

	op["name"] = name
	op["offset"] = offset
	op["buff"] = buffCopy
	op["flag"] = flag
	op["perm"] = perm
	op["useHandle"] = useHandle
	op["type"] = "write"
	op["callback"] = callback

	afo.operations = append(afo.operations, op)
}

func (afo *AsyncFileOps) getCountOpsForName(name string) int {
	count := 0

	for _, ioperation := range afo.operations {
		if ioperation["name"].(string) == name {
			count++
		}
	}
	return count
}

func (afo *AsyncFileOps) Run() {
	afo.loop()
}
