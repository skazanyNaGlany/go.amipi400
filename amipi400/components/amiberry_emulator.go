package components

import (
	"log"
	"os/exec"
	"strings"
	"time"
)

type AmiberryEmulator struct {
	EmulatorBase

	emulatorCommand *exec.Cmd
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

		output, err := ae.emulatorCommand.CombinedOutput()

		if err != nil {
			if ae.IsDebugMode() {
				log.Println(err)
			}

			break
		}

		if ae.IsVerboseMode() {
			strOutput := strings.TrimSpace(string(output))

			log.Println("Emulator output\n", strOutput)
		}
	}

	ae.SetRunning(false)
}
