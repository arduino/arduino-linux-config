package registry

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/arduino/arduino-linux-config/cmd/config"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/go-paths-helper"
)

type StatusFile struct {
	CurrentStatus StatusCarrier `json:"current_status"`
	NextStatus    StatusCarrier `json:"next_status"`
}

type StatusCarrier struct {
	Devices map[CarrierDeviceName]StatusInfo `json:"devices"`
}

type StatusInfo struct {
	Option    string    `json:"option"`
	CreatedAt time.Time `json:"created_at"`
}

type StatusDevice struct {
	Device string `json:"device"`
	Option string `json:"option"`
}

func StatusUpdate(cfg config.Configuration, carrierName string, statusUpdate map[CarrierDeviceName]string) {
	status, err := loadStatusFile(getStatusFile(cfg, carrierName))
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to load status file %v", err), feedback.ErrGeneric)
	}

	updateStatusStructure(status, carrierName, statusUpdate)
	if err := saveStatusFile(getStatusFile(cfg, carrierName), *status); err != nil {
		feedback.Fatal(fmt.Sprintf("failed to save status file: %v", err), feedback.ErrGeneric)
	}
}

func GetStatus(cfg config.Configuration, carrierName string) ([]StatusDevice, []StatusDevice) {
	status, err := loadStatusFile(getStatusFile(cfg, carrierName))
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to load status file %v", err), feedback.ErrGeneric)
	}

	bootTime, err := getBootTime()
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get boot time: %v", err), feedback.ErrGeneric)
	}

	return getStatusStructure(status, carrierName, bootTime)
}

func getStatusStructure(status *StatusFile, carrierName string, bootTime time.Time) ([]StatusDevice, []StatusDevice) {
	deviceNames := GetDevicesNames(carrierName)

	current := make([]StatusDevice, 0, len(deviceNames))
	next := make([]StatusDevice, 0, len(deviceNames))

	for _, deviceName := range deviceNames {

		// parse next, if they happened before the boot, move in actual
		if status.NextStatus.Devices[deviceName].CreatedAt.Before(bootTime) {
			status.CurrentStatus.Devices[deviceName] = status.NextStatus.Devices[deviceName]
			status.NextStatus.Devices[deviceName] = StatusInfo{}
		}
		current = append(current, StatusDevice{Device: string(deviceName), Option: getOrDefault(status.CurrentStatus.Devices[deviceName].Option)})
		next = append(next, StatusDevice{Device: string(deviceName), Option: getOrDefault(status.NextStatus.Devices[deviceName].Option)})
	}
	return current, next
}

func updateStatusStructure(status *StatusFile, carrierName string, statusUpdate map[CarrierDeviceName]string) {
	for _, deviceName := range GetDevicesNames(carrierName) {
		newInfo := StatusInfo{
			Option:    getOrDefault(statusUpdate[deviceName]),
			CreatedAt: time.Now().UTC(),
		}
		status.NextStatus.Devices[deviceName] = newInfo
	}
}

func getOrDefault(option string) string {
	return cmp.Or(option, string(None))
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

func getStatusFile(cfg config.Configuration, carrierName string) *paths.Path {
	filePath := filepath.Join(cfg.StatusDir().String(), carrierName+".json")
	return paths.New(filePath)
}

func loadStatusFile(statusFile *paths.Path) (*StatusFile, error) {
	data, err := statusFile.ReadFile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			newStatus := StatusFile{
				CurrentStatus: StatusCarrier{
					Devices: make(map[CarrierDeviceName]StatusInfo),
				},
				NextStatus: StatusCarrier{
					Devices: make(map[CarrierDeviceName]StatusInfo),
				},
			}
			return &newStatus, nil
		} else {
			return nil, err
		}
	}

	var status StatusFile
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, fmt.Errorf("could not parse json: %w", err)
	}

	return &status, nil
}

func saveStatusFile(statusFile *paths.Path, status StatusFile) error {
	data, err := json.MarshalIndent(status, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	// nolint:gosec // G306: Status file must be readable
	err = os.WriteFile(statusFile.String(), data, 0644)
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	return nil
}
