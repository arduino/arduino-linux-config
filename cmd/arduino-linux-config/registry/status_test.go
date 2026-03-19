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
		statusUpdate map[MediaCarrierDeviceName]string
		wantResult   map[MediaCarrierDeviceName]string
	}{
		{
			name: "update all devices",
			statusUpdate: map[MediaCarrierDeviceName]string{
				Camera1: "cam1",
				Camera2: "none",
				Display: "display",
			},
			wantResult: map[MediaCarrierDeviceName]string{
				Camera1: "cam1",
				Camera2: "none",
				Display: "display",
			},
		},
		{
			name: "fill empty devices",
			statusUpdate: map[MediaCarrierDeviceName]string{
				Display: "display",
			},
			wantResult: map[MediaCarrierDeviceName]string{
				Camera1: "none",
				Camera2: "none",
				Display: "display",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := &StatusFile{
				NextStatus: StatusCarrier{
					Devices: make(map[MediaCarrierDeviceName]StatusInfo),
				},
			}

			startTime := time.Now().UTC().Truncate(time.Second)

			updateStatusStructure(status, tt.statusUpdate)

			if len(status.NextStatus.Devices) != len(MediaCarrierDeviceList) {
				t.Errorf("got %d devices, want %d", len(status.NextStatus.Devices), len(MediaCarrierDeviceList))
			}

			if len(status.CurrentStatus.Devices) != 0 {
				t.Errorf("got %d devices, want %d", len(status.CurrentStatus.Devices), len(MediaCarrierDeviceList))
			}

			for _, name := range MediaCarrierDeviceList {
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
				CurrentStatus: StatusCarrier{Devices: make(map[MediaCarrierDeviceName]StatusInfo)},
				NextStatus: StatusCarrier{
					Devices: map[MediaCarrierDeviceName]StatusInfo{
						Camera1: {Option: "cam1", CreatedAt: beforeBoot},
						Display: {Option: "display", CreatedAt: afterBoot},
					},
				},
			},
			expectedCurrent: []StatusDevice{
				{Device: string(Camera1), Option: "cam1"},
				{Device: string(Camera2), Option: "none"},
				{Device: string(Display), Option: "none"},
			},
			expectedNext: []StatusDevice{
				{Device: string(Camera1), Option: "none"},
				{Device: string(Camera2), Option: "none"},
				{Device: string(Display), Option: "display"},
			},
		},
		{
			name: "Both devices after boot stay in next",
			initialStatus: &StatusFile{
				CurrentStatus: StatusCarrier{Devices: make(map[MediaCarrierDeviceName]StatusInfo)},
				NextStatus: StatusCarrier{
					Devices: map[MediaCarrierDeviceName]StatusInfo{
						Camera1: {Option: "cam1", CreatedAt: afterBoot},    // fresh
						Display: {Option: "display", CreatedAt: afterBoot}, // fresh
					},
				},
			},
			expectedCurrent: []StatusDevice{
				{Device: string(Camera1), Option: "none"},
				{Device: string(Camera2), Option: "none"},
				{Device: string(Display), Option: "none"},
			},
			expectedNext: []StatusDevice{
				{Device: string(Camera1), Option: "cam1"},
				{Device: string(Camera2), Option: "none"},
				{Device: string(Display), Option: "display"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, next := getStatusStructure(tt.initialStatus, bootTime)

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
			fileContent:   `{"CurrentStatus": {"Devices": {"Cam1": {"Option": "ON"}}}, "WantedStatus": {"Devices": {}}}`,
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
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "status.json")
			p := paths.New(filePath)

			if tt.shouldExist {
				os.WriteFile(filePath, []byte(tt.fileContent), 0644)
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
				if _, ok := status.CurrentStatus.Devices["Cam1"]; !ok {
					t.Error("Expected device 'Cam1' to be parsed from JSON")
				}
			}
		})
	}
}
