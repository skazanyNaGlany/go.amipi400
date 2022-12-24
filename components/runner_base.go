package components

import "github.com/skazanyNaGlany/go.amipi400/interfaces"

import "log"

type RunnerBase struct {
	running     bool
	verboseMode bool
	debugMode   bool
}

func (rb *RunnerBase) Run() {}

func (rb *RunnerBase) Start(_runner interfaces.Runner) error {
	if rb.verboseMode {
		log.Printf("Starting %T %p\n", _runner, &_runner)
	}

	rb.running = true

	go _runner.Run()

	return nil
}

func (rb *RunnerBase) Stop(_runner interfaces.Runner) error {
	if rb.verboseMode {
		log.Printf("Stopping %T %p\n", _runner, &_runner)
	}

	rb.running = false

	return nil
}

func (rb *RunnerBase) IsRunning() bool {
	return rb.running
}

func (rb *RunnerBase) SetVerboseMode(verboseMode bool) {
	rb.verboseMode = verboseMode
}

func (rb *RunnerBase) SetDebugMode(debugMode bool) {
	rb.debugMode = debugMode
}

func (rb *RunnerBase) GetVerboseMode() bool {
	return rb.verboseMode
}

func (rb *RunnerBase) GetDebugMode() bool {
	return rb.debugMode
}
