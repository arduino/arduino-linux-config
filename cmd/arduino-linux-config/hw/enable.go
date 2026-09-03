// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package hw

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/hw/completion"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/devicetree"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/status"
)

func newEnableCmd(reg registry.Registry, cfg config.Configuration) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "enable <name> [device=option...]",
		Short: "Enable a carrier or an addon, with its device options",
		Example: `  # Configure a media-carrier with an 8 inch display:
  arduino-linux-config hw enable media-carrier display=8-dsi-touch-a

  # Connect the automation addon on the hat connector:
  arduino-linux-config hw enable automation`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if os.Geteuid() != 0 && !dryRun {
				feedback.Fatal("Command 'enable' must be run as root", feedback.ErrPermissionDenied)
			}
			enableHandler(cmd.Context(), reg, cfg, args[0], args[1:], dryRun)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completion.CompleteMountName(reg, toComplete)
			}
			mount, exist := reg.FindByName(args[0])
			if !exist {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completion.CompleteDeviceOption(mount, args[1:], toComplete)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate the command without applying overlays or writing state")
	return cmd
}

// Since a board reboot can occur asynchronously with the configuration, we must
// track both the current and next states.
func enableHandler(ctx context.Context, reg registry.Registry, cfg config.Configuration, name string, deviceArgs []string, dryRun bool) {
	mount := findMount(reg, name)

	selection, err := parseUserArgs(deviceArgs)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrBadArgument)
	}
	if err := validateUserConfiguration(mount, selection); err != nil {
		feedback.Fatal(err.Error(), feedback.ErrBadArgument)
	}

	// The tool keeps one mount of a kind enabled, so the others are disabled.
	// Only the parts that change are reported back to the user.
	desired := devicetree.Desired{mount.Name: {Enable: true, StatusDevices: selection}}
	changed := []registry.Mount{mount}
	for _, other := range reg.ByKind(mount.Kind) {
		if other.Name == mount.Name {
			continue
		}
		desired[other.Name] = status.MountStatus{Enable: false}
		if _, next, err := status.Get(cfg, other); err == nil && next.Enable {
			changed = append(changed, other)
		}
	}

	command, incompatible, err := devicetree.Rebuild(ctx, reg, cfg, desired, dryRun)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}
	if len(incompatible) > 0 {
		feedback.Warnf("Incompatible overlays, removing %v", incompatible)
	}

	if dryRun {
		feedback.Printf("Dry-run: no changes applied for %s '%s'", mount.Kind.Label(), mount.Name)
		feedback.Print(command)
		return
	}

	feedback.Warnf("%s '%s' enabled (will take effect on next boot)", mount.Kind.Label(), mount.Name)
	showHandler(reg, cfg, changed)
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

func validateUserConfiguration(mount registry.Mount, selection []status.StatusDevice) error {
	for _, s := range selection {
		device, exist := mount.FindDeviceByName(registry.DeviceName(s.Device))
		if !exist {
			return fmt.Errorf("unknown device for %s: %q", mount.Name, s.Device)
		}
		if !slices.ContainsFunc(device.Options, func(o registry.DeviceOption) bool { return o.Name == s.Option }) {
			return fmt.Errorf("device %q does not support option %q", s.Device, s.Option)
		}
	}
	return nil
}
