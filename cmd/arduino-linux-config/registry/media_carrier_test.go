package registry

import (
	"testing"

	"github.com/stretchr/testify/require"
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

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name       string
		rawDevice  string
		rawOption  string
		wantDevice MediaCarrierDeviceName
		wantErr    bool
	}{
		{
			name:       "Valid device and valid option",
			rawDevice:  "camera1",
			rawOption:  "type1-2lane",
			wantDevice: Camera1,
			wantErr:    false,
		},
		{
			name:       "Unknown device",
			rawDevice:  "toaster",
			rawOption:  "burn",
			wantDevice: "",
			wantErr:    true,
		},
		{
			name:       "Known device but invalid option",
			rawDevice:  "camera1",
			rawOption:  "invalid-option",
			wantDevice: "",
			wantErr:    true,
		},
		{
			name:       "Known device none option",
			rawDevice:  "camera1",
			rawOption:  "none",
			wantDevice: Camera1,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDevice, err := ValidateInput(tt.rawDevice, tt.rawOption)
			if tt.wantErr {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}
			require.Equal(t, tt.wantDevice, gotDevice)
		})
	}
}
