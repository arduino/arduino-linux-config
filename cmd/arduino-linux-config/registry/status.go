package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/arduino/arduino-linux-config/cmd/config"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/go-paths-helper"
)

type StatusFile struct {
	DeviceName string
	Option     string
	CreatedAt  time.Time
	Path       *paths.Path
}

type StatusDevice struct {
	Device string `json:"device"`
	Option string `json:"option"`
}

func LoadStatus(cfg config.Configuration) ([]StatusFile, error) {
	entries, err := cfg.StatusDir().ReadDir()
	if err != nil {
		return nil, fmt.Errorf("failed to read overlays dir: %w", err)
	}

	const layout = "20060102-150405"
	files := make([]StatusFile, 0, len(entries))

	for _, entry := range entries {
		name := entry.Base() // e.g. "camera1_20260313-150945.dtbo"
		nameNoExt := strings.TrimSuffix(name, filepath.Ext(name))

		parts := strings.SplitN(nameNoExt, "_", 2)
		if len(parts) != 2 {
			continue
		}

		deviceParts := strings.SplitN(parts[0], "-", 2)
		if len(deviceParts) != 2 {
			continue
		}

		device := deviceParts[0] // "camera1"
		option := deviceParts[1] // "type1-2lane"

		createdAt, err := time.Parse(layout, parts[1])
		if err != nil {
			continue
		}

		files = append(files, StatusFile{
			DeviceName: device,
			Option:     option,
			CreatedAt:  createdAt,
			Path:       entry,
		})
	}

	return files, nil
}

func CreateStatusFile(cfg config.Configuration, deviceName string) error {
	const layout = "20060102-150405"
	timestamp := time.Now().UTC().Format(layout)
	filename := fmt.Sprintf("%s_%s.dtbo", deviceName, timestamp)
	return cfg.StatusDir().Join(filename).WriteFile([]byte{})
}

// cleanOldStates removes all overlay files for the specified device that were created after the last boot time.
func CleanOldStatus(deviceName string, files []StatusFile) error {
	bootTime, err := getBootTime()
	if err != nil {
		return fmt.Errorf("failed to get boot time: %w", err)
	}

	var deviceFiles []StatusFile
	for _, f := range files {
		if f.DeviceName == deviceName {
			deviceFiles = append(deviceFiles, f)
		}
	}

	var surviving []StatusFile
	for _, f := range deviceFiles {
		if f.CreatedAt.After(bootTime) {
			fmt.Printf("Removing %s-%s-%s boot %s\n", f.DeviceName, f.Option, f.CreatedAt, bootTime)
			if err := f.Path.Remove(); err != nil {
				return fmt.Errorf("failed to remove post-boot overlay file %s: %w", f.Path, err)
			}
		} else {
			surviving = append(surviving, f)
		}
	}

	if len(surviving) > 1 {
		sort.Slice(surviving, func(i, j int) bool {
			return surviving[i].CreatedAt.After(surviving[j].CreatedAt)
		})
		for _, f := range surviving[1:] {
			if err := f.Path.Remove(); err != nil {
				return fmt.Errorf("failed to remove old overlay file %s: %w", f.Path, err)
			}
		}
	}

	return nil
}

func getBootTime() (time.Time, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read /proc/uptime: %w", err)
	}

	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return time.Time{}, fmt.Errorf("unexpected /proc/uptime format")
	}

	uptimeSeconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse uptime: %w", err)
	}

	bootTime := time.Now().UTC().Add(-time.Duration(uptimeSeconds * float64(time.Second)))

	return bootTime, nil
}

// TODO improve error messages
func StatusUpdate(cfg config.Configuration, statusUpdate map[MediaCarrierDeviceName]string) {

	fileStatusList, err := LoadStatus(cfg)
	if err != nil {
		feedback.Warnf(err.Error(), feedback.ErrGeneric)
	}

	for device, optionValue := range statusUpdate {
		statusFileName := fmt.Sprintf("%s-%s", device, optionValue)

		if err := CreateStatusFile(cfg, statusFileName); err != nil {
			feedback.Warnf(err.Error(), feedback.ErrGeneric)
		}

		if err := CleanOldStatus(string(device), fileStatusList); err != nil {
			feedback.Warnf(err.Error(), feedback.ErrGeneric)
		}
	}
}

func GetStatuses(cfg config.Configuration) ([]StatusDevice, []StatusDevice) {
	CleanDuplicated(cfg)

	bootTime, err := getBootTime()
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get boot time: %v", err), feedback.ErrGeneric)
	}

	statusList, err := LoadStatus(cfg)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to read carrier status: %v", err), feedback.ErrGeneric)
	}

	current := []StatusDevice{}
	next := []StatusDevice{}

	for _, f := range statusList {
		if f.CreatedAt.Before(bootTime) {
			current = append(current, StatusDevice{Device: f.DeviceName, Option: f.Option})
		} else {
			next = append(next, StatusDevice{Device: f.DeviceName, Option: f.Option})
		}
	}
	return current, next
}

func CleanDuplicated(cfg config.Configuration) error {
	bootTime, err := getBootTime()
	if err != nil {
		return fmt.Errorf("failed to get boot time: %w", err)
	}

	statusList, err := LoadStatus(cfg)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to read carrier status: %v", err), feedback.ErrGeneric)
	}

	// filter pre-boot file and group by device name
	grouped := make(map[string][]StatusFile)
	for _, f := range statusList {
		if f.CreatedAt.Before(bootTime) {
			grouped[f.DeviceName] = append(grouped[f.DeviceName], f)
		}
	}

	// process by groups
	for _, files := range grouped {
		// newest first
		slices.SortFunc(files, func(a, b StatusFile) int {
			return b.CreatedAt.Compare(a.CreatedAt)
		})

		if len(files) > 1 {
			deleteFiles(files[1:])
		}
	}
	return nil
}

func deleteFiles(files []StatusFile) {
	for _, f := range files {
		err := os.Remove(f.Path.String())
		if err != nil && !os.IsNotExist(err) {
			fmt.Printf("Failed to delete %s: %v\n", f.Path.String(), err)
		}
	}
}
