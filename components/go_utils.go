package components

import (
	"io"
	"log"
	"os"
	"path/filepath"
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
