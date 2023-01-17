package amiga_disk_devices

import (
	"errors"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type DriveDevicesDiscovery struct {
	floppies []string
	cdroms   []string
}

func (ddd *DriveDevicesDiscovery) Refresh() error {
	if err := ddd.RefreshFloppies(); err != nil {
		return err
	}

	if err := ddd.RefreshCDROMs(); err != nil {
		return err
	}

	return nil
}

func (ddd *DriveDevicesDiscovery) RefreshFloppies() error {
	floppies := make([]string, 0)

	output, err := exec.Command("ufiformat", "--inquire", "--quiet").CombinedOutput()

	if err != nil {
		return err
	}

	output_lines := strings.Split(string(output), "\n")

	for _, line := range output_lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		line = strings.ReplaceAll(line, "  ", "")
		line_parts := strings.Split(line, " ")

		if len(line_parts) != 2 {
			return errors.New("Unable to parse ufiformat line: " + line)
		}

		devicePathname := strings.TrimSpace(line_parts[0])

		stat, err := os.Stat(devicePathname)

		if err != nil {
			return err
		}

		if !stat.IsDir() {
			floppies = append(floppies, devicePathname)
		}
	}

	sort.Strings(floppies)

	ddd.floppies = floppies

	return nil
}

func (ddd *DriveDevicesDiscovery) GetFloppies() []string {
	return ddd.floppies
}

func (ddd *DriveDevicesDiscovery) RefreshCDROMs() error {
	cdroms := make([]string, 0)

	output, err := exec.Command("hwinfo", "--cdrom", "--short").CombinedOutput()

	if err != nil {
		return err
	}

	output_lines := strings.Split(string(output), "\n")
	cdrom_data_started := false

	for _, line := range output_lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if line == "cdrom:" {
			cdrom_data_started = true
			continue
		}

		if !cdrom_data_started {
			continue
		}

		if !strings.HasPrefix(line, "/dev/") {
			continue
		}

		line = strings.ReplaceAll(line, "  ", "")
		line_parts := strings.SplitN(line, " ", 2)

		if len(line_parts) != 2 {
			return errors.New("Unable to parse hwinfo line: " + line)
		}

		devicePathname := strings.TrimSpace(line_parts[0])

		stat, err := os.Stat(devicePathname)

		if err != nil {
			return err
		}

		if !stat.IsDir() {
			cdroms = append(cdroms, devicePathname)
		}
	}

	sort.Strings(cdroms)

	ddd.cdroms = cdroms

	return nil
}

func (ddd *DriveDevicesDiscovery) GetCDROMs() []string {
	return ddd.cdroms
}
