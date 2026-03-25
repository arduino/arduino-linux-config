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

func newConfigureCmd(cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "configure <carrier-name> <device=option...>",
		Short: "Configure a carrier with the specified device options",

		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("missing <carrier-name>\nUsage: arduino-linux-config carrier configure <carrier-name> <device=option...>")
			}
			if len(args) < 1 {
				return fmt.Errorf("missing carrier configuration\nUsage: arduino-linux-config carrier configure <name>")
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			configureHandler(cmd.Context(), cfg, args[0], args[1:])
		},
	}
}

// Since a board reboot can occur asynchronously with the carrier configuration,
// we must track both the current and next status.
//
// State management is handled via a status file updated on device configurations.
// This file stores the device name, configuration options, and a metadata timestamp.
//
// Whenever a status request occurs, the system compares the last boot time against the configuration
// timestamp to update the current and next states.
func configureHandler(ctx context.Context, cfg config.Configuration, carrierName string, deviceArgs []string) {

	wantedDevicesList, err := parseArguments(carrierName, deviceArgs)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	overlayList := collectDtboFiles(cfg, wantedDevicesList)

	Reset(cfg)

	if err := applyOverlays(cfg, ctx, overlayList); err != nil {
		feedback.Fatal(
			fmt.Sprintf("failed to apply overlays (carrier has been reset): %v", err),
			feedback.ErrGeneric,
		)
	}

	registry.StatusUpdate(cfg, wantedDevicesList)

	feedback.PrintResult(cmdResult{CarrierName: carrierName})
	current, next := registry.GetStatus(cfg)
	feedback.PrintResult(showResult{
		CarrierName:    carrierName,
		CurrentDevices: current,
		NextDevices:    next,
	})
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
	// remove once implemented in the configuration file
	if len(dtboFiles) > 0 {
		dtboFiles = append(dtboFiles, filepath.Join(cfg.OverlaysDir().String(), "qrb2210-arduino-imola-carrier-media.dtbo"))
	}

	return dtboFiles
}

func applyOverlays(cfg config.Configuration, ctx context.Context, dtboFiles []string) error {
	if len(dtboFiles) == 0 {
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
