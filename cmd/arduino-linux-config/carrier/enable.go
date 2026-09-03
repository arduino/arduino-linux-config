// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package carrier

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/carrier/completion"
	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/dryrun"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/executor"
	"github.com/arduino/arduino-linux-config/internal/overlay"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/status"
)

func newEnableCmd(reg registry.Registry, cfg config.Configuration) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "enable <carrier-name> [device=option...]",
		Short: "Enable and configure a carrier with the specified device options",
		Example: `  # To configure a media-carrier with two cameras attached, one type1:
  arduino-linux-config carrier enable media-carrier camera0=type1-2lanes camera1=type1-4lanes`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if os.Geteuid() != 0 && !dryRun {
				feedback.Fatal("Command 'enable' must be run as root", feedback.ErrPermissionDenied)
			}

			carrierName := args[0]
			deviceOptions := args[1:]

			enableHandler(cmd.Context(), reg, cfg, carrierName, deviceOptions, dryRun)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completion.CompleteCarrierName(reg, args, toComplete)
			}

			carrier, exist := reg.FindByName(args[0])
			if !exist {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			return completion.CompleteDeviceOption(carrier, args[1:], toComplete)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate the command without applying overlays or writing state")
	return cmd
}

// Since a board reboot can occur asynchronously with the carrier configuration,
// we must track both the current and next states.
//
// State management is handled via a status file updated on device configurations.
// This file stores the device name, configuration options, and a string, the boot-id
//
// When a status request occurs, the system compares the boot-id stored in the
// the configuration abd update the current values.
func enableHandler(ctx context.Context, reg registry.Registry, cfg config.Configuration, carrierName string, deviceArgs []string, dryRun bool) {
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

	board, err := config.GetBoard()
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	exec, recorder := executor.Real(), executor.NewRecorder()
	if dryRun {
		exec = recorder
	}

	configuredAddonOverlays, _ := overlay.GetConfiguredAddonsOverlay(cfg, reg)
	overlayList = append(overlayList, configuredAddonOverlays...)
	if err := board.Apply(ctx, exec, overlayList); err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	err = status.Update(exec, cfg, carrier, status.CarrierStatus{
		Enable:        true,
		StatusDevices: nextDevicesConfiguration,
	})
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to update status for carrier %s: %v", carrierName, err), feedback.ErrGeneric)
	}

	if dryRun {
		feedback.PrintResult(dryrun.Result{Subject: fmt.Sprintf("carrier '%s'", carrier.Name), Effects: recorder.Effects()})
		return
	}

	current, next, err := status.Get(cfg, carrier)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get status for carrier %s: %v", carrierName, err), feedback.ErrGeneric)
	}
	feedback.Warnf("Carrier '%s' enabled (will take effect on next boot)", carrier.Name)
	feedback.PrintResult(populateShowResult(carrier, current, next))
}

func parseUserArgs(args []string) ([]status.StatusDevice, error) {
	selection := make([]status.StatusDevice, 0, len(args))
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

			if slices.ContainsFunc(selection, func(s status.StatusDevice) bool {
				return s.Device == deviceName
			}) {
				return nil, fmt.Errorf("duplicate device %q in arguments", deviceName)
			}

			selection = append(selection, status.StatusDevice{
				Device: deviceName,
				Option: optionName,
			})

		}
	}

	return selection, nil
}

// collectDtboFiles returns the overlay files required to enable the carrier with
// the given per-device option selection, warning about any base overlays removed
// because they were incompatible with the selection.
func collectDtboFiles(carrier registry.Carrier, userSelection []status.StatusDevice) []string {
	files, removedIncompatible := overlay.Collect(carrier, userSelection)
	// TODO This should not happens because we are reading from a consistent status
	if len(removedIncompatible) > 0 {
		feedback.Warnf("Incompatible overlays, removing %v", removedIncompatible)
	}
	return files
}

func validateUserConfiguration(carrier registry.Carrier, nextDevicesConfiguration []status.StatusDevice) error {
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
