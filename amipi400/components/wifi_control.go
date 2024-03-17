package components

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
	"github.com/skazanyNaGlany/go.amipi400/shared/interfaces"
	"github.com/thoas/go-funk"
)

type WIFIControl struct {
	components.RunnerBase

	operations       chan map[string]any
	thisType         string // TODO move to RunnerBase
	lastError        error
	pendingOperation bool
}

func NewWIFIControl() *WIFIControl {
	wc := WIFIControl{
		operations: make(chan map[string]any)}

	return &wc
}

func (wc *WIFIControl) loop() {
	for op := range wc.operations {
		wc.lastError = nil

		wc.pendingOperation = true

		if op["type"] == shared.WIFI_CONTROL_OP_DISCONNECT {
			wc.lastError = wc.disconnect()
		} else if op["type"] == shared.WIFI_CONTROL_OP_CONNECT {
			wc.lastError = wc.connect(
				op["countryCode"].(string),
				op["ssid"].(string),
				op["password"].(string))
		}

		if wc.lastError != nil {
			if wc.IsDebugMode() {
				wc.logPrintLn(wc.lastError)
			}
		}

		wc.pendingOperation = false
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
	op["countryCode"] = countryCode // ISO/IEC 3166-1 alpha2 (all capitalized)
	op["ssid"] = ssid
	op["password"] = password

	wc.operations <- op
}

func (wc *WIFIControl) disconnect() error {
	wc.logPrintLn("disconnecting from WIFI")

	ifaces, err := wc.Interfaces()

	if err != nil {
		return err
	}

	if len(ifaces) == 0 {
		return errors.New("no wireless interfaces")
	}

	firstInterfaceName := funk.Keys(ifaces).([]string)[0]

	if ifaces[firstInterfaceName] == "off/any" {
		return errors.New("not connected to any WIFI network")
	}

	if _, err = exec.Command("killall", "-9", "wpa_supplicant").CombinedOutput(); err != nil {
		return errors.New("kill wpa_supplicant: " + err.Error())
	}

	if _, err = exec.Command("ifconfig", firstInterfaceName, "down").CombinedOutput(); err != nil {
		return errors.New("ifconfig: " + err.Error())
	}

	if _, err = exec.Command("ifconfig", firstInterfaceName, "up").CombinedOutput(); err != nil {
		return errors.New("ifconfig: " + err.Error())
	}

	wc.logPrintLn("WIFI disconnected")

	return nil
}

func (wc *WIFIControl) Interfaces() (map[string]string, error) {
	ifaces := make(map[string]string)

	output, err := exec.Command("iwconfig").CombinedOutput()

	if err != nil {
		return nil, errors.New("cannot run iwconfig")
	}

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)

		matches := utils.RegExInstance.FindNamedMatches(
			shared.IWCONFIG_INTERFACE_TO_SSID_RE,
			line,
		)

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

func (wc *WIFIControl) connect(countryCode string, ssid string, password string) error {
	ifaces, err := wc.Interfaces()

	if err != nil {
		return err
	}

	if len(ifaces) == 0 {
		return errors.New("no wireless interfaces")
	}

	firstInterfaceName := funk.Keys(ifaces).([]string)[0]

	if ifaces[firstInterfaceName] != "off/any" {
		return errors.New(
			"WIFI is connected to " + ifaces[firstInterfaceName] + " network, disconnect first",
		)
	}

	if _, err := exec.Command("iw", "reg", "set", countryCode).CombinedOutput(); err != nil {
		return errors.New("iw: " + err.Error())
	}

	passphrase, err := exec.Command("wpa_passphrase", ssid, password).CombinedOutput()

	if err != nil {
		return errors.New("wpa_passphrase: " + err.Error())
	}

	passphraseStr := string(passphrase)
	passphraseStr = strings.TrimSpace(passphraseStr)
	passphraseStr += "\n\nctrl_interface=/var/run/wpa_supplicant\n\n"

	if _, err = utils.FileUtilsInstance.FileWriteBytes(
		shared.WPA_SUPPLICANT_CONF_PATHNAME,
		0,
		[]byte(passphraseStr),
		os.O_CREATE|os.O_RDWR|os.O_TRUNC,
		0644,
		nil); err != nil {
		return err
	}

	if _, err := exec.Command("killall", "-9", "wpa_supplicant").CombinedOutput(); err != nil {
		wc.logPrintLn("kill wpa_supplicant: " + err.Error())
	}

	if _, err := exec.Command("rfkill", "unblock", "wifi").CombinedOutput(); err != nil {
		return errors.New("rfkill: " + err.Error())
	}

	if _, err := exec.Command("wpa_supplicant", "-B", "-c", shared.WPA_SUPPLICANT_CONF_PATHNAME, "-i", firstInterfaceName).CombinedOutput(); err != nil {
		wc.logPrintLn("wpa_supplicant: " + err.Error())
	}

	wc.logPrintLn("WIFI connected")

	return nil
}

func (wc *WIFIControl) logPrintLn(v ...any) {
	v2 := make([]any, 0)

	v2 = append(v2, wc.thisType)
	v2 = append(v2, v...)

	log.Println(v2...)
}

func (wc *WIFIControl) GetLastError() error {
	return wc.lastError
}

func (wc *WIFIControl) Wait() {
	for {
		if len(wc.operations) == 0 {
			break
		}

		if !wc.pendingOperation {
			break
		}

		time.Sleep(time.Millisecond * 10)
	}
}

func (wc *WIFIControl) WaitUntilConnected(seconds int) error {
	for {
		ifaces, err := wc.Interfaces()

		if err != nil {
			return err
		}

		if len(ifaces) == 0 {
			return errors.New("no wireless interfaces")
		}

		firstInterfaceName := funk.Keys(ifaces).([]string)[0]

		if ifaces[firstInterfaceName] != "off/any" {
			break
		}

		seconds--
		time.Sleep(time.Second * 1)
	}

	return nil
}

func (wc *WIFIControl) WaitUntilDisconnected(seconds int) error {
	for {
		ifaces, err := wc.Interfaces()

		if err != nil {
			return err
		}

		if len(ifaces) == 0 {
			return errors.New("no wireless interfaces")
		}

		firstInterfaceName := funk.Keys(ifaces).([]string)[0]

		if ifaces[firstInterfaceName] == "off/any" {
			break
		}

		seconds--
		time.Sleep(time.Second * 1)
	}

	return nil
}
