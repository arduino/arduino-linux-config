package carrier

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/registry"
	"github.com/arduino/arduino-linux-config/cmd/config"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
)

type CarrierStatus struct {
	Configuration map[registry.CarrierDeviceName]string `json:"configuration,omitempty"`
}

func newConfigureCmd(cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "config <carrier-name> <device=option...>",
		Short: "Configure a carrier with the specified device options",

		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("missing <carrier-name>\nUsage: arduino-linux-config carrier config <carrier-name> <device=option...>")
			}
			if len(args) < 1 {
				return fmt.Errorf("missing carrier configuration\nUsage: arduino-linux-config carrier config <carrier-name>")
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			configHandler(cfg, args[0], args[1:])
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
func configHandler(cfg config.Configuration, carrierName string, deviceArgs []string) {
	nextDevicesConfiguration, err := parseArguments(carrierName, deviceArgs)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	overlayList, err := collectDtboFiles(carrierName, nextDevicesConfiguration)
	if err != nil {
		feedback.Fatal(
			fmt.Sprintf("incompatible configuration: %v", err),
			feedback.ErrGeneric,
		)
	}

	reset(cfg, carrierName)
	err = mergeOverlays(cfg, overlayList)
	if err != nil {
		feedback.Fatal(
			fmt.Sprintf("Error merging overlays: %v", err),
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

func parseArguments(carrierName string, args []string) (map[registry.CarrierDeviceName]string, error) {
	parsedUserSelection := make(map[registry.CarrierDeviceName]string)

	for _, arg := range args {
		// Handle "key=val,key2=val2"
		pairs := strings.Split(arg, ",")

		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}

			// Split the individual pair by "="
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid argument %q: expected device=option format", pair)
			}

			deviceName := parts[0]
			optionName := parts[1]

			if err := validateDeviceOption(carrierName, deviceName, optionName); err != nil {
				return nil, err
			}

			parsedUserSelection[registry.CarrierDeviceName(deviceName)] = optionName
		}
	}

	return parsedUserSelection, nil
}

func collectDtboFiles(carrierName string, userSelection map[registry.CarrierDeviceName]string) ([]string, error) {
	baseFiles := make([]string, 0)
	dtboFiles := make([]string, 0)

	for deviceName, optionName := range userSelection {
		device, _ := registry.FindDevice(carrierName, deviceName)
		for _, option := range device.Options {
			// always add the basic device configuration
			if option.Name == string(registry.None) {
				baseFiles = append(baseFiles, option.DtboFiles...)
			}
			// get the user selected option and collect incompatibilities
			if option.Name == optionName {
				dtboFiles = append(dtboFiles, option.DtboFiles...)
				break
			}
		}
	}

	// check for incompatible overlays in the basic configuration
	// the basic overlay here can be removed in favor of the device overlays
	incompatibleOverlays := getIntersection(dtboFiles, baseFiles)
	if len(incompatibleOverlays) > 0 {
		dtboFiles = slices.DeleteFunc(dtboFiles, func(overlay string) bool {
			return slices.Contains(dtboFiles, overlay)
		})
		feedback.Warnf("Incompatible ovelays, removing %v", incompatibleOverlays)
	}

	return dtboFiles, nil
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

	systemDtb := cfg.SystemDTB()
	overlaysPath := filepath.Dir(systemDtb.String())
	temporaryDtb := filepath.Join(overlaysPath, "qrb2210-arduino-imola.dtb.next")

	for i := range overlays {
		overlays[i] = filepath.Join(overlaysPath, overlays[i])
	}

	// TODO MARTA DELME
	fmt.Println(strings.Join(append([]string{overlayCommand, "-i", cfg.BaseDTB().String(), "-o", temporaryDtb}, overlays...), " "))
	args := append([]string{overlayCommand, "-i", cfg.BaseDTB().String(), "-o", temporaryDtb}, overlays...)
	cmd, err := paths.NewProcess(nil, args...)
	if err != nil {
		return fmt.Errorf("failed to create process: %w", err)
	}

	fmt.Println(strings.Join(append([]string{overlayCommand, "-i", cfg.BaseDTB().String(), "-o", temporaryDtb}, overlays...), " "))
	_, stderr, err := cmd.RunAndCaptureOutput(context.Background())
	if err != nil {
		os.Remove(temporaryDtb)
		return fmt.Errorf("overlay failed: %w\n%s", err, stderr)
	}

	if err := os.Rename(temporaryDtb, systemDtb.String()); err != nil {
		os.Remove(temporaryDtb)
		return fmt.Errorf("failed to move output file: %w", err)
	}

	return nil
}

func validateDeviceOption(carrierName string, rawDevice string, rawOption string) error {
	devices, exists := registry.GetDevices(carrierName)
	if !exists {
		return fmt.Errorf("carrier %q not supported", carrierName)
	}

	device, exists := deviceExists(rawDevice, devices)
	if !exists {
		return fmt.Errorf("unknown device for carrier %s: %q", carrierName, rawDevice)
	}

	if !isOptionValid(rawOption, device) {
		return fmt.Errorf("device %q does not support option %q", rawDevice, rawOption)
	}

	return nil
}

func deviceExists(deviceName string, devices []registry.Device) (registry.Device, bool) {
	for _, device := range devices {
		if device.Name == deviceName {
			return device, true
		}
	}
	return registry.Device{}, false
}

func isOptionValid(optionName string, device registry.Device) bool {
	for _, option := range device.Options {
		if optionName == option.Name {
			return true
		}
	}
	return false
}
