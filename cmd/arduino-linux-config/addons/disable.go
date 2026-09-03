// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package addons

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/dryrun"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/executor"
	"github.com/arduino/arduino-linux-config/internal/overlay"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/status"
)

func newDisableCmd(reg registry.Registry, cfg config.Configuration) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "disable [addon-name]",
		Short: "Disable an addon and restore the base DTB. With no addon, disables all of them",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if os.Geteuid() != 0 && !dryRun {
				feedback.Fatal("Command 'disable' must be run as root", feedback.ErrPermissionDenied)
			}
			addonName := ""
			if len(args) > 0 {
				addonName = args[0]
			}
			disableHandler(cmd.Context(), reg, cfg, addonName, dryRun)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate the command without applying changes or writing state")
	return cmd
}

func disableHandler(ctx context.Context, reg registry.Registry, cfg config.Configuration, addonName string, dryRun bool) {
	changes := make(map[registry.AddonName]bool)
	if addonName == "" {
		for _, addon := range reg.Addons {
			changes[addon.Name] = false
		}
	} else {
		addon, exist := reg.FindAddonByName(addonName)
		if !exist {
			feedback.Fatal(fmt.Sprintf("addon %q not supported", addonName), feedback.ErrBadArgument)
		}
		changes[addon.Name] = false
	}

	board, err := config.GetBoard()
	if err != nil {
		feedback.Fatal("board not supported", feedback.ErrBadArgument)
	}

	exec, recorder := executor.Real(), executor.NewRecorder()
	if dryRun {
		exec = recorder
	}

	configuredCarrierOverlays, _ := overlay.GetConfiguredCarriersOverlay(cfg, reg)
	if err := board.Apply(ctx, exec, configuredCarrierOverlays); err != nil {
		feedback.Fatal(fmt.Sprintf("failed to disable addons %v", err), feedback.ErrGeneric)
	}

	if err := status.UpdateAddons(exec, cfg, reg, changes); err != nil {
		feedback.Fatal(fmt.Sprintf("failed to update addons status: %v", err), feedback.ErrGeneric)
	}

	if dryRun {
		feedback.PrintResult(dryrun.Result{Effects: recorder.Effects()})
		return
	}
	feedback.Warnf("Addons disabled (will take effect on next boot)")
	printAllAddons(reg, cfg)
}
