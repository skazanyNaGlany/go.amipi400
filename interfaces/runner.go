package interfaces

type Runner interface {
	Run()
	Start(_runner Runner) error
	Stop() error
	IsRunning() bool
	SetVerboseMode(verboseMode bool)
	SetDebugMode(debugMode bool)
	GetVerboseMode() bool
	GetDebugMode() bool
}
