// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package status

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"

	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/testutil"
)

func TestUpdateStatusStructure(t *testing.T) {
	cleanup := testutil.SetupCompatUnoq()
	defer cleanup()
	cleanupOS := testutil.SetupDebian()
	defer cleanupOS()

	tests := []struct {
		name         string
		statusUpdate CarrierStatus
		wantResult   map[registry.CarrierDeviceName]string
	}{
		{
			name: "update all devices",
			statusUpdate: CarrierStatus{
				Enable: true,
				StatusDevices: []StatusDevice{
					{Device: string(registry.Camera0), Option: "cam1"},
					{Device: string(registry.Camera1), Option: "none"},
					{Device: string(registry.Display), Option: "display"},
				},
			},
			wantResult: map[registry.CarrierDeviceName]string{
				registry.Camera0: "cam1",
				registry.Camera1: "none",
				registry.Display: "display",
			},
		},
		{
			name: "fill empty devices",
			statusUpdate: CarrierStatus{
				Enable: true,
				StatusDevices: []StatusDevice{
					{Device: string(registry.Display), Option: "display"},
				},
			},
			wantResult: map[registry.CarrierDeviceName]string{
				registry.Camera0: "none",
				registry.Camera1: "none",
				registry.Display: "display",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := &StatusFile{
				NextStatus: StatusCarrier{
					Devices: make(map[registry.CarrierDeviceName]StatusInfo),
				},
			}

			currentBootId := "001"
			reg, err := registry.New()
			require.NoError(t, err)
			mediaCarrier, exist := reg.FindByName("media-carrier")
			if !exist {
				t.Fatal("media-carrier not found in registry")
			}
			updateStatusStructure(status, mediaCarrier, CarrierStatus{}, tt.statusUpdate, currentBootId)
			carrierDeviceLenght := len(mediaCarrier.Devices)
			if len(status.NextStatus.Devices) != carrierDeviceLenght {

				if len(status.CurrentStatus.Devices) != 0 {
					t.Errorf("got %d devices, want %d", len(status.CurrentStatus.Devices), carrierDeviceLenght)
				}

				for _, device := range mediaCarrier.Devices {
					info, exists := status.NextStatus.Devices[device.Name]
					if !exists {
						t.Fatalf("device %s missing from wanted status", device.Name)
					}
					if info.Option != tt.wantResult[device.Name] {
						t.Errorf("device %s: got option %s, want %s", device.Name, info.Option, tt.wantResult[device.Name])
					}

					if info.CreatedAt != currentBootId {
						t.Errorf("device %s: timestamp %v is too old", device.Name, info.CreatedAt)
					}
				}
			}
		})
	}
}

func TestGetStatusStructure(t *testing.T) {
	cleanup := testutil.SetupCompatUnoq()
	defer cleanup()
	cleanupOS := testutil.SetupDebian()
	defer cleanupOS()

	// set boot ids
	currentBootId := "123"
	prevCurrentBootId := "001"
	afterConfigurationBootId := currentBootId

	tests := []struct {
		name            string
		initialStatus   *StatusFile
		expectedCurrent []StatusDevice
		expectedNext    []StatusDevice
	}{
		{
			name: "Move outdated device to current, retain current status on show command",
			initialStatus: &StatusFile{
				CurrentStatus: StatusCarrier{
					Devices: make(map[registry.CarrierDeviceName]StatusInfo),
				},
				NextStatus: StatusCarrier{
					Devices: map[registry.CarrierDeviceName]StatusInfo{
						registry.Camera0: {Option: "cam1", CreatedAt: prevCurrentBootId},
					},
				},
			},
			expectedCurrent: []StatusDevice{
				{Device: string(registry.Camera0), Option: "cam1"},
				{Device: string(registry.Camera1), Option: "none"},
				{Device: string(registry.Display), Option: "none"},
			},
			expectedNext: []StatusDevice{
				{Device: string(registry.Camera0), Option: "cam1"},
				{Device: string(registry.Camera1), Option: "none"},
				{Device: string(registry.Display), Option: "none"},
			},
		},
		{
			name: "Both devices after boot stay in next",
			initialStatus: &StatusFile{
				CurrentStatus: StatusCarrier{
					Devices: make(map[registry.CarrierDeviceName]StatusInfo),
				},
				NextStatus: StatusCarrier{
					Devices: map[registry.CarrierDeviceName]StatusInfo{
						registry.Camera0: {Option: "cam1", CreatedAt: afterConfigurationBootId},    // fresh
						registry.Display: {Option: "display", CreatedAt: afterConfigurationBootId}, // fresh
					},
				},
			},
			expectedCurrent: []StatusDevice{
				{Device: string(registry.Camera0), Option: "none"},
				{Device: string(registry.Camera1), Option: "none"},
				{Device: string(registry.Display), Option: "none"},
			},
			expectedNext: []StatusDevice{
				{Device: string(registry.Camera0), Option: "cam1"},
				{Device: string(registry.Camera1), Option: "none"},
				{Device: string(registry.Display), Option: "display"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg, err := registry.New()
			require.NoError(t, err)
			mediaCarrier, exist := reg.FindByName("media-carrier")
			if !exist {
				t.Fatal("media-carrier not found in registry")
			}
			current, next := getStatusStructure(tt.initialStatus, mediaCarrier, currentBootId)

			// Validate Current Slice
			for i, want := range tt.expectedCurrent {
				if current.StatusDevices[i].Device != want.Device || current.StatusDevices[i].Option != want.Option {
					t.Errorf("%s [Current]: got %+v, want %+v", tt.name, current.StatusDevices[i], want)
				}
			}

			// Validate Next Slice
			for i, want := range tt.expectedNext {
				if next.StatusDevices[i].Device != want.Device || next.StatusDevices[i].Option != want.Option {
					t.Errorf("%s [Next]: got %+v, want %+v", tt.name, next.StatusDevices[i], want)
				}
			}
		})
	}
}

func TestLoadStatusFile(t *testing.T) {
	tests := []struct {
		name          string
		fileContent   string // Empty string means file doesn't exist
		shouldExist   bool
		wantErr       bool
		checkContents bool
	}{
		{
			name:        "File does not exist - returns initialized struct",
			shouldExist: false,
			wantErr:     false,
		},
		{
			name:          "Valid JSON file - returns parsed struct",
			shouldExist:   true,
			fileContent:   `{"current_status": {"devices": {"Cam0": {"option": "ON"}}}, "next_status": {"devices": {}}}`,
			wantErr:       false,
			checkContents: true,
		},
		{
			name:        "Invalid JSON - returns error",
			shouldExist: true,
			fileContent: `{"current_status": { invalid ]}`,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		carrierName := "media-carrier"
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, carrierName+".json")
			p := paths.New(filePath)

			if tt.shouldExist {
				if err := os.WriteFile(filePath, []byte(tt.fileContent), 0600); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
			}

			status, err := loadStatusFile(p)

			// 1. Check Error expectation
			if (err != nil) != tt.wantErr {
				t.Fatalf("loadStatusFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			// 2. Verify Initialization (The "os.ErrNotExist" case)
			if !tt.shouldExist {
				if status.CurrentStatus.Devices == nil || status.NextStatus.Devices == nil {
					t.Error("Maps should be initialized when file is missing")
				}
			}

			// 3. Verify Data Parsing
			if tt.checkContents {
				if _, ok := status.CurrentStatus.Devices["Cam0"]; !ok {
					t.Error("Expected device 'Cam0' to be parsed from JSON")
				}
			}
		})
	}
}
