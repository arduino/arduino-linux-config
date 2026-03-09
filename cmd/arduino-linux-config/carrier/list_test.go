package carrier

import (
	"testing"

	carrierinfo "github.com/arduino/arduino-linux-config/internal/carrierinfo"
)

func TestExtractCarrierResult(t *testing.T) {
	tests := []struct {
		name     string
		input    carrierinfo.Carrier
		expected CarrierResult
	}{
		{
			name: "Successfully groups and prepends none",
			input: carrierinfo.Carrier{
				Name: "arduino-carrier",
				Overlays: []carrierinfo.Overlay{
					{DeviceName: "camera1", HardwareData: "imx219-2lane"},
					{DeviceName: "camera1", HardwareData: "imx219-4lane"},
					{DeviceName: "display1", HardwareData: "8in-dsi"},
				},
			},
			expected: CarrierResult{
				Name: "arduino-carrier",
				Devices: []Device{
					{
						Name:             "leds",
						AvailableDevices: []string{},
					},
					{
						Name:             "camera1",
						AvailableDevices: []string{"none", "imx219-2lane", "imx219-4lane"},
					},
					{
						Name:             "camera2",
						AvailableDevices: []string{},
					},
					{
						Name:             "display1",
						AvailableDevices: []string{"none", "8in-dsi"},
					},
				},
			},
		},
		{
			name: "Empty input returns empty result",
			input: carrierinfo.Carrier{
				Name:     "empty",
				Overlays: []carrierinfo.Overlay{},
			},
			expected: CarrierResult{
				Name: "empty",
				Devices: []Device{
					{
						Name:             "leds",
						AvailableDevices: []string{},
					},
					{
						Name:             "camera1",
						AvailableDevices: []string{},
					},
					{
						Name:             "camera2",
						AvailableDevices: []string{},
					},
					{
						Name:             "display1",
						AvailableDevices: []string{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCarrierResult(tt.input)

			// Validate Carrier Name
			if got.Name != tt.expected.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.expected.Name)
			}

			// Validate Device grouping and "none" insertion
			if !compareDevices(got.Devices, tt.expected.Devices) {
				t.Errorf("Devices = %v, want %v", got.Devices, tt.expected.Devices)
			}
		})
	}
}

func TestCarriersResult_String(t *testing.T) {
	tests := []struct {
		name     string
		input    carriersResult
		expected string
	}{
		{
			name: "Standard carrier with multiple devices",
			input: carriersResult{
				MediaCarrier: CarrierResult{
					Name: "arduino-max-carrier",
					Devices: []Device{
						{
							Name:             "camera1",
							AvailableDevices: []string{"none", "imx219-2lane", "imx219-4lane"},
						},
						{
							Name:             "display1",
							AvailableDevices: []string{"none", "8in-touch-a-dsi"},
						},
					},
				},
			},
			// Note: The spaces below must match the tabwriter's calculation (8 tab-width, 2 padding)
			expected: "- arduino-max-carrier\n" +
				"  camera1   none, imx219-2lane, imx219-4lane\n" +
				"  display1  none, 8in-touch-a-dsi\n",
		},
		{
			name: "Empty devices list",
			input: carriersResult{
				MediaCarrier: CarrierResult{
					Name:    "empty-carrier",
					Devices: nil,
				},
			},
			expected: "- empty-carrier\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.String()

			if got != tt.expected {
				t.Errorf("String() mismatch.\nGot:\n%q\nWant:\n%q", got, tt.expected)
			}
		})
	}
}

func compareDevices(got, want []Device) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Name != want[i].Name {
			return false
		}
		// Check the options slice length
		if len(got[i].AvailableDevices) != len(want[i].AvailableDevices) {
			return false
		}
		// Check every string in the options slice
		for j := range got[i].AvailableDevices {
			if got[i].AvailableDevices[j] != want[i].AvailableDevices[j] {
				return false
			}
		}
	}
	return true
}
