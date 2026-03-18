package carrier

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/registry"
	"github.com/arduino/arduino-linux-config/cmd/config"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
)

type CarrierStatus struct {
	Configuration map[registry.MediaCarrierDeviceName]string `json:"configuration,omitempty"`
}

func newEnableCmd(cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "enable <carrier-name> [device=option...]",
		Short: "Enable a carrier with the specified device options",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			enableHandler(cmd.Context(), cfg, args[0], args[1:])
		},
	}
}

// Since a board reboot can occur asynchronously with the carrier configuration,
// we must track both the current and desired statuses.
//
// This is managed by creating a status file every time a device is congfigured.
// The status file records the device name, the configured options, and a timestamp.
//
// At show time we check the last boot time.
// Files with a timestamp before boot represent the current configuration.
// Files with a timestamp after boot represent the pending (next) configuration.
func enableHandler(ctx context.Context, cfg config.Configuration, carrierName string, deviceArgs []string) {

	wantedDevicesList, err := parseArguments(carrierName, deviceArgs)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	overlayList := collectDtboFiles(cfg, wantedDevicesList)

	disableHandler(cfg, ctx, carrierName)
	if err := applyOverlays(cfg, ctx, overlayList); err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}
	// TODO Add return feedback on configuration done and the configured values

	registry.StatusUpdate(cfg, wantedDevicesList)
}

func parseArguments(carrierName string, args []string) (map[registry.MediaCarrierDeviceName]string, error) {
	if carrierName != registry.MediaCarrierRegistry.Name {
		return nil, fmt.Errorf("carrier %q not supported", carrierName)
	}

	selection := make(map[registry.MediaCarrierDeviceName]string)
	for _, arg := range args {
		arg = strings.TrimRight(arg, ",")

		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid argument %q: expected device=option format", arg)
		}

		deviceName := parts[0]
		optionName := parts[1]

		mediaDeviceName, err := registry.ValidateInput(deviceName, optionName)
		if err != nil {
			return nil, err
		}

		selection[mediaDeviceName] = optionName
	}

	return selection, nil
}

func collectDtboFiles(cfg config.Configuration, selection map[registry.MediaCarrierDeviceName]string) []string {
	var dtboFiles []string

	for deviceName, optionName := range selection {
		if optionName == registry.DeviceOptionNone || optionName == "" {
			continue
		}

		for _, device := range registry.MediaCarrierRegistry.Devices {
			if device.Name != deviceName {
				continue
			}
			for _, opt := range device.Options {
				if opt.Name == optionName {
					if opt.DtboFile != "" {
						dtboFiles = append(dtboFiles, filepath.Join(cfg.OverlaysDir().String(), opt.DtboFile))
					}
					break
				}
			}
		}
	}

	// if at least one device is configured we activate the overlay for the media carrier
	if len(dtboFiles) > 0 {
		dtboFiles = append(dtboFiles, filepath.Join(cfg.OverlaysDir().String(), "qrb2210-arduino-imola-carrier-media.dtbo"))
	}

	return dtboFiles
}

func applyOverlays(cfg config.Configuration, ctx context.Context, dtboFiles []string) error {
	if len(dtboFiles) <= 0 {
		return nil
	}

	args := append([]string{"fdtoverlay", "-i", cfg.FactoryDTB().String(), "-o", cfg.ActualDTB().String()}, dtboFiles...)

	proc, err := paths.NewProcess(nil, args...)
	if err != nil {
		return fmt.Errorf("failed to create process: %w", err)
	}
	stdout, stderr, err := proc.RunAndCaptureOutput(ctx)
	if err != nil {
		return fmt.Errorf("fdtoverlay failed: %w\n%s", err, stderr)
	}
	if len(stdout) > 0 {
		feedback.Print(string(stdout))
	}
	return nil
}
