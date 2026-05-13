// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package carrier

import (
	"reflect"
	"testing"

	"github.com/arduino/arduino-linux-config/internal/status"
)

func Test_parseArguments(t *testing.T) {
	tests := []struct {
		name        string
		carrierName string
		args        []string
		want        []status.StatusDevice
		wantErr     bool
	}{
		{
			name: "One single configuration",
			args: []string{"display=8-dsi-touch-a"},
			want: []status.StatusDevice{
				{Device: "display", Option: "8-dsi-touch-a"},
			},
			wantErr: false,
		},
		{
			name: "Two configuration in one string",
			args: []string{"display=8-dsi-touch-a,camera0=type1-2lanes"},
			want: []status.StatusDevice{
				{Device: "display", Option: "8-dsi-touch-a"},
				{Device: "camera0", Option: "type1-2lanes"},
			},
			wantErr: false,
		},
		{
			name: "Happy path: multiple arguments and spaces",
			args: []string{" display=8-dsi-touch-a ", " camera0 = type1-2lanes "},
			want: []status.StatusDevice{
				{Device: "display", Option: "8-dsi-touch-a"},
				{Device: "camera0", Option: "type1-2lanes"},
			},
			wantErr: false,
		},
		{
			name: "One option defined with a comma",
			args: []string{"camera0=type1-2lanes,"},
			want: []status.StatusDevice{
				{Device: "camera0", Option: "type1-2lanes"},
			},
			wantErr: false,
		},
		{
			name:    "Error: Values containing more than one equal sign, invalid format",
			args:    []string{"camera0=camera1=type1-2lanes"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Error: missing equals sign, invalid format",
			args:    []string{"camera0"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Error: duplicated device in arguments, invalid format",
			args:    []string{"camera0=type1-2lanes", "camera0=type1-4lanes"},
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
