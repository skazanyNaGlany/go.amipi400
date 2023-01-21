package components

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/components/utils"
	"github.com/skazanyNaGlany/go.amipi400/consts"
)

type AmiberryEmulator struct {
	EmulatorBase

	emulatorCommand    *exec.Cmd
	executablePathname string
	configPathname     string
	adfs               [consts.MAX_ADFS]string
	commander          *AmiberryCommander
}

func (ae *AmiberryEmulator) SetAmiberryCommander(commander *AmiberryCommander) {
	ae.commander = commander
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

func (ae *AmiberryEmulator) getEmulatorCommandLine(configPathname string) []string {
	commandLine := make([]string, 0)

	commandLine = append(commandLine, ae.executablePathname)
	commandLine = append(commandLine, "--config")
	commandLine = append(commandLine, configPathname)

	return commandLine
}

func (ae *AmiberryEmulator) getEmulatorProcessedConfig() (string, error) {
	templateContent, n, err := utils.FileUtilsInstance.FileReadBytes(
		ae.configPathname,
		0,
		-1,
		0,
		0,
		nil)

	if err != nil {
		return "", err
	}

	if n <= 0 {
		return "", errors.New("Cannot process config file template " + ae.configPathname)
	}

	templateContentStr := string(templateContent)

	// adfs, needs to be the same as
	// nr_floppies= option in the config
	// otherwise some of the config data
	// will be not filled, and it will
	// stay at {{floppyN}}
	for i, pathname := range ae.adfs {
		key := fmt.Sprintf("{{floppy%v}}", i)

		templateContentStr = strings.ReplaceAll(templateContentStr, key, pathname)
	}

	configPathname := filepath.Join(
		os.TempDir(),
		consts.AMIBERRY_TEMPORARY_CONFIG_FILENAME)

	n, err = utils.FileUtilsInstance.FileWriteBytes(
		configPathname,
		0,
		[]byte(templateContentStr),
		os.O_CREATE|os.O_WRONLY,
		0777,
		nil)

	if err != nil {
		return "", err
	}

	if n <= 0 {
		return "", errors.New("Cannot write processed config file " + configPathname)
	}

	return configPathname, nil
}

func (ae *AmiberryEmulator) loop() {
	for ae.IsRunning() {
		time.Sleep(time.Second * 3)

		configPathname, err := ae.getEmulatorProcessedConfig()

		if err != nil {
			if ae.IsDebugMode() {
				log.Println(err)
			}

			break
		}

		commandLine := ae.getEmulatorCommandLine(configPathname)

		if ae.IsVerboseMode() {
			log.Println(
				"Running emulator",
				strings.Join(commandLine, " "))
		}

		ae.emulatorCommand = exec.Command(commandLine[0], commandLine[1:]...)
		ae.emulatorCommand.Dir = filepath.Dir(ae.executablePathname)

		buffer := components.New(make([]byte, 0, 10485760), 10485760)

		ae.emulatorCommand.Stdout = buffer
		ae.emulatorCommand.Stderr = buffer

		ae.emulatorCommand.Start()
		ae.commander.SetProcess(ae.emulatorCommand.Process)
		ae.emulatorCommand.Wait()

		if ae.IsVerboseMode() {
			output := buffer.Buffer.String()
			strOutput := strings.TrimSpace(string(output))

			log.Println("Emulator output")

			utils.GoUtilsInstance.LogPrintLines(strOutput)
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
	if ae.adfs[index] != "" {
		ae.commander.PutSetFloppyCommand(index, "")
		ae.commander.PutConfigChangedCommand()
		ae.commander.PutLocalCommitCommand()
		ae.commander.PutLocalSleepCommand(1)
	}

	ae.adfs[index] = pathname

	ae.commander.PutSetFloppyCommand(index, pathname)
	ae.commander.PutConfigChangedCommand()
	ae.commander.PutLocalCommitCommand()
	ae.commander.PutLocalSleepCommand(1)

	ae.commander.Execute()

	return nil
}

func (ae *AmiberryEmulator) DetachAdf(index int) error {
	if ae.adfs[index] == "" {
		return nil
	}

	ae.commander.PutSetFloppyCommand(index, "")
	ae.commander.PutConfigChangedCommand()
	ae.commander.PutLocalCommitCommand()
	ae.commander.PutLocalSleepCommand(1)

	ae.commander.Execute()

	ae.adfs[index] = ""

	return nil
}

func (ae *AmiberryEmulator) GetAdf(index int) string {
	return ae.adfs[index]
}

func (ae *AmiberryEmulator) SoftReset() error {
	ae.commander.PutUAEResetCommand()
	ae.commander.Execute()

	return nil
}
