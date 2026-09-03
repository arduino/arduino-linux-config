// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package carrier

import (
	"context"
	"fmt"
	"os"

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

func newDisableCmd(reg registry.Registry, cfg config.Configuration) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "disable <carrier-name>",
		Short: "Disable a carrier and restore the base DTB",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if os.Geteuid() != 0 && !dryRun {
				feedback.Fatal("Command 'disable' must be run as root", feedback.ErrPermissionDenied)
			}

			carrierName := args[0]
			disableHandler(cmd.Context(), reg, cfg, carrierName, dryRun)
		},

		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completion.CompleteCarrierName(reg, args, toComplete)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate the command without applying overlays or writing state")
	return cmd
}

func disableHandler(ctx context.Context, reg registry.Registry, cfg config.Configuration, carrierName string, dryRun bool) {
	carrier, exist := reg.FindByName(carrierName)
	if !exist {
		feedback.Fatal(fmt.Sprintf("carrier %s not supported", carrierName), feedback.ErrBadArgument)
	}

	exec, recorder := executor.Real(), executor.NewRecorder()
	if dryRun {
		exec = recorder
	}

	if err := disable(ctx, exec, cfg, carrier); err != nil {
		feedback.Fatal(fmt.Sprintf("failed to disable carrier %s: %v", carrierName, err), feedback.ErrGeneric)
	}

	if dryRun {
		feedback.PrintResult(dryrun.Result{Subject: fmt.Sprintf("carrier '%s'", carrier.Name), Effects: recorder.Effects()})
		return
	}
	feedback.Warnf("Carrier '%s' disabled (will take effect on next boot)", carrier.Name)

	current, next, err := status.Get(cfg, carrier)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get status for carrier %s: %v", carrierName, err), feedback.ErrGeneric)
	}
	feedback.PrintResult(populateShowResult(carrier, current, next))
}

func disable(ctx context.Context, exec executor.Executor, cfg config.Configuration, carrier registry.Carrier) error {
	baseFiles := overlay.CollectDisabled(carrier)

	board, err := config.GetBoard()
	if err != nil {
		return err
	}

	if err := board.Apply(ctx, exec, baseFiles); err != nil {
		return err
	}

	if err := status.Update(exec, cfg, carrier, status.CarrierStatus{Enable: false}); err != nil {
		return fmt.Errorf("cannot update status: %w", err)
	}
	return nil
}
