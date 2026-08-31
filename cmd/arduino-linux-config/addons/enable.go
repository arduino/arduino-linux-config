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

	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/status"
)

func newEnableCmd(reg registry.Registry, cfg config.Configuration) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "enable <addon-name>",
		Short: "Enable an addon",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if os.Geteuid() != 0 && !dryRun {
				feedback.Fatal("Command 'enable' must be run as root", feedback.ErrPermissionDenied)
			}
			enableHandler(cmd.Context(), reg, cfg, args[0], dryRun)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate the command without applying changes or writing state")
	return cmd
}

func enableHandler(ctx context.Context, reg registry.Registry, cfg config.Configuration, addonName string, dryRun bool) {
	addon, exist := reg.FindAddonByName(addonName)
	if !exist {
		feedback.Fatal(fmt.Sprintf("addon %q not supported", addonName), feedback.ErrBadArgument)
	}
	overlayList := addon.EnabledDtbos

	board, err := config.GetBoard()
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	command, err := board.Apply(ctx, overlayList, dryRun)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	if dryRun {
		feedback.Printf("Dry-run: no changes applied for addon '%s'", addon.Name)
		feedback.Print(command)
		return
	}

	if err := status.UpdateAddons(cfg, reg, map[registry.AddonName]bool{addon.Name: true}); err != nil {
		feedback.Fatal(fmt.Sprintf("failed to update status for addon %s: %v", addonName, err), feedback.ErrGeneric)
	}
	feedback.Warnf("Addon '%s' enabled (will take effect on next boot)", addon.Name)
	printAllAddons(reg, cfg)
}
