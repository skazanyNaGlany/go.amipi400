package interfaces

type Runner interface {
	Start() error
	Stop() error
	IsRunning() bool
}
