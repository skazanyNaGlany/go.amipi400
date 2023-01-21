package utils

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

type GoUtils struct{}

var GoUtilsInstance GoUtils

func (gu *GoUtils) CheckPlatform() {
	if runtime.GOOS != "linux" {
		log.Fatalln("This app can be used only on Linux.")
	}
}

func (gu *GoUtils) GetExeDirectory() string {
	return filepath.Dir(os.Args[0])
}

// Change current working directory to directory
// where the executable file is located
func (gu *GoUtils) CwdToExeOrScript() (string, error) {
	exe, err := gu.GetScriptOrExecutable()

	if err != nil {
		return "", err
	}

	exeDir := filepath.Dir(exe)

	os.Chdir(exeDir)

	return exeDir, nil
}

func (gu *GoUtils) MustCwdToExeOrScript() string {
	exeDir, err := gu.CwdToExeOrScript()

	if err != nil {
		log.Fatalln(err)
	}

	return exeDir
}

// // Change current working directory to directory
// // where the executable file is located
// func (gu *GoUtils) CwdToExe() string {
// 	exeDir := gu.GetExeDirectory()

// 	os.Chdir(exeDir)

// 	return exeDir
// }

// Set the logger to output to screen
// as well to exe_name.txt
func (gu *GoUtils) DuplicateLog(parentDir string) (string, error) {
	logFilename := filepath.Join(
		parentDir,
		filepath.Base(os.Args[0])+".txt",
	)
	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)

	if err != nil {
		return "", err
	}

	mw := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mw)

	return logFilename, nil
}

func (gu *GoUtils) MustDuplicateLog(exeDir string) string {
	logFilePathname, err := gu.DuplicateLog(exeDir)

	if err != nil {
		log.Fatalln(err)
	}

	return logFilePathname
}

func (gu *GoUtils) GetProcess(pid int32) (*process.Process, error) {
	processes, err := process.Processes()

	if err != nil {
		return nil, err
	}

	for _, process := range processes {
		if process.Pid == pid {
			return process, nil
		}
	}

	return nil, nil
}

// If you run your app by executing build binary, it will return the path to your binary.
// If you run your app by executing a .go file with go run file.go command, it will try to
// return path to the script.
//
// Only plain "go run script.go" syntax is supported.
//
// Requires github.com/shirou/gopsutil/v3/process
func (gu *GoUtils) GetScriptOrExecutable() (string, error) {
	process, err := gu.GetProcess(int32(os.Getppid()))

	if err != nil || process == nil {
		return os.Executable()
	}

	cmdLine, err := process.Cmdline()

	if err != nil {
		return "", err
	}

	before, after, found := strings.Cut(cmdLine, " run ")

	if !found {
		return os.Executable()
	}

	before = strings.TrimSpace(before)

	if before != "go" && before != "go.exe" {
		return os.Executable()
	}

	after = strings.TrimSpace(after)

	if !strings.HasSuffix(after, ".go") {
		return os.Executable()
	}

	pcwd, err := process.Cwd()

	if err != nil {
		return os.Executable()
	}

	fullPathname := filepath.Join(pcwd, after)

	if _, err := os.Stat(fullPathname); err != nil {
		return os.Executable()
	}

	return fullPathname, nil
}

func (gu *GoUtils) Bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (gu *GoUtils) BoolToStrInt(b bool) string {
	if b {
		return "1"
	}

	return "0"
}

func (gu *GoUtils) ByteSliceToStruct(data []byte, target any) error {
	reader := bytes.NewReader(data)

	if err := binary.Read(reader, binary.LittleEndian, target); err != nil {
		return err
	}

	return nil
}

func (gu *GoUtils) StructToByteSlice(source any) ([]byte, error) {
	var buffer bytes.Buffer

	err := binary.Write(&buffer, binary.LittleEndian, source)

	return buffer.Bytes(), err
}

func (gu *GoUtils) LogPrintLines(lines string) {
	for _, iline := range strings.Split(lines, "\n") {
		iline = strings.TrimSpace(iline)

		log.Println(iline)
	}
}
