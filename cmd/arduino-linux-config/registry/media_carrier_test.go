package registry

import (
	"testing"
)

func TestIsOptionValid(t *testing.T) {
	tests := []struct {
		name       string
		deviceName MediaCarrierDeviceName
		option     string
		want       bool
	}{
		{name: "Valid device and valid option",
			deviceName: Camera1,
			option:     "type1-2lane",
			want:       true,
		},
		{name: "Valid device and invalid option",
			deviceName: Camera1,
			option:     "non-existent",
			want:       false,
		},
		{name: "Empty strings",
			deviceName: "",
			option:     "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsOptionValid(tt.deviceName, tt.option)
			if got != tt.want {
				t.Errorf("IsOptionValid(%v, %v) = %v; want %v",
					tt.deviceName, tt.option, got, tt.want)
			}
		})
	}
}
