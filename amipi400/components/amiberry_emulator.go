package components

import (
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/consts"
)

type AmiberryEmulator struct {
	EmulatorBase

	emulatorCommand    *exec.Cmd
	executablePathname string
	configPathname     string
	adfs               [consts.MAX_ADFS]string
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

func (ae *AmiberryEmulator) AttachAdf(index int, pathname string) error {
	ae.adfs[index] = pathname

	return nil
}

func (ae *AmiberryEmulator) DetachAdf(index int) error {
	ae.adfs[index] = ""

	return nil
}
