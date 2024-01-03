package components

import (
	"os"

	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
	"github.com/subpop/go-ini"
)

type MainConfig struct {
	pathname string `ini:"-"`

	AmiPi400 struct {
		Zoom            bool   `ini:"zoom"`
		WIFIManage      bool   `ini:"wifi_manage"`
		WIFICountryCode string `ini:"wifi_country_code"`
		WIFISSID        string `ini:"wifi_ssid"`
		WIFIPassword    string `ini:"wifi_password"`
	} `ini:"amipi400"`
}

func NewMainConfig(pathname string) *MainConfig {
	mc := MainConfig{}
	mc.pathname = pathname

	return &mc
}

func (mc *MainConfig) Load() error {
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

func (mc *MainConfig) Save() error {
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
