// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package status

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/registry"
)

type StatusFile struct {
	CurrentStatus StatusMount `json:"current_status"`
	NextStatus    StatusMount `json:"next_status"`
}

type StatusMount struct {
	Status    bool   `json:"status"`
	CreatedAt string `json:"created_at"`

	Devices map[registry.DeviceName]StatusInfo `json:"devices"`
}

type StatusInfo struct {
	Option    string `json:"option"`
	CreatedAt string `json:"created_at"`
}

type MountStatus struct {
	Enable        bool
	StatusDevices []StatusDevice
}

type StatusDevice struct {
	Device string
	Option string
}

// Called by config and reset
func Update(cfg config.Configuration, carrier registry.Mount, statusUpdate MountStatus) error {
	status, err := loadStatusFile(getStatusFile(cfg, carrier.Name))
	if err != nil {
		return fmt.Errorf("failed to load status file %w", err)
	}

	currentBootId, err := getCurrentBootID()
	if err != nil {
		return fmt.Errorf("failed to get boot-id: %w", err)
	}
	currentStatus, _ := getStatusStructure(status, carrier, currentBootId)
	updateStatusStructure(status, carrier, currentStatus, statusUpdate, currentBootId)

	if err := saveStatusFile(getStatusFile(cfg, carrier.Name), *status); err != nil {
		return fmt.Errorf("failed to save status file: %w", err)
	}
	return nil
}

// Called by show, load the status structure and apply status fixes before returning
// Do not update the status file on the disk because show is running as non-root user
func Get(cfg config.Configuration, carrier registry.Mount) (MountStatus, MountStatus, error) {
	status, err := loadStatusFile(getStatusFile(cfg, carrier.Name))
	if err != nil {
		return MountStatus{}, MountStatus{}, fmt.Errorf("failed to load status file %v", err)
	}

	currentBootId, err := getCurrentBootID()
	if err != nil {
		return MountStatus{}, MountStatus{}, fmt.Errorf("failed to get boot-id: %v", err)
	}

	current, next := getStatusStructure(status, carrier, currentBootId)
	return current, next, nil
}

func getStatusStructure(status *StatusFile, carrier registry.Mount, currentBootId string) (MountStatus, MountStatus) {
	current := MountStatus{
		Enable:        false,
		StatusDevices: make([]StatusDevice, 0, len(carrier.Devices)),
	}
	next := MountStatus{
		Enable:        false,
		StatusDevices: make([]StatusDevice, 0, len(carrier.Devices)),
	}

	for _, device := range carrier.Devices {
		// parse next structure, move in actual events happened before the boot
		if status.NextStatus.Devices[device.Name].CreatedAt != currentBootId {
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

	if status.NextStatus.CreatedAt != currentBootId {
		status.CurrentStatus.Status = status.NextStatus.Status
	}

	next.Enable = status.NextStatus.Status
	current.Enable = status.CurrentStatus.Status

	return current, next
}

func updateStatusStructure(status *StatusFile, carrier registry.Mount, currentStatus MountStatus, statusUpdate MountStatus, bootId string) {
	// set curr
	for _, dev := range currentStatus.StatusDevices {
		currInfo := StatusInfo{
			Option:    getOrDefault(dev.Option),
			CreatedAt: bootId,
		}
		status.CurrentStatus.Devices[registry.DeviceName(dev.Device)] = currInfo
	}
	status.CurrentStatus.Status = currentStatus.Enable
	status.CurrentStatus.CreatedAt = bootId

	findOption := func(deviceName registry.DeviceName) string {
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
			CreatedAt: bootId,
		}

	}
	status.NextStatus.Status = statusUpdate.Enable
	status.NextStatus.CreatedAt = bootId
}

func getOrDefault(option string) string {
	return cmp.Or(option, string(registry.None))
}

func getStatusFile(cfg config.Configuration, carrierName registry.MountName) *paths.Path {
	return cfg.StatusDir().Join(string(carrierName) + ".json")
}

func loadStatusFile(statusFile *paths.Path) (*StatusFile, error) {
	data, err := statusFile.ReadFile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			newStatus := StatusFile{
				CurrentStatus: StatusMount{
					Devices: make(map[registry.DeviceName]StatusInfo),
				},
				NextStatus: StatusMount{
					Devices: make(map[registry.DeviceName]StatusInfo),
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

func getCurrentBootID() (string, error) {
	bootID, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return "", err
	}

	cleanedBytes := bytes.TrimSpace(bootID)
	return string(cleanedBytes), nil
}
