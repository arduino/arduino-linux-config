// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package dto

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestBuildFdtoverlayCommand(t *testing.T) {
	fixedTime := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		overlays    []string
		wantCommand string
	}{
		{
			name:        "single overlay",
			overlays:    []string{"overlay1.dtbo"},
			wantCommand: "fdtoverlay -i /boot/efi/dtb/qcom/qrb2210-arduino-imola-base.dtb -o /boot/efi/dtb/qcom/temporaryDeviceTree.1777291200000000000.temp /boot/efi/dtb/qcom/overlay1.dtbo",
		},
		{
			name:        "multiple overlays",
			overlays:    []string{"overlay1.dtbo", "overlay2.dtbo", "overlay3.dtbo"},
			wantCommand: "fdtoverlay -i /boot/efi/dtb/qcom/qrb2210-arduino-imola-base.dtb -o /boot/efi/dtb/qcom/temporaryDeviceTree.1777291200000000000.temp /boot/efi/dtb/qcom/overlay1.dtbo /boot/efi/dtb/qcom/overlay2.dtbo /boot/efi/dtb/qcom/overlay3.dtbo",
		},
		{
			name:        "empty overlays",
			overlays:    []string{},
			wantCommand: "cp /boot/efi/dtb/qcom/qrb2210-arduino-imola-base.dtb /boot/efi/dtb/qcom/temporaryDeviceTree.1777291200000000000.temp",
		},
	}

	var overlaysDir = paths.New("/boot/efi/dtb/qcom/")
	var baseDtbFile = "qrb2210-arduino-imola-base.dtb"

	// Generate unique temp file name using nanosecond timestamp to prevent
	// race conditions when multiple instances run concurrently
	tempFileName := fmt.Sprintf("temporaryDeviceTree.%d.temp", fixedTime.UnixNano())
	temporaryDtb := overlaysDir.Join(tempFileName)
	fullPathBaseDtbFile := overlaysDir.Join(baseDtbFile)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := buildOverlayCommand(overlaysDir, fullPathBaseDtbFile.String(), temporaryDtb, tt.overlays)
			require.Equal(t, tt.wantCommand, strings.Join(args, " "))
		})
	}
}
