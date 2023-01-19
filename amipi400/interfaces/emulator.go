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
	AttachAdf(index int, pathname string) error
	DetachAdf(index int) error
	GetAdf(index int) string
}
