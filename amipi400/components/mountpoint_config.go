package components

import (
	"os"

	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
	"github.com/subpop/go-ini"
)

type MountpointConfig struct {
	pathname string `ini:"-"`

	AmiPi400 struct {
		DefaultFile string `ini:"default_file"`
	} `ini:"amipi400"`
}

func NewMountpointConfig(pathname string) *MountpointConfig {
	mc := MountpointConfig{}
	mc.pathname = pathname

	return &mc
}

func (mc *MountpointConfig) Load() error {
	data, _, err := utils.FileUtilsInstance.FileReadBytes(
		mc.pathname,
		0,
		-1,
		0,
		0,
		nil)

	if err != nil {
		return err
	}

	if err := ini.Unmarshal(data, mc); err != nil {
		return err
	}

	return nil
}

func (mc *MountpointConfig) Save() error {
	data, err := ini.Marshal(mc)

	if err != nil {
		return err
	}

	_, err = utils.FileUtilsInstance.FileWriteBytes(
		mc.pathname,
		0,
		data,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0777,
		nil)

	if err != nil {
		return err
	}

	return nil
}
