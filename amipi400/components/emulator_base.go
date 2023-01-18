package components

import (
	components_base "github.com/skazanyNaGlany/go.amipi400/components"
)

type EmulatorBase struct {
	components_base.RunnerBase
}

func (eb *EmulatorBase) Pause() error {
	panic("unimplemented")
}

func (eb *EmulatorBase) Resume() error {
	panic("unimplemented")
}

func (eb *EmulatorBase) SoftReset() error {
	panic("unimplemented")
}

func (eb *EmulatorBase) HardReset() error {
	panic("unimplemented")
}

func (eb *EmulatorBase) SetExecutablePathname(pathname string) {
	panic("unimplemented")
}

func (eb *EmulatorBase) GetExecutablePathname() string {
	panic("unimplemented")
}

func (eb *EmulatorBase) SetConfigPathname(pathname string) {
	panic("unimplemented")
}

func (eb *EmulatorBase) GetConfigPathname() string {
	panic("unimplemented")
}
