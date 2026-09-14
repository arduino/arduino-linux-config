// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package hw

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/dryrun"
	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/hw/completion"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/devicetree"
	"github.com/arduino/arduino-linux-config/internal/executor"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/status"
)

// The JSON of the "carrier" group keeps the shape of the v0.2.x releases, so
// the tools that read it keep working. Only the carriers are reported, because
// the old releases knew nothing about the hats.

func newLegacyListCmd(reg registry.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the carriers available for this board",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			feedback.PrintResult(buildLegacyListResult(reg))
		},
	}
}

func buildLegacyListResult(reg registry.Registry) legacyListResult {
	return legacyListResult{inner: buildListResult(reg.ByKind(registry.KindCarrier))}
}

type legacyListResult struct {
	inner listResult
}

func (r legacyListResult) String() string {
	return r.inner.String()
}

func (r legacyListResult) Data() interface{} {
	return legacyListData(r.inner.Mounts)
}

type legacyCarriersResult struct {
	Carriers []legacyCarrierResult `json:"carriers"`
}

type legacyCarrierResult struct {
	Name    string                      `json:"name"`
	Devices []legacyCarrierDeviceResult `json:"devices"`
}

type legacyCarrierDeviceResult struct {
	Name             string   `json:"name"`
	DeviceType       string   `json:"device_type"`
	AvailableDevices []string `json:"available_devices"`
}

func legacyListData(mounts []listMount) legacyCarriersResult {
	result := legacyCarriersResult{Carriers: make([]legacyCarrierResult, 0, len(mounts))}
	for _, mount := range mounts {
		devices := make([]legacyCarrierDeviceResult, 0, len(mount.Devices))
		for _, device := range mount.Devices {
			devices = append(devices, legacyCarrierDeviceResult{
				Name:             device.Name,
				DeviceType:       device.DeviceType,
				AvailableDevices: device.Options,
			})
		}
		result.Carriers = append(result.Carriers, legacyCarrierResult{
			Name:    mount.Name,
			Devices: devices,
		})
	}
	return result
}

type legacyShowResult struct {
	Carriers []legacyShowCarrierResult `json:"carriers"`
}

type legacyShowCarrierResult struct {
	CarrierName    string         `json:"carrier_name"`
	CurrentEnabled bool           `json:"current_enabled"`
	NextEnabled    bool           `json:"next_enabled"`
	CurrentDevices []deviceResult `json:"current"`
	NextDevices    []deviceResult `json:"next"`
}

func legacyShowData(mounts []showMount) legacyShowResult {
	result := legacyShowResult{Carriers: make([]legacyShowCarrierResult, 0, len(mounts))}
	for _, mount := range mounts {
		result.Carriers = append(result.Carriers, legacyShowMount(mount))
	}
	return result
}

func legacyShowMount(mount showMount) legacyShowCarrierResult {
	return legacyShowCarrierResult{
		CarrierName:    mount.Name,
		CurrentEnabled: mount.CurrentEnabled,
		NextEnabled:    mount.NextEnabled,
		CurrentDevices: mount.CurrentDevices,
		NextDevices:    mount.NextDevices,
	}
}

func newLegacyShowCmd(reg registry.Registry, cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "show [name]",
		Short: "Show the configuration of the board, or of one carrier",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			carriers := reg.ByKind(registry.KindCarrier)
			var mountName string
			if len(args) > 0 && args[0] != "" {
				mountName = string(findMount(carriers, args[0]).Name)
			}
			feedback.PrintResult(buildLegacyShowResult(cfg, carriers, mountName, false))
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completion.CompleteMountName(reg.ByKind(registry.KindCarrier), toComplete)
		},
	}
}

// The v0.2.x enable and disable reported the affected carrier alone, out of any
// list. Set single to keep that shape.
func buildLegacyShowResult(cfg config.Configuration, reg registry.Registry, mountName string, single bool) legacyShow {
	inner := showResult{Mounts: make([]showMount, 0, len(reg.Mounts))}
	for _, mount := range reg.Mounts {
		if mountName != "" && mountName != string(mount.Name) {
			continue
		}
		inner.Mounts = append(inner.Mounts, toShowMount(cfg, mount))
	}
	return legacyShow{inner: inner, single: single}
}

type legacyShow struct {
	inner  showResult
	single bool
}

func (r legacyShow) String() string {
	return r.inner.String()
}

func (r legacyShow) Data() any {
	if r.single && len(r.inner.Mounts) == 1 {
		return legacyShowMount(r.inner.Mounts[0])
	}
	return legacyShowData(r.inner.Mounts)
}

