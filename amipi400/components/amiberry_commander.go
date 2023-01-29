package components

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/components/utils"
	"golang.org/x/exp/slices"
)

// CO - config option
type AmiberryCommander struct {
	components.RunnerBase

	tmpIniPathname string
	process        *os.Process
	commands       []string
	emulatorPaused bool
	executeLoops   int
}

func (ac *AmiberryCommander) SetEmulatorPaused(paused bool) {
	ac.emulatorPaused = paused
}

func (ac *AmiberryCommander) SetTmpIniPathname(pathname string) {
	ac.tmpIniPathname = pathname
}

func (ac *AmiberryCommander) SetProcess(process *os.Process) {
	ac.process = process
}

func (ac *AmiberryCommander) writeTmpIni(commands string) error {
	if ac.IsVerboseMode() {
		log.Println("Writing", ac.tmpIniPathname)
	}

	commands = "[commands]\n" + commands

	byteCommands := []byte(commands)

	// just in case
	os.Remove(ac.tmpIniPathname)

	n, err := utils.FileUtilsInstance.FileWriteBytes(
		ac.tmpIniPathname,
		0,
		byteCommands,
		os.O_CREATE|os.O_RDWR|os.O_TRUNC,
		0644,
		nil)

	if err != nil {
		return err
	}

	if n < len(byteCommands) {
		return errors.New("Cannot write tmp ini file " + ac.tmpIniPathname)
	}

	return nil
}

func (ac *AmiberryCommander) sendUSR1Signal() error {
	if ac.IsVerboseMode() {
		log.Println("Sending USR1 signal to the emulator")
	}

	if ac.process == nil {
		return errors.New("emulator not running")
	}

	return ac.process.Signal(syscall.SIGUSR1)
}

func (ac *AmiberryCommander) blockTillTmpIniExists() {
	for {
		time.Sleep(time.Millisecond * 10)

		_, err := os.Stat(ac.tmpIniPathname)

		if err != nil {
			return
		}
	}
}

func (ac *AmiberryCommander) executeLocalCommand(command string, currentCommands string) bool {
	if command == "local-commit" {
		if currentCommands == "" {
			return false
		}

		if ac.IsVerboseMode() {
			log.Println("Commiting")
		}

		if err := ac.sendCommands(currentCommands); err != nil {
			if ac.IsDebugMode() {
				log.Println(err)
			}
		}

		return true
	} else if strings.HasPrefix(command, "local-sleep ") {
		command := strings.ReplaceAll(command, "local-sleep ", "")

		secs, err := strconv.ParseInt(command, 10, 32)

		if err != nil {
			return true
		}

		if ac.IsVerboseMode() {
			log.Println("Sleeping for", secs, "seconds")
		}

		time.Sleep(time.Second * time.Duration(secs))

		return true
	}

	return false
}

func (ac *AmiberryCommander) executeCommands() {
	currentCommands := ""
	index := 0

	for len(ac.commands) > 0 {
		icommand := ac.commands[0]
		ac.commands = slices.Delete(ac.commands, 0, 0+1)

		if strings.HasPrefix(icommand, "local-") {
			if ac.executeLocalCommand(icommand, currentCommands) {
				currentCommands = ""
				index = 0
			}

			continue
		}

		currentCommands += fmt.Sprintf("cmd%v=%v\n", index, icommand)
		index++
	}

	if currentCommands != "" {
		if err := ac.sendCommands(currentCommands); err != nil {
			if ac.IsDebugMode() {
				log.Println(err)
			}
		}
	}
}

func (ac *AmiberryCommander) sendCommands(commands string) error {
	commands = strings.TrimSpace(commands)

	if ac.IsVerboseMode() {
		log.Println("Sending commands to the emulator")

		utils.GoUtilsInstance.LogPrintLines(commands)
	}

	if err := ac.writeTmpIni(commands); err != nil {
		return err
	}

	if err := ac.sendUSR1Signal(); err != nil {
		return err
	}

	ac.blockTillTmpIniExists()

	if ac.IsVerboseMode() {
		log.Println("Commands sent")
	}

	return nil
}

func (ac *AmiberryCommander) loop() {
	for ac.IsRunning() {
		time.Sleep(time.Millisecond * 10)

		for ac.executeLoops > 0 {
			ac.executeLoops--

			ac.executeCommands()
		}
	}

	ac.SetRunning(false)
}

