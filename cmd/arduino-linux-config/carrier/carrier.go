// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package carrier

import (
	"github.com/spf13/cobra"

	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/registry"
)

func NewCarrierCmd(reg registry.Registry) *cobra.Command {
	carrierCmd := &cobra.Command{
		Use:   "carrier",
		Short: "Manage Arduino Carriers",
		Long:  "Manage Arduino Carriers, including listing, configuring, and resetting.",
	}

	cfg := config.New()

	carrierCmd.AddCommand(newListCmd(reg))
	carrierCmd.AddCommand(newEnableCmd(reg, cfg))
	carrierCmd.AddCommand(newDisableCmd(reg, cfg))
	carrierCmd.AddCommand(newShowCmd(reg, cfg))

	return carrierCmd
}
