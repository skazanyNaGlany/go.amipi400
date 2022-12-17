package components

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

type GoUtils struct{}

var regExUtils RegExUtils

func (gu *GoUtils) GetExeDirectory() string {
	return filepath.Dir(os.Args[0])
}

// Change current working directory to directory
// where the executable file is located
func (gu *GoUtils) CwdToExe() string {
	exeDir := gu.GetExeDirectory()

	os.Chdir(exeDir)

	return exeDir
}

// Set the logger to output to screen
// as well to exe_name.txt
func (gu *GoUtils) DuplicateLog() string {
	logFilename := filepath.Base(os.Args[0]) + ".txt"
	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)

	if err != nil {
		panic(err)
	}

	mw := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mw)

	return logFilename
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

	return filepath.Join(pcwd, after), nil
}
