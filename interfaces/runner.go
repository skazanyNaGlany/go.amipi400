package interfaces

type Runner interface {
	Run()
	Start(_runner Runner) error
	Stop(_runner Runner) error
	IsRunning() bool
	SetVerboseMode(verboseMode bool)
	SetDebugMode(debugMode bool)
	IsVerboseMode() bool
	IsDebugMode() bool
}
