// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package addons

import (
	"github.com/spf13/cobra"

	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/registry"
)

func NewAddonsCmd() *cobra.Command {
	addonsCmd := &cobra.Command{
		Use:   "addons",
		Short: "Manage Arduino Addons",
		Long:  "Manage Arduino Addons, including listing, configuring, and resetting.",
	}

	cfg := config.New()
	reg := registry.New()

	addonsCmd.AddCommand(newListCmd(reg))
	addonsCmd.AddCommand(newShowCmd(reg, cfg))
	addonsCmd.AddCommand(newEnableCmd(reg, cfg))
	addonsCmd.AddCommand(newDisableCmd(reg, cfg))

	return addonsCmd
}
