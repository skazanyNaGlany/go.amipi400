package components

import (
	"log"

	"github.com/skazanyNaGlany/go.amipi400/shared/interfaces"
)

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

func (rb *RunnerBase) IsVerboseMode() bool {
	return rb.verboseMode
}

func (rb *RunnerBase) IsDebugMode() bool {
	return rb.debugMode
}

func (rb *RunnerBase) SetRunning(running bool) {
	rb.running = running
}
