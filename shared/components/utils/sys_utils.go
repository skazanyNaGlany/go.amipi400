package utils

import (
	"fmt"
	"log"
	"os/exec"
)

type SysUtils struct{}

var SysUtilsInstance SysUtils

func (su *SysUtils) CheckForExecutables(executables []string) {
	for _, iExe := range executables {
		if _, err := exec.LookPath(iExe); err != nil {
			msg := fmt.Sprintf("(executable: %v)", iExe)

			log.Fatalln(err, msg)
		}
	}
}
