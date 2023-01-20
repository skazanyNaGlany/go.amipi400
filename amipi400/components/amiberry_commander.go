package components

import (
	"fmt"
	"os"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/components"
)

type AmiberryCommander struct {
	components.RunnerBase

	tmpIniPathname string
	process        *os.Process
	commands       []string
	emulatorPaused bool
	lastCommand    string
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

func (ac *AmiberryCommander) loop() {
	for ac.IsRunning() {
		time.Sleep(time.Millisecond * 10)

	}

	ac.SetRunning(false)
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

	if ac.lastCommand == command {
		// do not add the same command
		return
	}

	ac.commands = append(ac.commands, command)

	// TODO clear at execution
	ac.lastCommand = command
}

func (ac *AmiberryCommander) PutLocalCommitCommand(sleepSecs int) {
	ac.PutCommand("local-commit", false, false)

	if sleepSecs > 0 {
		sleepSecsStr := fmt.Sprintf("%v", sleepSecs)

		ac.PutCommand("local-sleep "+sleepSecsStr, false, false)
	}
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


def send_SIGUSR1_signal():
    if not AUTOSEND_SIGNAL:
        return False

    print_log('Sending SIGUSR1 signal to Amiberry emulator')

    try:
        sh.killall('-USR1', 'amiberry')

        return True
    except sh.ErrorReturnCode_1:
        print_log('No process found')

        return False


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


def write_tmp_ini(str_commands: str):
    with open(emulator_tmp_ini_pathname, 'w+', newline=None) as f:
        f.write('[commands]' + os.linesep)
        f.write(str_commands)


def block_till_tmp_ini_exists():
    while os.path.exists(emulator_tmp_ini_pathname) and \
        check_emulator_running():
        time.sleep(0)


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
