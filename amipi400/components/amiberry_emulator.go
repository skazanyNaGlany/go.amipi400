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

	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
)

type AmiberryEmulator struct {
	components.RunnerBase

	emulatorCommand       *exec.Cmd
	executablePathname    string
	configPathname        string
	adfs                  [shared.MAX_ADFS]string
	hdfs                  [shared.MAX_HDFS]string
	hdfsBootPriority      [shared.MAX_HDFS]int
	cds                   [shared.MAX_CDS]string
	commander             *AmiberryCommander
	floppySoundVolumeDisk [shared.MAX_ADFS]int // volume per disk
	isAutoHeight          bool
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
		key, value := ae.commander.FormatFloppyCO(i, pathname)

		templateContentStr = strings.ReplaceAll(templateContentStr, "{{"+key+"}}", value)
	}

	// hdfs
	hard_drives := ""

	for i, pathname := range ae.hdfs {
		if pathname == "" {
			continue
		}

		hdfType, err := ae.getHdfType(pathname)

		if err != nil {
			if ae.IsDebugMode() {
				log.Println(pathname, err)
			}

			continue
		}

		bootPriority := ae.hdfsBootPriority[i]
		sectors := 0
		surfaces := 0
		reserved := 0
		blocksize := 512

		if hdfType == shared.HDF_TYPE_HDF {
			sectors = 32
			surfaces = 1
			reserved = 2
			blocksize = 512
		}

		key, value := ae.commander.FormatHardFile2_UaeController_CO(i, pathname, sectors, surfaces, reserved, blocksize, bootPriority, i)
		hard_drives += key + "=" + value + "\n"

		key, value = ae.commander.FormatUaeHf_UaeController_CO(i, pathname, sectors, surfaces, reserved, blocksize, bootPriority, i)
		hard_drives += key + "=" + value + "\n"
	}

	hard_drives = strings.TrimSpace(hard_drives)
	templateContentStr = strings.ReplaceAll(templateContentStr, "{{hard_drives}}", hard_drives)

	// cds
	for i, pathname := range ae.cds {
		key, value := ae.commander.FormatCdImageCO(i, pathname)

		templateContentStr = strings.ReplaceAll(templateContentStr, "{{"+key+"}}", value)
	}

	// floppy sound volume
	for i, volume := range ae.floppySoundVolumeDisk {
		key, value := ae.commander.FormatFloppySoundConfigOption(i, volume > 0)
		templateContentStr = strings.ReplaceAll(templateContentStr, "{{"+key+"}}", value)

		key, value = ae.commander.FormatFloppySoundVolumeDiskCO(i, volume)
		templateContentStr = strings.ReplaceAll(templateContentStr, "{{"+key+"}}", value)
	}

	configPathname := filepath.Join(
		os.TempDir(),
		shared.AMIBERRY_TEMPORARY_CONFIG_FILENAME)

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

		if !shared.AUTORUN_EMULATOR {
			continue
		}

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

		buffer := components.New(
			make([]byte, 0, shared.OUTPUT_BUFFER_MAX_SIZE),
			shared.OUTPUT_BUFFER_MAX_SIZE)

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
		ae.commander.PutFloppyCO(index, "")
		ae.commander.PutConfigChangedCommand()
		ae.commander.PutLocalCommitCommand()
		ae.commander.PutLocalSleepCommand(1)
	}

	ae.adfs[index] = pathname

	ae.commander.PutFloppyCO(index, pathname)
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

	ae.commander.PutFloppyCO(index, "")
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

func (ae *AmiberryEmulator) GetIso(index int) string {
	return ae.cds[index]
}

func (ae *AmiberryEmulator) getHdfType(pathname string) (int, error) {
	stat, err := os.Stat(pathname)

	if err != nil {
		return 0, err
	}

	header, n, err := utils.FileUtilsInstance.FileReadBytes(
		pathname,
		0,
		shared.HDD_SECTOR_SIZE,
		0,
		0,
		nil)

	if err != nil {
		return 0, err
	}

	if n < shared.HDD_SECTOR_SIZE {
		return 0, fmt.Errorf("cannot read file header %v", pathname)
	}

	if header[0] == 'R' && header[1] == 'D' && header[2] == 'S' && header[3] == 'K' {
		return shared.HDF_TYPE_HDFRDB, nil
	} else if header[0] == 'D' && header[1] == 'O' && header[2] == 'S' {
		if stat.Size() < 4*1024*1024 {
			return shared.HDF_TYPE_DISKIMAGE, nil
		}

		return shared.HDF_TYPE_HDF, nil
	}

	return 0, fmt.Errorf("cannot determine HDF type %v", pathname)
}

