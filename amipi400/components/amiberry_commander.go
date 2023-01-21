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

/*
def execute_commands():
    # TODO limit execution once every second
    global commands

    str_commands = ''
    index = 0

    while commands:
        icommand = commands.pop(0).strip()

        if icommand.startswith('local-'):
            if process_local_command(icommand, str_commands):
                str_commands = ''
                index = 0

            continue

        str_commands += 'cmd{index}={cmd}\n'.format(
            index=index,
            cmd=icommand
        )
        index += 1

    if str_commands:
        write_tmp_ini(str_commands)

        if send_SIGUSR1_signal():
            block_till_tmp_ini_exists()
*/

func (ac *AmiberryCommander) writeTmpIni(commands string) error {
	if ac.IsVerboseMode() {
		log.Println("Writing", ac.tmpIniPathname)
	}

	commands = "[commands]\n" + commands

	byteCommands := []byte(commands)

	n, err := utils.FileUtilsInstance.FileWriteBytes(
		ac.tmpIniPathname,
		0,
		byteCommands,
		os.O_CREATE|os.O_RDWR,
		0777,
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

/*
def process_local_command(command: str, str_commands: str):
    if command == 'local-commit':
        if not str_commands:
            return False

        print_log('Committing')

        write_tmp_ini(str_commands)

        if send_SIGUSR1_signal():
            block_till_tmp_ini_exists()

            return True

        return False
    elif command.startswith('local-sleep '):
        parts = command.split(' ')

        if len(parts) != 2:
            return False

        seconds = int(parts[1])

        print_log('Sleeping for {seconds} seconds'.format(
            seconds=seconds
        ))

        time.sleep(seconds)

    return False
*/

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
		log.Println(commands)
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

func (ac *AmiberryCommander) PutConfigChangedCommand() {
	ac.PutCommand("config_changed 1", false, false)
}

func (ac *AmiberryCommander) PutSetConfigOptionCommand(option string, value string) {
	full := fmt.Sprintf("cfgfile_parse_line_type_all %v=%v", option, value)

	ac.PutCommand(full, false, false)
}

func (ac *AmiberryCommander) PutSetFloppyCommand(index int, pathname string) {
	option := fmt.Sprintf("floppy%v", index)

	ac.PutSetConfigOptionCommand(option, pathname)
}

func (ac *AmiberryCommander) PutLocalCommitCommand() {
	ac.PutCommand("local-commit", false, false)
}

func (ac *AmiberryCommander) PutLocalSleepCommand(sleepSecs int) {
	sleepSecsStr := fmt.Sprintf("%v", sleepSecs)

	ac.PutCommand("local-sleep "+sleepSecsStr, false, false)
}

/*
// def put_command(command: str, reset: bool = False, force = False):
//     global commands

//     if reset:
//         commands = []

//     if is_emulator_paused and not force:
//         return

//     if command:
//         if commands:
//             if commands[len(commands) - 1] == command:
//                 # do not add same command
//                 return

//         commands.append(command)


// def put_local_commit_command(sleep_seconds: int = 0):
//     put_command('local-commit')

//     if sleep_seconds:
//         put_command('local-sleep 1')


// def send_SIGUSR1_signal():
//     if not AUTOSEND_SIGNAL:
//         return False

//     print_log('Sending SIGUSR1 signal to Amiberry emulator')

//     try:
//         sh.killall('-USR1', 'amiberry')

//         return True
//     except sh.ErrorReturnCode_1:
//         print_log('No process found')

//         return False


def process_local_command(command: str, str_commands: str):
    if command == 'local-commit':
        if not str_commands:
            return False

        print_log('Committing')

        write_tmp_ini(str_commands)

        if send_SIGUSR1_signal():
            block_till_tmp_ini_exists()

            return True

        return False
    elif command.startswith('local-sleep '):
        parts = command.split(' ')

        if len(parts) != 2:
            return False

        seconds = int(parts[1])

        print_log('Sleeping for {seconds} seconds'.format(
            seconds=seconds
        ))

        time.sleep(seconds)

    return False


// def write_tmp_ini(str_commands: str):
//     with open(emulator_tmp_ini_pathname, 'w+', newline=None) as f:
//         f.write('[commands]' + os.linesep)
//         f.write(str_commands)


// def block_till_tmp_ini_exists():
//     while os.path.exists(emulator_tmp_ini_pathname) and \
//         check_emulator_running():
//         time.sleep(0)


*/
