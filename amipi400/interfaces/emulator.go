package interfaces

type Emulator interface {
	Run() error
	Stop() error
	Pause() error
	Resume() error
	SoftReset() error
	HardReset() error
}
