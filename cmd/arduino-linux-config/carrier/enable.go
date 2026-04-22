package carrier

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/carrier/completion"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
)

func newEnableCmd(reg registry.CarrierRegistry, cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "enable <carrier-name> [device=option...]",
		Short: "Enable and configure a carrier with the specified device options",
		Example: `  # To configure a media-carrier with two cameras attached, one type1:
  arduino-linux-config carrier enable media-carrier camera0=type1-2lanes camera1=type1-4lanes`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			carrierName := args[0]
			deviceOptions := args[1:]

			enableHandler(reg, cfg, carrierName, deviceOptions)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completion.CompleteCarrierName(reg, args, toComplete)
			}

			carrier, exist := reg.FindByName(args[0])
			if !exist {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			return completion.CompleteDeviceOption(carrier, toComplete)
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
func enableHandler(reg registry.CarrierRegistry, cfg config.Configuration, carrierName string, deviceArgs []string) {
	nextDevicesConfiguration, err := parseUserArgs(deviceArgs)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	carrier, exist := reg.FindByName(carrierName)
	if !exist {
		feedback.Fatal(fmt.Sprintf("carrier %q not supported", carrierName), feedback.ErrGeneric)
	}

	err = validateUserConfiguration(carrier, nextDevicesConfiguration)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	overlayList := collectDtboFiles(carrier, nextDevicesConfiguration)

	err = disable(cfg, carrier)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to reset carrier %s: %v", carrierName, err), feedback.ErrGeneric)
	}
	err = mergeOverlays(cfg, overlayList)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	err = registry.StatusUpdate(cfg, carrier, registry.CarrierStatus{
		Enable:        true,
		StatusDevices: nextDevicesConfiguration,
	})
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to update status for carrier %s: %v", carrierName, err), feedback.ErrGeneric)
	}

	fmt.Printf("Carrier %s enabled (will take effect on next boot)\n", carrier.Name)

	current, next, err := registry.GetStatus(cfg, carrier)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get status for carrier %s: %v", carrierName, err), feedback.ErrGeneric)
	}
	feedback.PrintResult(populateShowResult(carrier, current, next))
}

func parseUserArgs(args []string) ([]registry.StatusDevice, error) {
	selection := make([]registry.StatusDevice, 0, len(args))
	for _, arg := range args {
		// Handle "key=val,key2=val2"
		pairs := strings.Split(arg, ",")

		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}

			parts := strings.Split(pair, "=")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid argument %q: expected device=option format", pair)
			}

			deviceName, optionName := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			selection = append(selection, registry.StatusDevice{
				Device: deviceName,
				Option: optionName,
			})

		}
	}

	return selection, nil
}

func collectDtboFiles(carrier registry.Carrier, userSelection []registry.StatusDevice) []string {
	var baseFiles, dtboFiles, incompatibleFiles []string

	for _, selection := range userSelection {
		device, exist := carrier.FindDeviceByName(registry.CarrierDeviceName(selection.Device))
		if !exist {
			continue
		}
		for _, option := range device.Options {
			// get the user selected option and collect incompatibilities
			if option.Name == selection.Option {
				dtboFiles = append(dtboFiles, option.DtboFiles...)
				incompatibleFiles = append(incompatibleFiles, option.IncompatibleDtbo...)
				break
			}
		}
	}

	// collect base dtbo files for the media carrier.
	baseFiles = append(baseFiles, carrier.EnabledDtbos...)

	// check for incompatible overlays in the basic configuration
	// in this case, the basic overlay can be removed in favor of the device overlays
	incompatibleOverlays := getIntersection(baseFiles, incompatibleFiles)

	// remove incompatible layer and proceed
	if len(incompatibleOverlays) > 0 {
		baseFiles = slices.DeleteFunc(baseFiles, func(overlay string) bool {
			return slices.Contains(incompatibleFiles, overlay)
		})
		feedback.Warnf("Incompatible ovelays, removing %v", incompatibleOverlays)
	}

	return append(dtboFiles, baseFiles...)
}

func getIntersection(a, b []string) []string {
	var result []string
	for _, v := range a {
		if slices.Contains(b, v) {
			result = append(result, v)
		}
	}
	slices.Sort(result)
	return slices.Compact(result)
}

var overlayCommand = "/usr/bin/fdtoverlay"

func mergeOverlays(cfg config.Configuration, overlays []string) error {
	if len(overlays) == 0 {
		return nil
	}

	slices.Sort(overlays)
	overlays = slices.Compact(overlays)

	temporaryDtb := cfg.DtbDir().Join("qrb2210-arduino-imola.dtb.next")
	defer func() { _ = temporaryDtb.Remove() }()

	for i := range overlays {
		overlays[i] = cfg.DtbDir().Join(overlays[i]).String()
	}

	args := append([]string{overlayCommand, "-i", cfg.BaseDTB().String(), "-o", temporaryDtb.String()}, overlays...)
	cmd, err := paths.NewProcess(nil, args...)
	if err != nil {
		return fmt.Errorf("failed to create process: %w", err)
	}

	_, stderr, err := cmd.RunAndCaptureOutput(context.Background())
	if err != nil {
		return fmt.Errorf("overlay failed: %w\n%s", err, stderr)
	}

	if err := temporaryDtb.Rename(cfg.SystemDTB()); err != nil {
		return fmt.Errorf("failed to move output file: %w", err)
	}

	return nil
}

func validateUserConfiguration(carrier registry.Carrier, nextDevicesConfiguration []registry.StatusDevice) error {
	for _, selection := range nextDevicesConfiguration {
		device, exist := carrier.FindDeviceByName(registry.CarrierDeviceName(selection.Device))
		if !exist {
			return fmt.Errorf("unknown device for carrier %s: %q", carrier.Name, selection.Device)
		}
		if !isOptionValid(selection.Option, device) {
			return fmt.Errorf("device %q does not support option %q", selection.Device, selection.Option)
		}
	}
	return nil
}

func isOptionValid(optionName string, device registry.Device) bool {
	return slices.ContainsFunc(device.Options, func(o registry.DeviceOption) bool {
		return o.Name == optionName
	})
}
