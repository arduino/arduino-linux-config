package registry

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/go-paths-helper"
)

type StateFile struct {
	Carriers []CarrierState `json:"carriers"`
}

func (s *StateFile) FindOrCreateCarrierState(carrier Carrier) CarrierState {
	idx := slices.IndexFunc(s.Carriers, func(s CarrierState) bool {
		return s.Name == carrier.Name
	})
	if idx == -1 {
		newCarrier := CarrierState{
			Name: carrier.Name,
			CurrentStatus: StatusCarrier{
				Status:  false,
				Devices: make(map[CarrierDeviceName]StatusInfo, len(carrier.Devices)),
			},
			NextStatus: StatusCarrier{
				Status:  false,
				Devices: make(map[CarrierDeviceName]StatusInfo, len(carrier.Devices)),
			},
		}
		s.Carriers = append(s.Carriers, newCarrier)
		return newCarrier
	}
	return s.Carriers[idx]
}

type CarrierState struct {
	Name          CarrierName   `json:"name"`
	CurrentStatus StatusCarrier `json:"current_status"`
	NextStatus    StatusCarrier `json:"next_status"`
}

type StatusCarrier struct {
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"created_at"`

	Devices map[CarrierDeviceName]StatusInfo `json:"devices"`
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
func StatusUpdate(cfg config.Configuration, carrier Carrier, statusUpdate CarrierStatus) error {
	state, err := loadStateFile(cfg.StateFile())
	if err != nil {
		return fmt.Errorf("failed to load state file: %w", err)
	}

	// save current state if reboot occurred
	bootTime, err := getBootTime()
	if err != nil {
		return fmt.Errorf("failed to get boot time: %w", err)
	}
	currentStatus, _ := getStatusStructure(state, carrier, bootTime)

	newstate := updateStatusStructure(state, carrier, currentStatus, statusUpdate)

	if err := saveStatusFile(cfg.StateFile(), newstate); err != nil {
		return fmt.Errorf("failed to save status file: %w", err)
	}
	return nil
}

// Called by show, load the status structure and apply status fixes before returning
// Do not update the status file on the disk because show is running as non-root user
func GetStatus(cfg config.Configuration, carrier Carrier) (CarrierStatus, CarrierStatus, error) {
	state, err := loadStateFile(cfg.StateFile())
	if err != nil {
		return CarrierStatus{}, CarrierStatus{}, fmt.Errorf("failed to load status file %v", err)
	}

	bootTime, err := getBootTime()
	if err != nil {
		return CarrierStatus{}, CarrierStatus{}, fmt.Errorf("failed to get boot time: %v", err)
	}

	current, next := getStatusStructure(state, carrier, bootTime)
	return current, next, nil
}

func getStatusStructure(state StateFile, carrier Carrier, bootTime time.Time) (CarrierStatus, CarrierStatus) {
	current := CarrierStatus{
		Enable:        false,
		StatusDevices: make([]StatusDevice, 0, len(carrier.Devices)),
	}
	next := CarrierStatus{
		Enable:        false,
		StatusDevices: make([]StatusDevice, 0, len(carrier.Devices)),
	}

	carrierState := state.FindOrCreateCarrierState(carrier)

	for _, device := range carrier.Devices {
		// parse next structure, move in actual events happened before the boot
		if carrierState.NextStatus.Devices[device.Name].CreatedAt.Before(bootTime) {
			carrierState.CurrentStatus.Devices[device.Name] = carrierState.NextStatus.Devices[device.Name]
		}
		current.StatusDevices = append(current.StatusDevices, StatusDevice{
			Device: string(device.Name),
			Option: getOrDefault(carrierState.CurrentStatus.Devices[device.Name].Option),
		})
		next.StatusDevices = append(next.StatusDevices, StatusDevice{
			Device: string(device.Name),
			Option: getOrDefault(carrierState.NextStatus.Devices[device.Name].Option),
		})
	}

	if carrierState.NextStatus.CreatedAt.Before(bootTime) {
		carrierState.CurrentStatus.Status = carrierState.NextStatus.Status
	}

	next.Enable = carrierState.NextStatus.Status
	current.Enable = carrierState.CurrentStatus.Status

	return current, next
}

func updateStatusStructure(state StateFile, carrier Carrier, currentStatus CarrierStatus, statusUpdate CarrierStatus) StateFile {
	now := time.Now().UTC()

	carrierState := state.FindOrCreateCarrierState(carrier)

	// set curr
	for _, dev := range currentStatus.StatusDevices {
		currInfo := StatusInfo{
			Option:    getOrDefault(dev.Option),
			CreatedAt: now,
		}
		carrierState.CurrentStatus.Devices[CarrierDeviceName(dev.Device)] = currInfo
	}
	carrierState.CurrentStatus.Status = currentStatus.Enable
	carrierState.CurrentStatus.CreatedAt = now

	findOption := func(deviceName CarrierDeviceName) string {
		for _, dev := range statusUpdate.StatusDevices {
			if dev.Device == string(deviceName) {
				return dev.Option
			}
		}
		return "none"
	}

	// set next
	for _, device := range carrier.Devices {
		carrierState.NextStatus.Devices[device.Name] = StatusInfo{
			Option:    findOption(device.Name),
			CreatedAt: now,
		}

	}
	carrierState.NextStatus.Status = statusUpdate.Enable
	carrierState.NextStatus.CreatedAt = now
	return state
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

func loadStateFile(stateFile *paths.Path) (StateFile, error) {
	data, err := stateFile.ReadFile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return StateFile{}, nil
		} else {
			return StateFile{}, err
		}
	}

	var status StateFile
	err = json.Unmarshal(data, &status)
	if err != nil {
		return StateFile{}, fmt.Errorf("could not parse json: %w", err)
	}

	return status, nil
}

func saveStatusFile(statusFile *paths.Path, state StateFile) error {
	data, err := json.MarshalIndent(state, "", "    ")
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
