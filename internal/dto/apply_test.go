// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package dto

import (
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"

	"github.com/arduino/arduino-linux-config/internal/executor"
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

func TestApplyDryRunPrintsEveryEffect(t *testing.T) {
	recorder := executor.NewRecorder()

	combinedDtb := paths.New(t.TempDir(), "combined-dtb.dtb")
	require.NoError(t, combinedDtb.WriteFile(append(fakeDeviceTree("qcom,other"), fakeDeviceTree("arduino,monza")...)))

	board := VentunoQ{
		BaseDtbFullPath: combinedDtb.String(),
		OverlaysDir:     paths.New("/var/lib/overlays/"),
		DtbFileName:     "combined-dtb.dtb",
	}
	require.NoError(t, board.Apply(t.Context(), recorder, []string{"b.dtbo", "a.dtbo", "a.dtbo"}))

	temp := regexp.MustCompile(`temporaryDeviceTree\.\d+\.temp`)
	effects := recorder.Effects()
	for i, effect := range effects {
		effects[i] = temp.ReplaceAllString(effect, "TEMP")
	}

	require.Equal(t, []string{
		"mkdir -p /run/arduino-linux-config/dtb",
		"mount -t vfat /dev/disk/by-partlabel/dtb_a /run/arduino-linux-config/dtb",
		"# write /run/arduino-linux-config/dtb/monza.dtb (24 bytes)",
		"fdtoverlay -i /run/arduino-linux-config/dtb/monza.dtb -o /run/arduino-linux-config/dtb/TEMP /var/lib/overlays/a.dtbo /var/lib/overlays/b.dtbo",
		"# write /run/arduino-linux-config/dtb/TEMP (48 bytes)",
		"mv /run/arduino-linux-config/dtb/TEMP /run/arduino-linux-config/dtb/combined-dtb.dtb",
		"sync",
		"rm -f /run/arduino-linux-config/dtb/TEMP",
		"rm -f /run/arduino-linux-config/dtb/monza.dtb",
		"sync",
		"umount -l /run/arduino-linux-config/dtb",
	}, effects)
}

func TestPackUnchangedDeviceTreePreservesCombinedDtb(t *testing.T) {
	mountPoint := paths.New(t.TempDir())
	other := fakeDeviceTree("qcom,other")
	monza := fakeDeviceTreeWithSize("arduino,monza", 23)
	combined := append(append([]byte{}, other...), monza...)
	temporaryDtb := mountPoint.Join("temporary.dtb")
	require.NoError(t, temporaryDtb.WriteFile(monza))

	packedDtb, err := pack(executor.Real(), temporaryDtb, &unpackedDeviceTree{
		mountPoint:  mountPoint,
		monzaIndex:  1,
		deviceTrees: [][]byte{other, monza},
	})
	require.NoError(t, err)

	packed, err := os.ReadFile(packedDtb.String())
	require.NoError(t, err)
	require.Equal(t, combined, packed)
}

// A minimal flattened device tree: magic, total size and the compatible string.
func fakeDeviceTree(compatible string) []byte {
	return fakeDeviceTreeWithSize(compatible, 24)
}

func fakeDeviceTreeWithSize(compatible string, size int) []byte {
	payload := make([]byte, size)
	binary.BigEndian.PutUint32(payload, 0xd00dfeed)
	binary.BigEndian.PutUint32(payload[4:], uint32(len(payload))) //nolint:gosec
	copy(payload[8:], compatible)
	return payload
}
