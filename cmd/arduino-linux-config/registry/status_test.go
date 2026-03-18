package registry

import (
	"testing"
	"time"
)

func TestUpdateStatusStructure(t *testing.T) {
	tests := []struct {
		name         string
		statusUpdate map[MediaCarrierDeviceName]string
		currResult   map[MediaCarrierDeviceName]string
		wantResult   map[MediaCarrierDeviceName]string
	}{
		{
			name: "update all devices",
			statusUpdate: map[MediaCarrierDeviceName]string{
				Camera1: "cam1",
				Display: "display",
			},
			currResult: map[MediaCarrierDeviceName]string{
				Camera1: "none",
				Camera2: "none",
				Display: "none",
			},
			wantResult: map[MediaCarrierDeviceName]string{
				Camera1: "cam1",
				Camera2: "none",
				Display: "display",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize StatusFile with an empty map to prevent nil panic
			status := &StatusFile{
				WantedStatus: StatusCarrier{
					Devices: make(map[MediaCarrierDeviceName]StatusInfo),
				},
			}

			// Capture time before execution to verify CreatedAt
			startTime := time.Now().UTC().Truncate(time.Second)

			updateStatusStructure(status, tt.statusUpdate)

			if len(status.WantedStatus.Devices) != len(MediaCarrierDeviceList) {
				t.Errorf("got %d devices, want %d", len(status.WantedStatus.Devices), len(MediaCarrierDeviceList))
			}

			// The input structure is not meant to be empty
			if len(status.CurrentStatus.Devices) != 0 {
				t.Errorf("got %d devices, want %d", len(status.CurrentStatus.Devices), len(MediaCarrierDeviceList))
			}

			for _, name := range MediaCarrierDeviceList {
				info, exists := status.WantedStatus.Devices[name]
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
