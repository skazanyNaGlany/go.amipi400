package interfaces

type Emulator interface {
	Pause() error
	Resume() error
	SoftReset() error
	HardReset() error
	SetExecutablePathname(pathname string)
	GetExecutablePathname() string
	SetConfigPathname(pathname string)
	GetConfigPathname() string
	GetEmulatorCommandLine() []string
}
