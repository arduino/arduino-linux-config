package carrier

import (
	"testing"
)

func TestCarriersResult_String(t *testing.T) {
	tests := []struct {
		name     string
		input    carriersResult
		expected string
	}{
		{
			name: "Single carrier with multiple devices",
			input: carriersResult{
				Carriers: []Carrier{
					{
						Name: "media-carrier",
						Devices: []Device{
							{
								Name: "camera1",
								Options: []DeviceOption{
									{Name: "none", DtboFile: ""},
									{Name: "type1-2lane", DtboFile: "camera-imx219-csi0-2lanes.dtbo"},
									{Name: "type1-4lane", DtboFile: "camera-imx219-csi0-4lanes.dtbo"},
								},
							},
							{
								Name: "display1",
								Options: []DeviceOption{
									{Name: "none", DtboFile: ""},
									{Name: "8-dsi-touch-a", DtboFile: "panel-8in-touch-a-dsi.dtbo"},
								},
							},
						},
					},
				},
			},
			expected: "media-carrier:\n" +
				"  - camera1: none | type1-2lane | type1-4lane\n" +
				"  - display1: none | 8-dsi-touch-a\n",
		},
		{
			name: "Multiple carriers",
			input: carriersResult{
				Carriers: []Carrier{
					{
						Name: "media-carrier",
						Devices: []Device{
							{
								Name: "camera1",
								Options: []DeviceOption{
									{Name: "none", DtboFile: ""},
									{Name: "type1-2lane", DtboFile: "camera-imx219-csi0-2lanes.dtbo"},
								},
							},
						},
					},
					{
						Name: "builtin",
						Devices: []Device{
							{
								Name: "camera1",
								Options: []DeviceOption{
									{Name: "none", DtboFile: ""},
									{Name: "type1-4lane", DtboFile: "camera-imx219-csi0-4lanes.dtbo"},
								},
							},
						},
					},
				},
			},
			expected: "media-carrier:\n" +
				"  - camera1: none | type1-2lane\n" +
				"builtin:\n" +
				"  - camera1: none | type1-4lane\n",
		},
		{
			name: "Carrier with no devices",
			input: carriersResult{
				Carriers: []Carrier{
					{
						Name:    "empty-carrier",
						Devices: []Device{},
					},
				},
			},
			expected: "empty-carrier:\n",
		},
		{
			name:     "Empty registry",
			input:    carriersResult{Carriers: []Carrier{}},
			expected: "",
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

func TestGetCarrierRegistry(t *testing.T) {
	carriers := getCarrierRegistry()

	if len(carriers) == 0 {
		t.Fatal("registry is empty, expected at least one carrier")
	}

	for _, carrier := range carriers {
		t.Run(carrier.Name, func(t *testing.T) {
			if carrier.Name == "" {
				t.Error("carrier has empty name")
			}

			for _, device := range carrier.Devices {
				if device.Name == "" {
					t.Errorf("device in carrier %q has empty name", carrier.Name)
				}

				if len(device.Options) == 0 {
					t.Errorf("device %q in carrier %q has no options", device.Name, carrier.Name)
				}

				// First option must always be "none"
				if device.Options[0].Name != "none" {
					t.Errorf("device %q in carrier %q: first option must be 'none', got %q",
						device.Name, carrier.Name, device.Options[0].Name)
				}

				// "none" must have no dtbo file
				if device.Options[0].DtboFile != "" {
					t.Errorf("device %q in carrier %q: 'none' option must have empty DtboFile, got %q",
						device.Name, carrier.Name, device.Options[0].DtboFile)
				}

				// All other options must have a dtbo file
				for _, opt := range device.Options[1:] {
					if opt.DtboFile == "" {
						t.Errorf("device %q in carrier %q: option %q has empty DtboFile",
							device.Name, carrier.Name, opt.Name)
					}
				}
			}
		})
	}
}
