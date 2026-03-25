package carrier

import (
	"context"
	"fmt"
	"strings"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/registry"
	"github.com/arduino/arduino-linux-config/cmd/config"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
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
// we must track both the current and next states.
//
// State management is handled via a status file updated on device configurations.
// This file stores the device name, configuration options, and a metadata timestamp.
//
// When a status request occurs, the system compares the last boot time with
// the configuration timestamp to update the current and next states.
func configureHandler(ctx context.Context, cfg config.Configuration, carrierName string, deviceArgs []string) {
	nextDevicesConfiguration, err := parseArguments(carrierName, deviceArgs)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	overlayList := collectDtboFiles(cfg, nextDevicesConfiguration)

	reset(cfg, carrierName)

	if err := applyOverlays(ctx, cfg, overlayList); err != nil {
		feedback.Fatal(
			fmt.Sprintf("failed to apply overlays (carrier has been reset): %v", err),
			feedback.ErrGeneric,
		)
	}

	registry.StatusUpdate(cfg, carrierName, nextDevicesConfiguration)

	feedback.PrintResult(cmdResult{CarrierName: carrierName})
	current, next := registry.GetStatus(cfg, carrierName)
	feedback.PrintResult(showResult{
		CarrierName:    carrierName,
		CurrentDevices: current,
		NextDevices:    next,
	})
}

func parseArguments(carrierName string, args []string) (map[registry.MediaCarrierDeviceName]string, error) {
	parsedUserSelection := make(map[registry.MediaCarrierDeviceName]string)

	for _, arg := range args {
		// 1. Split by comma first to handle "key=val,key2=val2"
		pairs := strings.Split(arg, ",")

		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}

			// 2. Now split the individual pair by "="
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid argument %q: expected device=option format", pair)
			}

			deviceName := parts[0]
			optionName := parts[1]

			isValid, err := registry.ValidateInput(carrierName, deviceName, optionName)
			if !isValid {
				return nil, err
			}

			parsedUserSelection[registry.MediaCarrierDeviceName(deviceName)] = optionName
		}
	}

	return parsedUserSelection, nil
}

// TODO: implement the dtbo collect logic
// For each device:
// start from the dtbo of the "none" configuration
// collect the "dtboFiles" to a list
// collect the "incompatibleDtbo" to a list
// if the intersection between dtboFiles collection and incompatibleDtbo collection is not null the configuration can't be done
// else remove the incompatibleDtbo list from the dtboFiles and create the final dtbo collection.
func collectDtboFiles(cfg config.Configuration, selection map[registry.MediaCarrierDeviceName]string) []string {
	var dtboFiles []string
	fmt.Printf("%v %s", cfg, selection)

	// TODO
	return dtboFiles
}

func applyOverlays(_ context.Context, _ config.Configuration, dtboFiles []string) error {
	fmt.Printf("TODO Executing fdtoverlay\n")

	if len(dtboFiles) == 0 {
		return nil
	}

	// args := append([]string{"fdtoverlay", "-i", cfg.SystemDTB().String(), "-o", cfg.SystemDTB().String()}, dtboFiles...)

	// proc, err := paths.NewProcess(nil, args...)
	// if err != nil {
	// 	return fmt.Errorf("failed to create process: %w", err)
	// }

	// stdout, stderr, err := proc.RunAndCaptureOutput(ctx)
	// if err != nil {
	// 	return fmt.Errorf("fdtoverlay failed: %w\n%s", err, stderr)
	// }
	// if len(stdout) > 0 {
	// 	feedback.Print(string(stdout))
	// }
	return nil
}
