package dto

import (
	"context"
	"fmt"
	"slices"

	"github.com/arduino/go-paths-helper"
)

var fdtoverlayPath = paths.New("/usr/bin/fdtoverlay")

var qcomDBTDir = paths.New("/boot/efi/dtb/qcom/")
var systemDTB = qcomDBTDir.Join("qrb2210-arduino-imola.dtb")
var baseDTB = qcomDBTDir.Join("qrb2210-arduino-imola-base.dtb")

func Apply(overlays []string) error {
	if len(overlays) == 0 {
		return nil
	}

	slices.Sort(overlays)
	overlays = slices.Compact(overlays)

	args, tempFile := buildFdtoverlayCommand(overlays)

	cmd, err := paths.NewProcess(nil, args...)
	if err != nil {
		return fmt.Errorf("failed to create process: %w", err)
	}

	defer func() { _ = tempFile.Remove() }()

	_, stderr, err := cmd.RunAndCaptureOutput(context.Background())
	if err != nil {
		return fmt.Errorf("fdtoverlay failed with command %v: %w (stderr: %s)", args, err, stderr)
	}

	if err := tempFile.Rename(systemDTB); err != nil {
		return fmt.Errorf("failed to move %s to %s: %w", tempFile, systemDTB, err)
	}

	return nil
}

func buildFdtoverlayCommand(overlays []string) ([]string, *paths.Path) {
	temporaryDtb := qcomDBTDir.Join("qrb2210-arduino-imola.dtb.next")

	overlayPaths := make([]string, len(overlays))
	for i, overlay := range overlays {
		overlayPaths[i] = qcomDBTDir.Join(overlay).String()
	}

	args := append([]string{fdtoverlayPath.String(), "-i", baseDTB.String(), "-o", temporaryDtb.String()}, overlayPaths...)

	return args, temporaryDtb
}
