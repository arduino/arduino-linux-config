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
	hwCmd.AddCommand(newShowCmd(reg, cfg, false))
	hwCmd.AddCommand(newEnableCmd(reg, cfg, false))
	hwCmd.AddCommand(newDisableCmd(reg, cfg, false))
	hwCmd.AddCommand(newReloadCmd(reg, cfg, false))

	return hwCmd
}

// NewCarrierCmd is the previous name of the hw group.
// It is used to handle legacy code, it is hidded in the new versions.
func NewCarrierCmd() *cobra.Command {
	cfg := config.New()
	reg := registry.New()

	carrierCmd := &cobra.Command{
		Use:    "carrier",
		Short:  "Manage the carriers connected to the board",
		Hidden: true,
	}

	carrierCmd.AddCommand(newLegacyListCmd(reg))
	carrierCmd.AddCommand(newShowCmd(reg, cfg, true))
	carrierCmd.AddCommand(newEnableCmd(reg, cfg, true))
	carrierCmd.AddCommand(newDisableCmd(reg, cfg, true))
	carrierCmd.AddCommand(newReloadCmd(reg, cfg, true))

	for _, sub := range carrierCmd.Commands() {
		sub.Hidden = true
	}
	return carrierCmd
}

// The legacy "carrier" group selects and reports the carriers only, while the
// device tree is still rebuilt from every mount of the board.
func selected(reg registry.Registry, legacyCarrier bool) registry.Registry {
	if legacyCarrier {
		return reg.ByKind(registry.KindCarrier)
	}
	return reg
}
