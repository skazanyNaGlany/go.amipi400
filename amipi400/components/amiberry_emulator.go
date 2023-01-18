package components

import (
	"log"
	"os/exec"
	"strings"
	"time"
)

type AmiberryEmulator struct {
	EmulatorBase

	executablePathname string
	configPathname     string
	emulatorCommand    *exec.Cmd
}

func (ae *AmiberryEmulator) Run() {
	if ae.executablePathname == "" {
		if ae.IsDebugMode() {
			log.Println("Emulator executable pathname not set")
		}

		ae.SetRunning(false)
		return
	}

	if ae.configPathname == "" {
		if ae.IsDebugMode() {
			log.Println("Emulator config pathname not set")
		}

		ae.SetRunning(false)
		return
	}

	ae.loop()
}

func (ae *AmiberryEmulator) GetEmulatorCommandLine() []string {
	commandLine := make([]string, 0)

	commandLine = append(commandLine, ae.executablePathname)
	commandLine = append(commandLine, "--config")
	commandLine = append(commandLine, ae.configPathname)

	return commandLine
}

func (ae *AmiberryEmulator) loop() {
	for ae.IsRunning() {
		time.Sleep(time.Second * 3)

		commandLine := ae.GetEmulatorCommandLine()

		if ae.IsVerboseMode() {
			log.Println(
				"Running emulator",
				strings.Join(commandLine, " "))
		}

		ae.emulatorCommand = exec.Command(commandLine[0], commandLine[1:]...)

		if err := ae.emulatorCommand.Start(); err != nil {
			if ae.IsDebugMode() {
				log.Println(err)
			}

			break
		}

		output, err := ae.emulatorCommand.CombinedOutput()

		if err != nil {
			if ae.IsDebugMode() {
				log.Println(err)
			}

			break
		}

		if ae.IsVerboseMode() {
			log.Println("Emulator output", output)
		}
	}

	ae.SetRunning(false)
}

func (ae *AmiberryEmulator) SetExecutablePathname(pathname string) {
	ae.executablePathname = pathname
}

func (ae *AmiberryEmulator) GetExecutablePathname() string {
	return ae.executablePathname
}

func (ae *AmiberryEmulator) SetConfigPathname(pathname string) {
	ae.configPathname = pathname
}

func (ae *AmiberryEmulator) GetConfigPathname() string {
	return ae.configPathname
}