func newLegacyEnableCmd(reg registry.Registry, cfg config.Configuration) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "enable <name> [device=option...]",
		Short: "Enable a carrier, with its device options",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if os.Geteuid() != 0 && !dryRun {
				feedback.Fatal("Command 'enable' must be run as root", feedback.ErrPermissionDenied)
			}
			legacyEnableHandler(cmd.Context(), reg, cfg, args[0], args[1:], dryRun)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			carriers := reg.ByKind(registry.KindCarrier)
			if len(args) == 0 {
				return completion.CompleteMountName(carriers, toComplete)
			}
			mount, exist := carriers.FindByName(args[0])
			if !exist {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completion.CompleteDeviceOption(mount, args[1:], toComplete)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate the command without applying overlays or writing state")
	return cmd
}

func legacyEnableHandler(ctx context.Context, reg registry.Registry, cfg config.Configuration, name string, deviceArgs []string, dryRun bool) {
	carriers := reg.ByKind(registry.KindCarrier)
	mount := findMount(carriers, name)

	selection, err := parseUserArgs(deviceArgs)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrBadArgument)
	}
	if err := validateUserConfiguration(mount, selection); err != nil {
		feedback.Fatal(err.Error(), feedback.ErrBadArgument)
	}

	exec, recorder := executor.Real(), executor.NewRecorder()
	if dryRun {
		exec = recorder
	}

	applyEnable(ctx, reg, cfg, mount, selection, exec)

	if dryRun {
		subject := fmt.Sprintf("%s '%s'", string(mount.Kind), mount.Name)
		feedback.PrintResult(dryrun.Result{Subject: subject, Effects: recorder.Effects()})
		return
	}

	feedback.Warnf("Carrier '%s' enabled (will take effect on next boot)", mount.Name)
	feedback.PrintResult(buildLegacyShowResult(cfg, carriers, string(mount.Name), true))
}

func newLegacyDisableCmd(reg registry.Registry, cfg config.Configuration) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "disable [name]",
		Short: "Disable a carrier. With no name, disable every carrier",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if os.Geteuid() != 0 && !dryRun {
				feedback.Fatal("Command 'disable' must be run as root", feedback.ErrPermissionDenied)
			}
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			legacyDisableHandler(cmd.Context(), reg, cfg, name, dryRun)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completion.CompleteMountName(reg.ByKind(registry.KindCarrier), toComplete)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate the command without applying overlays or writing state")
	return cmd
}

func legacyDisableHandler(ctx context.Context, reg registry.Registry, cfg config.Configuration, name string, dryRun bool) {
	carriers := reg.ByKind(registry.KindCarrier)
	shown := ""
	if name != "" {
		shown = string(findMount(carriers, name).Name)
	}

	desired := devicetree.Desired{}
	for _, mount := range carriers.Mounts {
		if shown == "" || shown == string(mount.Name) {
			desired[mount.Name] = status.MountStatus{Enable: false}
		}
	}

	exec, recorder := executor.Real(), executor.NewRecorder()
	if dryRun {
		exec = recorder
	}

	incompatible, err := devicetree.Rebuild(ctx, exec, reg, cfg, desired)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}
	if len(incompatible) > 0 {
		feedback.Warnf("Incompatible overlays, removing %v", incompatible)
	}

	if dryRun {
		feedback.PrintResult(dryrun.Result{Effects: recorder.Effects()})
		return
	}

	result := buildLegacyShowResult(cfg, carriers, shown, true)
	for _, mount := range result.inner.Mounts {
		feedback.Warnf("Carrier '%s' disabled (will take effect on next boot)", mount.Name)
	}
	feedback.PrintResult(result)
}

func newLegacyReloadCmd(reg registry.Registry, cfg config.Configuration) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "reload",
		Short: "Reload the current configuration and regenerate the device tree",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if os.Geteuid() != 0 && !dryRun {
				feedback.Fatal("Command 'reload' must be run as root", feedback.ErrPermissionDenied)
			}
			legacyReloadHandler(cmd.Context(), reg, cfg, dryRun)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate the command without applying overlays or writing state")
	return cmd
}

func legacyReloadHandler(ctx context.Context, reg registry.Registry, cfg config.Configuration, dryRun bool) {
	result := reloadResult{
		BoardID:          config.GetBoardID(),
		DryRun:           dryRun,
		ReloadedCarriers: make([]string, 0),
		ReloadedHats:     make([]string, 0),
	}

	// Only the enabled carriers are reported: a disabled one adds no overlay.
	for _, mount := range reg.ByKind(registry.KindCarrier).Mounts {
		_, next, err := status.Get(cfg, mount)
		if err != nil {
			feedback.Fatal(fmt.Sprintf("failed to get status for %s: %v", mount.Name, err), feedback.ErrGeneric)
		}
		if !next.Enable {
			continue
		}
		result.ReloadedCarriers = append(result.ReloadedCarriers, string(mount.Name))
	}

	exec, recorder := executor.Real(), executor.NewRecorder()
	if dryRun {
		exec = recorder
	}

	if _, err := devicetree.Rebuild(ctx, exec, reg, cfg, nil); err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	result.Effects = recorder.Effects()
	feedback.PrintResult(result)
}
