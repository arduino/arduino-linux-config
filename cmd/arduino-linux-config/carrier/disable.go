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
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/dto"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/status"
)

func newDisableCmd(reg registry.CarrierRegistry, cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "disable <carrier-name>",
		Short: "Disable a carrier and restore the base DTB",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if os.Geteuid() != 0 {
				feedback.Fatal("Command 'disable' must be run as root", feedback.ErrPermissionDenied)
			}

			carrierName := args[0]
			disableHandler(cmd.Context(), reg, cfg, carrierName)
		},

		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completion.CompleteCarrierName(reg, args, toComplete)
		},
	}
}

func disableHandler(ctx context.Context, reg registry.CarrierRegistry, cfg config.Configuration, carrierName string) {
	carrier, exist := reg.FindByName(carrierName)
	if !exist {
		feedback.Fatal(fmt.Sprintf("carrier %s not supported", carrierName), feedback.ErrBadArgument)
	}

	err := disable(ctx, cfg, carrier)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to disable carrier %s: %v", carrierName, err), feedback.ErrGeneric)
	}

	feedback.Warnf("Carrier '%s' disabled (will take effect on next boot)", carrier.Name)

	current, next, err := status.Get(cfg, carrier)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get status for carrier %s: %v", carrierName, err), feedback.ErrGeneric)
	}
	feedback.PrintResult(populateShowResult(carrier, current, next))
}

func disable(ctx context.Context, cfg config.Configuration, carrier registry.Carrier) error {
	baseFiles := make([]string, 0)

	for _, device := range carrier.Devices {
		for _, option := range device.Options {
			if option.Name == string(registry.None) {
				baseFiles = append(baseFiles, option.DtboFiles...)
			}
		}
	}

	// Add disable dtbs to restore original board configuration.
	baseFiles = append(baseFiles, carrier.DisabledDtbos...)

	err := dto.Apply(ctx, baseFiles)
	if err != nil {
		return err
	}

	err = status.Update(cfg, carrier, status.CarrierStatus{Enable: false})
	if err != nil {
		return fmt.Errorf("cannot update status: %w", err)
	}
	return nil
}
