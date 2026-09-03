// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package hw

import (
	"context"
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

func newDisableCmd(reg registry.Registry, cfg config.Configuration) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "disable [name]",
		Short: "Disable a carrier or a hat. With no name, disable everything",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if os.Geteuid() != 0 && !dryRun {
				feedback.Fatal("Command 'disable' must be run as root", feedback.ErrPermissionDenied)
			}
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			disableHandler(cmd.Context(), reg, cfg, name, dryRun)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completion.CompleteMountName(reg, toComplete)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate the command without applying overlays or writing state")
	return cmd
}

func disableHandler(ctx context.Context, reg registry.Registry, cfg config.Configuration, name string, dryRun bool) {
	// With no name every mount is disabled, and the whole board is reported.
	shown := ""
	if name != "" {
		shown = string(findMount(reg, name).Name)
	}

	desired := devicetree.Desired{}
	for _, mount := range reg.Mounts {
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

	feedback.Warnf("Disabled (will take effect on next boot)")
	showHandler(cfg, reg, shown)
}
