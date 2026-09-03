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
	"fmt"

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

	hwCmd.AddCommand(newListCmd(reg))
	hwCmd.AddCommand(newShowCmd(reg, cfg))
	hwCmd.AddCommand(newEnableCmd(reg, cfg))
	hwCmd.AddCommand(newDisableCmd(reg, cfg))

	return hwCmd
}

// NewCarrierCmd is the previous name of the hw group. Cobra prints the
// deprecation on stderr, and keeps the command out of the help and of the
// completion.
func NewCarrierCmd() *cobra.Command {
	carrierCmd := NewHwCmd()
	carrierCmd.Use = "carrier"
	carrierCmd.Aliases = nil
	carrierCmd.Deprecated = `use "hw" instead`

	for _, sub := range carrierCmd.Commands() {
		sub.Deprecated = fmt.Sprintf("use %q instead of %q", "hw "+sub.Name(), "carrier "+sub.Name())
	}
	return carrierCmd
}
