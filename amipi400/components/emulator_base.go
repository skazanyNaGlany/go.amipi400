package components

import (
	components_base "github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/consts"
)

type EmulatorBase struct {
	components_base.RunnerBase

	executablePathname string
	configPathname     string
	adfs               [consts.MAX_ADFS]string
}

func (eb *EmulatorBase) Run() {
	panic("unimplemented")
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
	eb.executablePathname = pathname
}

func (eb *EmulatorBase) GetExecutablePathname() string {
	return eb.executablePathname
}

func (eb *EmulatorBase) SetConfigPathname(pathname string) {
	eb.configPathname = pathname
}

func (eb *EmulatorBase) GetConfigPathname() string {
	return eb.configPathname
}

func (eb *EmulatorBase) AttachAdf(index int, pathname string) error {
	eb.adfs[index] = pathname

	return nil
}

func (eb *EmulatorBase) DetachAdf(index int) error {
	eb.adfs[index] = ""

	return nil
}