func (ae *AmiberryEmulator) AttachHdf(index int, bootPriority int, pathname string) error {
	_, err := ae.getHdfType(pathname)

	if err != nil {
		return err
	}

	ae.hdfs[index] = pathname
	ae.hdfsBootPriority[index] = bootPriority

	return nil
}

func (ae *AmiberryEmulator) DetachHdf(index int) error {
	ae.hdfs[index] = ""

	return nil
}

func (ae *AmiberryEmulator) GetHdf(index int) string {
	return ae.hdfs[index]
}

func (ae *AmiberryEmulator) SoftReset() error {
	ae.commander.PutUAEResetCommand()
	ae.commander.Execute()

	return nil
}

func (ae *AmiberryEmulator) HardReset() error {
	// isAutoHeight is temporary so it must be reset when hard
	// resetting the emulator
	ae.isAutoHeight = false

	if ae.emulatorCommand == nil {
		return nil
	}

	if ae.emulatorCommand.Process == nil {
		return nil
	}

	if err := ae.emulatorCommand.Process.Kill(); err != nil {
		if ae.IsDebugMode() {
			log.Println(err)
		}
	}

	return nil
}

func (ae *AmiberryEmulator) AttachCd(index int, pathname string) error {
	if ae.cds[index] != "" {
		ae.commander.PutCdImageCO(index, "")
		ae.commander.PutConfigChangedCommand()
		ae.commander.PutLocalCommitCommand()
		ae.commander.PutLocalSleepCommand(1)
	}

	ae.cds[index] = pathname

	ae.commander.PutCdImageCO(index, pathname)
	ae.commander.PutConfigChangedCommand()
	ae.commander.PutLocalCommitCommand()
	ae.commander.PutLocalSleepCommand(1)

	ae.commander.Execute()

	return nil
}

func (ae *AmiberryEmulator) DetachCd(index int) error {
	if ae.cds[index] == "" {
		return nil
	}

	ae.commander.PutCdImageCO(index, "")
	ae.commander.PutConfigChangedCommand()
	ae.commander.PutLocalCommitCommand()
	ae.commander.PutLocalSleepCommand(1)

	ae.commander.Execute()

	ae.cds[index] = ""

	return nil
}

// This will enable/disable sound for a floppy and set its volume
func (ae *AmiberryEmulator) SetFloppySoundVolumeDisk(index int, volume int) error {
	ae.floppySoundVolumeDisk[index] = volume

	ae.commander.PutFloppySoundCO(index, volume > 0)
	ae.commander.PutFloppySoundVolumeDiskCO(index, volume)

	ae.commander.PutLocalCommitCommand()

	ae.commander.Execute()

	return nil
}

func (ae *AmiberryEmulator) GetFloppySoundVolumeDisk(index int) int {
	return ae.floppySoundVolumeDisk[index]
}

func (ae *AmiberryEmulator) SetAutoHeight(autoHeight bool) {
	// WARNING these settings are temporary and will not be
	// synced with the config at getEmulatorProcessedConfig
	if autoHeight {
		ae.commander.PutAmiberryGfxAutoCropCO(true)
		ae.commander.PutGfxCenterHorizontalCO(true)
		ae.commander.PutGfxCenterVerticalCO(true)
		ae.commander.PutGfxHeightCO(shared.AMIBERRY_DEFAULT_WINDOW_HEIGHT)
		ae.commander.PutGfxHeightWindowedCO(shared.AMIBERRY_DEFAULT_WINDOW_HEIGHT)
	} else {
		ae.commander.PutAmiberryGfxAutoCropCO(false)
		ae.commander.PutGfxCenterHorizontalCO(false)
		ae.commander.PutGfxCenterVerticalCO(false)
		ae.commander.PutGfxHeightCO(shared.AMIBERRY_DEFAULT_WINDOW_HEIGHT)
		ae.commander.PutGfxHeightWindowedCO(shared.AMIBERRY_DEFAULT_WINDOW_HEIGHT)
	}

	ae.commander.PutConfigChangedCommand()
	ae.commander.PutLocalCommitCommand()

	ae.commander.Execute()

	ae.isAutoHeight = autoHeight
}

func (ae *AmiberryEmulator) ToggleAutoHeight() {
	ae.SetAutoHeight(!ae.isAutoHeight)
}
