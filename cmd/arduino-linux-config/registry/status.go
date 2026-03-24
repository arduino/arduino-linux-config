package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/arduino/arduino-linux-config/cmd/config"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/go-paths-helper"
)

type StatusFile struct {
	CurrentStatus StatusCarrier `json:"CurrentStatus"`
	NextStatus    StatusCarrier `json:"NextStatus"`
}

type StatusCarrier struct {
	Devices map[MediaCarrierDeviceName]StatusInfo `json:"Devices"`
}

type StatusInfo struct {
	Option    string    `json:"Option"`
	CreatedAt time.Time `json:"CreatedAt"`
}

type StatusDevice struct {
	Device string `json:"device"`
	Option string `json:"option"`
}

func StatusUpdate(cfg config.Configuration, statusUpdate map[MediaCarrierDeviceName]string) {
	status, err := loadStatusFile(cfg.StatusFile())
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get boot time: %v", err), feedback.ErrGeneric)
	}

	updateStatusStructure(status, statusUpdate)
	if err := saveStatusFile(cfg.StatusFile(), *status); err != nil {
		feedback.Fatal(fmt.Sprintf("failed to save status file: %v", err), feedback.ErrGeneric)
	}
}

func GetStatus(cfg config.Configuration) ([]StatusDevice, []StatusDevice) {
	status, err := loadStatusFile(cfg.StatusFile())
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get boot time: %v", err), feedback.ErrGeneric)
	}

	bootTime, err := getBootTime()
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get boot time: %v", err), feedback.ErrGeneric)
	}

	return getStatusStructure(status, bootTime)
}

func getStatusStructure(status *StatusFile, bootTime time.Time) ([]StatusDevice, []StatusDevice) {
	current := []StatusDevice{}
	next := []StatusDevice{}
	for _, deviceName := range MediaCarrierDeviceList {

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

func updateStatusStructure(status *StatusFile, statusUpdate map[MediaCarrierDeviceName]string) {
	for _, deviceName := range MediaCarrierDeviceList {
		newInfo := StatusInfo{
			Option:    getOrDefault(statusUpdate[deviceName]),
			CreatedAt: time.Now().UTC(),
		}
		status.NextStatus.Devices[deviceName] = newInfo
	}
}

func getOrDefault(option string) string {
	if option == "" {
		option = DeviceOptionNone
	}
	return option
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

func loadStatusFile(statusFile *paths.Path) (*StatusFile, error) {
	data, err := statusFile.ReadFile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			newStatus := StatusFile{
				CurrentStatus: StatusCarrier{
					Devices: make(map[MediaCarrierDeviceName]StatusInfo),
				},
				NextStatus: StatusCarrier{
					Devices: make(map[MediaCarrierDeviceName]StatusInfo),
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

	err = os.WriteFile(statusFile.String(), data, 0600)
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	return nil
}
