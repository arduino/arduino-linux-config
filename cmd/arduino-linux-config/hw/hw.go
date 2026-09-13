// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

// Package hw implements the commands that list, enable, disable and show the
// parts plugged into the board. A carrier and a hat use different connectors,
// but the same model, the same status files and the same device tree. The user
// selects a part by name, and the output shows what kind of part it is.
package hw

import (
	"github.com/spf13/cobra"

	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/registry"
)

// NewHwCmd groups the commands that configure the parts connected to the board.
// The board itself keeps its own command group.
func NewHwCmd() *cobra.Command {
	cfg := config.New()
	reg := registry.New()

	hwCmd := &cobra.Command{
		Use:     "hw",
		Aliases: []string{"hardware"},
		Short:   "Manage the carriers and the hats connected to the board",
		Long:    "Manage the carriers and the hats connected to the board, including listing, configuring and resetting.",
	}

	hwCmd.AddCommand(newListCmd(reg, false))
	hwCmd.AddCommand(newShowCmd(reg, cfg, false))
	hwCmd.AddCommand(newEnableCmd(reg, cfg, false))
	hwCmd.AddCommand(newDisableCmd(reg, cfg, false))
	hwCmd.AddCommand(newReloadCmd(reg, cfg, false))

	return hwCmd
}

// NewCarrierCmd is the previous name of the hw group. It stays out of the help
// and of the completion, but it keeps working so that the existing scripts do
// not break, with the behaviour of the v0.2.x releases.
func NewCarrierCmd() *cobra.Command {
	cfg := config.New()
	reg := registry.New()

	carrierCmd := &cobra.Command{
		Use:    "carrier",
		Short:  "Manage the carriers connected to the board",
		Hidden: true,
	}

	carrierCmd.AddCommand(newListCmd(reg, true))
	carrierCmd.AddCommand(newShowCmd(reg, cfg, true))
	carrierCmd.AddCommand(newEnableCmd(reg, cfg, true))
	carrierCmd.AddCommand(newDisableCmd(reg, cfg, true))
	carrierCmd.AddCommand(newReloadCmd(reg, cfg, true))

	for _, sub := range carrierCmd.Commands() {
		sub.Hidden = true
	}
	return carrierCmd
}
