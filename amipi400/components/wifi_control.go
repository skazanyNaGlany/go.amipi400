package components

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
	"github.com/skazanyNaGlany/go.amipi400/shared/interfaces"
	"github.com/thoas/go-funk"
)

type WIFIControl struct {
	components.RunnerBase

	operations chan map[string]any
	thisType   string // TODO move to RunnerBase
}

func NewWIFIControl() *WIFIControl {
	wc := WIFIControl{
		operations: make(chan map[string]any)}

	return &wc
}

func (wc *WIFIControl) loop() {
	for op := range wc.operations {
		if op["type"] == shared.WIFI_CONTROL_OP_DISCONNECT {
			wc.disconnect()
		} else if op["type"] == shared.WIFI_CONTROL_OP_CONNECT {
			wc.connect(
				op["countryCode"].(string),
				op["ssid"].(string),
				op["password"].(string))
		}
	}

	wc.SetRunning(false)
}

func (wc *WIFIControl) Run() {
	wc.thisType = fmt.Sprintf("%T:", wc)

	wc.loop()
}

func (wc *WIFIControl) Stop(_runner interfaces.Runner) error {
	close(wc.operations)

	wc.RunnerBase.Stop(_runner)

	return nil
}

func (wc *WIFIControl) Disconnect() {
	op := make(map[string]any)

	op["type"] = shared.WIFI_CONTROL_OP_DISCONNECT

	wc.operations <- op
}

func (wc *WIFIControl) Connect(countryCode string, ssid string, password string) {
	op := make(map[string]any)

	op["type"] = shared.WIFI_CONTROL_OP_CONNECT
	op["countryCode"] = countryCode // ISO/IEC 3166-1 alpha2
	op["ssid"] = ssid
	op["password"] = password

	wc.operations <- op
}

func (wc *WIFIControl) disconnect() {
	wc.logPrintLn("disconnecting from WIFI")

	if err := os.Remove(shared.WPA_SUPPLICANT_CONF_PATHNAME); err != nil {
		wc.logPrintLn(err)
	}

	ifaces, err := wc.Interfaces()

	if err != nil {
		wc.logPrintLn(err)
		return
	}

	if len(ifaces) == 0 {
		wc.logPrintLn("no wireless interfaces")
		return
	}

	firstInterfaceName := funk.Keys(ifaces).([]string)[0]

	if ifaces[firstInterfaceName] == "off/any" {
		wc.logPrintLn("not connected to any WIFI network")
		return
	}

	if _, err = exec.Command("killall", "-9", "wpa_supplicant").CombinedOutput(); err != nil {
		wc.logPrintLn("killall -9 wpa_supplicant", err)
		return
	}

	if _, err = exec.Command("ifconfig", firstInterfaceName, "down").CombinedOutput(); err != nil {
		wc.logPrintLn("ifconfig", err)
		return
	}

	if _, err = exec.Command("ifconfig", firstInterfaceName, "up").CombinedOutput(); err != nil {
		wc.logPrintLn("ifconfig", err)
		return
	}

	wc.logPrintLn("WIFI disconnected")
}

func (wc *WIFIControl) Interfaces() (map[string]string, error) {
	ifaces := make(map[string]string)

	output, err := exec.Command("iwconfig").CombinedOutput()

	if err != nil {
		return nil, errors.New("cannot run iwconfig")
	}

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)

		matches := utils.RegExInstance.FindNamedMatches(shared.IWCONFIG_INTERFACE_TO_SSID_RE, line)

		if len(matches) == 0 {
			continue
		}

		name := strings.TrimSpace(matches["name"])
		ssid := strings.TrimSpace(matches["ssid"])

		ssid = strings.TrimPrefix(ssid, "\"")
		ssid = strings.TrimSuffix(ssid, "\"")
		ssid = strings.TrimSpace(ssid)

		ifaces[name] = ssid
	}

	return ifaces, nil
}

func (wc *WIFIControl) connect(countryCode string, ssid string, password string) {
	ifaces, err := wc.Interfaces()

	if err != nil {
		wc.logPrintLn(err)
		return
	}

	if len(ifaces) == 0 {
		wc.logPrintLn("no wireless interfaces")
		return
	}

	firstInterfaceName := funk.Keys(ifaces).([]string)[0]

	if ifaces[firstInterfaceName] != "off/any" {
		wc.logPrintLn("WIFI is connected to", ifaces[firstInterfaceName], "network, disconnect first")
		return
	}

	if _, err := exec.Command("iw", "reg", "set", countryCode).CombinedOutput(); err != nil {
		wc.logPrintLn("iw", err)
		return
	}

	passphrase, err := exec.Command("wpa_passphrase", ssid, password).CombinedOutput()

	if err != nil {
		wc.logPrintLn("wpa_passphrase", err)
		return
	}

	if _, err = utils.FileUtilsInstance.FileWriteBytes(
		shared.WPA_SUPPLICANT_CONF_PATHNAME,
		0,
		passphrase,
		os.O_CREATE|os.O_RDWR|os.O_TRUNC,
		0644,
		nil); err != nil {
		wc.logPrintLn(err)
		return
	}

	if _, err := exec.Command("killall", "-9", "wpa_supplicant").CombinedOutput(); err != nil {
		wc.logPrintLn("killall -9 wpa_supplicant", err)
	}

	if _, err := exec.Command("rfkill", "unblock", "wifi").CombinedOutput(); err != nil {
		wc.logPrintLn("rfkill", err)
		return
	}

	if _, err := exec.Command("wpa_supplicant", "-B", "-c", shared.WPA_SUPPLICANT_CONF_PATHNAME, "-i", firstInterfaceName).CombinedOutput(); err != nil {
		wc.logPrintLn("wpa_supplicant", err)
		return
	}

	wc.logPrintLn("WIFI connected")
}

func (wc *WIFIControl) logPrintLn(v ...any) {
	v2 := make([]any, 0)

	v2 = append(v2, wc.thisType)
	v2 = append(v2, v...)

	log.Println(v2...)
}
