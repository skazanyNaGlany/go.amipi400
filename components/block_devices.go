package components

import (
	"errors"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/interfaces"
)

var lsblkPattern *regexp.Regexp = regexp.MustCompile(`NAME="(?P<NAME>\w*)" SIZE="(?P<SIZE>\d*)" TYPE="(?P<TYPE>\w*)" MOUNTPOINT="(?P<MOUNTPOINT>.*)" LABEL="(?P<LABEL>.*)" PATH="(?P<PATH>.*)" FSTYPE="(?P<FSTYPE>.*)" PTTYPE="(?P<PTTYPE>.*)" RO="(?P<RO>.*)"`)

type BlockDevices struct {
	running          bool
	attachedHandlers []interfaces.AttachedBlockDeviceCallback
	detachedHandlers []interfaces.DetachedBlockDeviceCallback
}

func (bd *BlockDevices) loop() {
	old_parsed_output := make(map[string]map[string]string)

	for bd.running {
		time.Sleep(time.Millisecond * 10)

		// lsblk -P -o name,size,type,mountpoint,label,path,fstype,pttype,ro -n -b
		output, err := exec.Command(
			"lsblk",
			"-P",
			"-o",
			"name,size,type,mountpoint,label,path,fstype,pttype,ro",
			"-n",
			"-b").CombinedOutput()

		if err != nil {
			log.Println("lsblk:", err)
			break
		}

		parsed_output, err := bd.parseLsblkOutput(string(output))

		if err != nil {
			log.Println("lsblk:", err)
			break
		}

		err = bd.notifyHandlers(old_parsed_output, parsed_output)

		if err != nil {
			log.Println("lsblk:", err)
			break
		}

		old_parsed_output = parsed_output
	}

	bd.running = false
}

func (bd *BlockDevices) notifyHandlers(old_block_devices, block_devices map[string]map[string]string) error {
	err := bd.notifyDetachedHandlers(old_block_devices, block_devices)

	if err != nil {
		return err
	}

	return bd.notifyAttachedHandlers(old_block_devices, block_devices)
}

func (bd *BlockDevices) notifyAttachedHandlers(old_block_devices, block_devices map[string]map[string]string) error {
	for name := range block_devices {
		new_block_device_data := block_devices[name]
		old_block_device_data, exists := old_block_devices[name]

		if exists {
			property_changed := bd.blockDevicePropertyChanged(
				old_block_device_data,
				new_block_device_data)

			exists = !property_changed
		}

		if !exists {
			converted, err := bd.convertDataMap(new_block_device_data)

			if err != nil {
				return err
			}

			for _, handler := range bd.attachedHandlers {
				handler(
					converted["NAME"].(string),
					converted["SIZE"].(uint64),
					converted["TYPE"].(string),
					converted["MOUNTPOINT"].(string),
					converted["LABEL"].(string),
					converted["PATH"].(string),
					converted["FSTYPE"].(string),
					converted["PTTYPE"].(string),
					converted["RO"].(bool))
			}
		}
	}

	return nil
}

func (bd *BlockDevices) notifyDetachedHandlers(old_block_devices, block_devices map[string]map[string]string) error {
	for name := range old_block_devices {
		old_block_device_data := old_block_devices[name]
		new_block_device_data, exists := block_devices[name]

		if exists {
			property_changed := bd.blockDevicePropertyChanged(
				old_block_device_data,
				new_block_device_data)

			exists = !property_changed
		}

		if !exists {
			converted, err := bd.convertDataMap(old_block_device_data)

			if err != nil {
				return err
			}

			for _, handler := range bd.detachedHandlers {
				handler(
					converted["NAME"].(string),
					converted["SIZE"].(uint64),
					converted["TYPE"].(string),
					converted["MOUNTPOINT"].(string),
					converted["LABEL"].(string),
					converted["PATH"].(string),
					converted["FSTYPE"].(string),
					converted["PTTYPE"].(string),
					converted["RO"].(bool))
			}
		}
	}

	return nil
}

func (*BlockDevices) blockDevicePropertyChanged(
	current_block_device_data map[string]string,
	new_block_device_data map[string]string) bool {
	property_changed := current_block_device_data["SIZE"] != new_block_device_data["SIZE"] ||
		current_block_device_data["TYPE"] != new_block_device_data["TYPE"] ||
		current_block_device_data["LABEL"] != new_block_device_data["LABEL"] ||
		current_block_device_data["FSTYPE"] != new_block_device_data["FSTYPE"] ||
		current_block_device_data["PTTYPE"] != new_block_device_data["PTTYPE"] ||
		current_block_device_data["RO"] != new_block_device_data["RO"]

	return property_changed
}

func (bd *BlockDevices) convertDataMap(data map[string]string) (map[string]any, error) {
	converted := make(map[string]any)

	size, err := strconv.ParseUint(data["SIZE"], 10, 64)

	if err != nil {
		return nil, err
	}

	converted["SIZE"] = size

	// read-only
	ro, err := strconv.Atoi(data["RO"])

	if err != nil {
		return nil, err
	}

	converted["RO"] = ro != 0

	converted["NAME"] = data["NAME"]
	converted["TYPE"] = data["TYPE"]
	converted["MOUNTPOINT"] = data["MOUNTPOINT"]
	converted["LABEL"] = data["LABEL"]
	converted["PATH"] = data["PATH"]
	converted["FSTYPE"] = data["FSTYPE"]
	converted["PTTYPE"] = data["PTTYPE"]

	return converted, nil
}

func (bd *BlockDevices) parseLsblkOutput(output string) (map[string]map[string]string, error) {
	parsed := make(map[string]map[string]string)

	output_lines := strings.Split(output, "\n")

	for _, line := range output_lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		matches := regExUtils.FindNamedMatches(lsblkPattern, line)

		if len(matches) != 9 {
			return nil, errors.New("cannot parse line: " + line)
		}

		parsed[matches["NAME"]] = matches
	}

	return parsed, nil
}

func (bd *BlockDevices) AddAttachedHandler(handler interfaces.AttachedBlockDeviceCallback) {
	bd.attachedHandlers = append(bd.attachedHandlers, handler)
}

func (bd *BlockDevices) AddDetachedHandler(handler interfaces.DetachedBlockDeviceCallback) {
	bd.detachedHandlers = append(bd.detachedHandlers, handler)
}

func (bd *BlockDevices) Start() error {
	log.Printf("Starting BlockDevices %p\n", bd)

	bd.running = true

	go bd.loop()

	return nil
}

func (bd *BlockDevices) Stop() error {
	log.Printf("Stopping BlockDevices %p\n", bd)

	bd.running = false

	return nil
}

func (bd *BlockDevices) IsRunning() bool {
	return bd.running
}
