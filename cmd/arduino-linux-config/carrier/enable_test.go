package carrier

import (
	"reflect"
	"testing"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/registry"
)

func Test_parseArguments(t *testing.T) {
	tests := []struct {
		name        string
		carrierName string
		args        []string
		want        map[registry.CarrierDeviceName]string
		wantErr     bool
	}{
		{
			name: "One single configuration",
			args: []string{"display=8-dsi-touch-a"},
			want: map[registry.CarrierDeviceName]string{
				"display": "8-dsi-touch-a",
			},
			wantErr: false,
		},
		{
			name: "Two configuration in one string",
			args: []string{"display=8-dsi-touch-a,camera0=type1-2lane"},
			want: map[registry.CarrierDeviceName]string{
				"display": "8-dsi-touch-a",
				"camera0": "type1-2lane",
			},
			wantErr: false,
		},
		{
			name: "Happy path: multiple arguments and spaces",
			args: []string{" display=8-dsi-touch-a ", " camera0 = type1-2lane "},
			want: map[registry.CarrierDeviceName]string{
				"display": "8-dsi-touch-a",
				"camera0": "type1-2lane",
			},
			wantErr: false,
		},
		{
			name: "One option defined with a comma",
			args: []string{"camera0=type1-2lane,"},
			want: map[registry.CarrierDeviceName]string{
				"camera0": "type1-2lane",
			},
			wantErr: false,
		},
		{
			name:    "Error: Values containing more than one equal sign, invalid format",
			args:    []string{"camera0=camera1=type1-2lane"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Error: missing equals sign, invalid format",
			args:    []string{"camera0"},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUserArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseArguments() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseArguments() got = %v, want %v", got, tt.want)
			}
		})
	}
}
