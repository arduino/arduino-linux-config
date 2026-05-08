package status

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/go-paths-helper"
)

type StatusFile struct {
	CurrentStatus StatusCarrier `json:"current_status"`
	NextStatus    StatusCarrier `json:"next_status"`
}

type StatusCarrier struct {
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"created_at"`

	Devices map[registry.CarrierDeviceName]StatusInfo `json:"devices"`
}

type StatusInfo struct {
	Option    string    `json:"option"`
	CreatedAt time.Time `json:"created_at"`
}

type CarrierStatus struct {
	Enable        bool
	StatusDevices []StatusDevice
}

type StatusDevice struct {
	Device string
	Option string
}

// Called by config and reset
func Update(cfg config.Configuration, carrier registry.Carrier, statusUpdate CarrierStatus) error {
	status, err := loadStatusFile(getStatusFile(cfg, carrier.Name))
	if err != nil {
		return fmt.Errorf("failed to load status file %w", err)
	}

	// save current state if reboot occurred
	bootTime, err := getBootTime()
	if err != nil {
		return fmt.Errorf("failed to get boot time: %w", err)
	}
	currentStatus, _ := getStatusStructure(status, carrier, bootTime)
	updateStatusStructure(status, carrier, currentStatus, statusUpdate)

	if err := saveStatusFile(getStatusFile(cfg, carrier.Name), *status); err != nil {
		return fmt.Errorf("failed to save status file: %w", err)
	}
	return nil
}

// Called by show, load the status structure and apply status fixes before returning
// Do not update the status file on the disk because show is running as non-root user
func Get(cfg config.Configuration, carrier registry.Carrier) (CarrierStatus, CarrierStatus, error) {
	status, err := loadStatusFile(getStatusFile(cfg, carrier.Name))
	if err != nil {
		return CarrierStatus{}, CarrierStatus{}, fmt.Errorf("failed to load status file %v", err)
	}

	bootTime, err := getBootTime()
	if err != nil {
		return CarrierStatus{}, CarrierStatus{}, fmt.Errorf("failed to get boot time: %v", err)
	}

	current, next := getStatusStructure(status, carrier, bootTime)
	return current, next, nil
}

func getStatusStructure(status *StatusFile, carrier registry.Carrier, bootTime time.Time) (CarrierStatus, CarrierStatus) {
	current := CarrierStatus{
		Enable:        false,
		StatusDevices: make([]StatusDevice, 0, len(carrier.Devices)),
	}
	next := CarrierStatus{
		Enable:        false,
		StatusDevices: make([]StatusDevice, 0, len(carrier.Devices)),
	}

	for _, device := range carrier.Devices {
		// parse next structure, move in actual events happened before the boot
		if status.NextStatus.Devices[device.Name].CreatedAt.Before(bootTime) {
			status.CurrentStatus.Devices[device.Name] = status.NextStatus.Devices[device.Name]
		}
		current.StatusDevices = append(current.StatusDevices, StatusDevice{
			Device: string(device.Name),
			Option: getOrDefault(status.CurrentStatus.Devices[device.Name].Option),
		})
		next.StatusDevices = append(next.StatusDevices, StatusDevice{
			Device: string(device.Name),
			Option: getOrDefault(status.NextStatus.Devices[device.Name].Option),
		})
	}

	if status.NextStatus.CreatedAt.Before(bootTime) {
		status.CurrentStatus.Status = status.NextStatus.Status
	}

	next.Enable = status.NextStatus.Status
	current.Enable = status.CurrentStatus.Status

	return current, next
}

func updateStatusStructure(status *StatusFile, carrier registry.Carrier, currentStatus CarrierStatus, statusUpdate CarrierStatus) {
	now := time.Now().UTC()
	// make sure system time is greater than or equal to the time used in the state file.
	forceTimeSynchronizationPersistence()

	// set curr
	for _, dev := range currentStatus.StatusDevices {
		currInfo := StatusInfo{
			Option:    getOrDefault(dev.Option),
			CreatedAt: now,
		}
		status.CurrentStatus.Devices[registry.CarrierDeviceName(dev.Device)] = currInfo
	}
	status.CurrentStatus.Status = currentStatus.Enable
	status.CurrentStatus.CreatedAt = now

	findOption := func(deviceName registry.CarrierDeviceName) string {
		for _, dev := range statusUpdate.StatusDevices {
			if dev.Device == string(deviceName) {
				return dev.Option
			}
		}
		return "none"
	}

	// set next
	for _, device := range carrier.Devices {
		status.NextStatus.Devices[device.Name] = StatusInfo{
			Option:    findOption(device.Name),
			CreatedAt: now,
		}

	}
	status.NextStatus.Status = statusUpdate.Enable
	status.NextStatus.CreatedAt = now
}

func getOrDefault(option string) string {
	return cmp.Or(option, string(registry.None))
}

// Compute the boot time by subtracting the uptime from the current time.
// This will be compared with the timestamp stored in the configuration file.
// An extra 5 seconds to account for the duration of the shutdown procedure.
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
	uptimeSeconds -= 5 // take into account shutdown time
	uptime := time.Duration(math.Round(uptimeSeconds)) * time.Second
	bootTime := time.Now().UTC().Add(-uptime)

	return bootTime, nil
}

func getStatusFile(cfg config.Configuration, carrierName registry.CarrierName) *paths.Path {
	return cfg.StatusDir().Join(string(carrierName) + ".json")
}

func loadStatusFile(statusFile *paths.Path) (*StatusFile, error) {
	data, err := statusFile.ReadFile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			newStatus := StatusFile{
				CurrentStatus: StatusCarrier{
					Devices: make(map[registry.CarrierDeviceName]StatusInfo),
				},
				NextStatus: StatusCarrier{
					Devices: make(map[registry.CarrierDeviceName]StatusInfo),
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

// forceTimeSynchronizationPersistence manually persists the current
// system time to disk. The systemd-timesyncd service normally updates
// this file during every synchronization, or every 60 seconds if no
// updates occur.
// See: https://www.man7.org/linux/man-pages/man5/timesyncd.conf.5.html
func forceTimeSynchronizationPersistence() {
	clockFile := "/var/lib/systemd/timesync/clock"
	if !paths.New(clockFile).Exist() {
		feedback.Warnf("Clock time synchronization service file %s not found", clockFile)
		return
	}
	cmd := exec.Command("touch", clockFile)
	if err := cmd.Run(); err != nil {
		feedback.Warnf("Error touch clock time synchronization service file %s", clockFile)
	}
}