func (ac *AmiberryCommander) Execute() {
	ac.executeLoops++
}

func (ac *AmiberryCommander) Run() {
	ac.loop()
}

func (ac *AmiberryCommander) PutCommand(command string, reset bool, force bool) {
	if reset {
		ac.commands = make([]string, 0)
	}

	if ac.emulatorPaused && !force {
		return
	}

	if command == "" {
		return
	}

	ac.commands = append(ac.commands, command)
}

func (ac *AmiberryCommander) PutUAEResetCommand() {
	ac.PutCommand("uae_reset 1,1", false, false)
}

func (ac *AmiberryCommander) PutConfigChangedCommand() {
	ac.PutCommand("config_changed 1", false, false)
}

func (ac *AmiberryCommander) PutSetConfigOptionCommand(option string, value string) {
	full := fmt.Sprintf("cfgfile_parse_line_type_all %v=%v", option, value)

	ac.PutCommand(full, false, false)
}

func (ac *AmiberryCommander) FormatFloppyCO(index int, pathname string) (string, string) {
	key := fmt.Sprintf("floppy%v", index)

	return key, pathname
}

func (ac *AmiberryCommander) PutFloppyCO(index int, pathname string) {
	key, value := ac.FormatFloppyCO(index, pathname)

	ac.PutSetConfigOptionCommand(key, value)
}

func (ac *AmiberryCommander) PutLocalCommitCommand() {
	ac.PutCommand("local-commit", false, false)
}

func (ac *AmiberryCommander) PutLocalSleepCommand(sleepSecs int) {
	sleepSecsStr := fmt.Sprintf("%v", sleepSecs)

	ac.PutCommand("local-sleep "+sleepSecsStr, false, false)
}

func (ac *AmiberryCommander) FormatCdImageCO(index int, pathname string) (string, string) {
	key := fmt.Sprintf("cdimage%v", index)
	value := pathname + ",image"

	return key, value
}

func (ac *AmiberryCommander) PutCdImageCO(index int, pathname string) {
	key, value := ac.FormatCdImageCO(index, pathname)

	ac.PutSetConfigOptionCommand(key, value)
}

func (ac *AmiberryCommander) FormatHardFile2CO(
	driveIndex int,
	pathname string,
	sectors int,
	surfaces int,
	reserved int,
	blockSize int,
	bootPriority int,
	controllerIndex int) (string, string) {
	key := "hardfile2"
	value := fmt.Sprintf("rw,DH%v:%v,%v,%v,%v,%v,%v,,ide%v_mainboard,0", driveIndex, pathname, sectors, surfaces, reserved, blockSize, bootPriority, controllerIndex)

	return key, value
}

func (ac *AmiberryCommander) FormatUaeHfCO(
	driveIndex int,
	pathname string,
	sectors int,
	surfaces int,
	reserved int,
	blockSize int,
	bootPriority int,
	controllerIndex int) (string, string) {
	key := fmt.Sprintf("uaehf%v", driveIndex)
	value := fmt.Sprintf("hdf,rw,DH%v:%v,%v,%v,%v,%v,%v,,ide%v_mainboard,0", driveIndex, pathname, sectors, surfaces, reserved, blockSize, bootPriority, controllerIndex)

	return key, value
}

// Enable/disable floppy sound
func (ac *AmiberryCommander) FormatFloppySoundConfigOption(index int, enable bool) (string, string) {
	value := "0"

	if enable {
		value = "1"
	}

	key := fmt.Sprintf("floppy%vsound", index)

	return key, value
}

func (ac *AmiberryCommander) PutFloppySoundCO(index int, enable bool) {
	key, value := ac.FormatFloppySoundConfigOption(index, enable)

	ac.PutSetConfigOptionCommand(key, value)
}

func (ac *AmiberryCommander) FormatFloppySoundVolumeDiskCO(index, volume int) (string, string) {
	key := fmt.Sprintf("floppy%vsoundvolume_disk", index)
	value := fmt.Sprintf("%v", 100-volume)

	return key, value
}

func (ac *AmiberryCommander) PutFloppySoundVolumeDiskCO(index, volume int) {
	key, value := ac.FormatFloppySoundVolumeDiskCO(index, volume)

	ac.PutSetConfigOptionCommand(key, value)
}
