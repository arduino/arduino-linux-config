package dto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildFdtoverlayCommand(t *testing.T) {
	tests := []struct {
		name        string
		overlays    []string
		wantCommand string
	}{
		{
			name:        "single overlay",
			overlays:    []string{"overlay1.dtbo"},
			wantCommand: "/usr/bin/fdtoverlay -i /boot/efi/dtb/qcom/qrb2210-arduino-imola-base.dtb -o /boot/efi/dtb/qcom/qrb2210-arduino-imola.dtb.next /boot/efi/dtb/qcom/overlay1.dtbo",
		},
		{
			name:        "multiple overlays",
			overlays:    []string{"overlay1.dtbo", "overlay2.dtbo", "overlay3.dtbo"},
			wantCommand: "/usr/bin/fdtoverlay -i /boot/efi/dtb/qcom/qrb2210-arduino-imola-base.dtb -o /boot/efi/dtb/qcom/qrb2210-arduino-imola.dtb.next /boot/efi/dtb/qcom/overlay1.dtbo /boot/efi/dtb/qcom/overlay2.dtbo /boot/efi/dtb/qcom/overlay3.dtbo",
		},
		{
			name:        "empty overlays",
			overlays:    []string{},
			wantCommand: "/usr/bin/fdtoverlay -i /boot/efi/dtb/qcom/qrb2210-arduino-imola-base.dtb -o /boot/efi/dtb/qcom/qrb2210-arduino-imola.dtb.next",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, _ := buildFdtoverlayCommand(tt.overlays)
			require.Equal(t, tt.wantCommand, strings.Join(args, " "))
		})
	}
}
