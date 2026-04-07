package registry

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arduino/go-paths-helper"
)

func TestUpdateStatusStructure(t *testing.T) {
	tests := []struct {
		name         string
		statusUpdate map[CarrierDeviceName]string
		wantResult   map[CarrierDeviceName]string
	}{
		{
			name: "update all devices",
			statusUpdate: map[CarrierDeviceName]string{
				Camera0: "cam1",
				Camera1: "none",
				Display: "display",
			},
			wantResult: map[CarrierDeviceName]string{
				Camera0: "cam1",
				Camera1: "none",
				Display: "display",
			},
		},
		{
			name: "fill empty devices",
			statusUpdate: map[CarrierDeviceName]string{
				Display: "display",
			},
			wantResult: map[CarrierDeviceName]string{
				Camera0: "none",
				Camera1: "none",
				Display: "display",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := &StatusFile{
				NextStatus: StatusCarrier{
					Devices: make(map[CarrierDeviceName]StatusInfo),
				},
			}

			startTime := time.Now().UTC().Truncate(time.Second)

			updateStatusStructure(status, "media-carrier", []StatusDevice{}, tt.statusUpdate)
			carrierDeviceLenght := len(GetDevicesNames("media-carrier"))
			if len(status.NextStatus.Devices) != carrierDeviceLenght {

				if len(status.CurrentStatus.Devices) != 0 {
					t.Errorf("got %d devices, want %d", len(status.CurrentStatus.Devices), carrierDeviceLenght)
				}

				for _, name := range GetDevicesNames("media-carrier") {
					info, exists := status.NextStatus.Devices[name]
					if !exists {
						t.Fatalf("device %s missing from wanted status", name)
					}
					if info.Option != tt.wantResult[name] {
						t.Errorf("device %s: got option %s, want %s", name, info.Option, tt.wantResult[name])
					}

					if info.CreatedAt.Before(startTime) {
						t.Errorf("device %s: timestamp %v is too old", name, info.CreatedAt)
					}
				}
			}
		})
	}
}

func TestGetStatusStructure(t *testing.T) {
	// set boot times
	bootTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	beforeBoot := bootTime.Add(-1 * time.Hour)
	afterBoot := bootTime.Add(1 * time.Hour)

	tests := []struct {
		name            string
		initialStatus   *StatusFile
		expectedCurrent []StatusDevice
		expectedNext    []StatusDevice
	}{
		{
			name: "Move outdated device to current",
			initialStatus: &StatusFile{
				CurrentStatus: StatusCarrier{Devices: make(map[CarrierDeviceName]StatusInfo)},
				NextStatus: StatusCarrier{
					Devices: map[CarrierDeviceName]StatusInfo{
						Camera0: {Option: "cam1", CreatedAt: beforeBoot},
						Display: {Option: "display", CreatedAt: afterBoot},
					},
				},
			},
			expectedCurrent: []StatusDevice{
				{Device: string(Camera0), Option: "cam1"},
				{Device: string(Camera1), Option: "none"},
				{Device: string(Display), Option: "none"},
			},
			expectedNext: []StatusDevice{
				{Device: string(Camera0), Option: "none"},
				{Device: string(Camera1), Option: "none"},
				{Device: string(Display), Option: "display"},
			},
		},
		{
			name: "Both devices after boot stay in next",
			initialStatus: &StatusFile{
				CurrentStatus: StatusCarrier{Devices: make(map[CarrierDeviceName]StatusInfo)},
				NextStatus: StatusCarrier{
					Devices: map[CarrierDeviceName]StatusInfo{
						Camera0: {Option: "cam1", CreatedAt: afterBoot},    // fresh
						Display: {Option: "display", CreatedAt: afterBoot}, // fresh
					},
				},
			},
			expectedCurrent: []StatusDevice{
				{Device: string(Camera0), Option: "none"},
				{Device: string(Camera1), Option: "none"},
				{Device: string(Display), Option: "none"},
			},
			expectedNext: []StatusDevice{
				{Device: string(Camera0), Option: "cam1"},
				{Device: string(Camera1), Option: "none"},
				{Device: string(Display), Option: "display"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, next := getFixedStatusStructure(tt.initialStatus, "media-carrier", bootTime)

			// Validate Current Slice
			for i, want := range tt.expectedCurrent {
				if current[i].Device != want.Device || current[i].Option != want.Option {
					t.Errorf("%s [Current]: got %+v, want %+v", tt.name, current[i], want)
				}
			}

			// Validate Next Slice
			for i, want := range tt.expectedNext {
				if next[i].Device != want.Device || next[i].Option != want.Option {
					t.Errorf("%s [Next]: got %+v, want %+v", tt.name, next[i], want)
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
			fileContent:   `{"CurrentStatus": {"Devices": {"Cam0": {"Option": "ON"}}}, "WantedStatus": {"Devices": {}}}`,
			wantErr:       false,
			checkContents: true,
		},
		{
			name:        "Invalid JSON - returns error",
			shouldExist: true,
			fileContent: `{"CurrentStatus": { invalid ]}`,
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
