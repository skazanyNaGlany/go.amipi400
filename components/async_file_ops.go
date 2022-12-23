package components

import (
	"io"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/ncw/directio"
	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"golang.org/x/exp/slices"
)

type AsyncFileOps struct {
	running    bool
	operations []map[string]any
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
		flag = os.O_RDONLY
	}

	if useHandle == nil {
		_, ok := handles[name]

		if !ok {
			useHandle, err = directio.OpenFile(name, flag, perm)

			if err != nil {
				callback(name, nil, n, offset, useHandle, err)
				return
			}

			handles[name] = useHandle
		}
	}

	useHandle = handles[name]

	if _, err = useHandle.Seek(offset, io.SeekStart); err != nil {
		callback(name, nil, n, offset, useHandle, err)
		return
	}

	block := directio.AlignedBlock(directio.BlockSize)

	n, err = io.ReadFull(useHandle, block)

	if err != nil {
		callback(name, block, n, offset, useHandle, err)
		return
	}

	callback(name, block, n, offset, useHandle, nil)
}

func (afo *AsyncFileOps) executeOperation(ioperation map[string]any, handles map[string]*os.File) {
	if ioperation["type"] == "direct_read" {
		afo.executeDirectReadOperation(ioperation, handles)
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

func (afo *AsyncFileOps) getCountOpsForName(name string) int {
	count := 0

	for _, ioperation := range afo.operations {
		if ioperation["name"].(string) == name {
			count++
		}
	}
	return count
}

func (afo *AsyncFileOps) Start() error {
	log.Printf("Starting AsyncFileOps %p\n", afo)

	afo.running = true

	go afo.loop()

	return nil
}

func (afo *AsyncFileOps) Stop() error {
	log.Printf("Stopping AsyncFileOps %p\n", afo)

	afo.running = false

	return nil
}

func (afo *AsyncFileOps) IsRunning() bool {
	return afo.running
}
