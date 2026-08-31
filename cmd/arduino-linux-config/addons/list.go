// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package addons

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/registry"
)

func newListCmd(reg registry.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists the available addons",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			listHandler(reg)
		},
	}
}

func listHandler(reg registry.Registry) {
	result := AddonsListResult{Addons: make([]string, 0, len(reg.Addons))}
	for _, addon := range reg.Addons {
		result.Addons = append(result.Addons, string(addon.Name))
	}
	feedback.PrintResult(result)
}

type AddonsListResult struct {
	Addons []string `json:"addons"`
}

func (r AddonsListResult) String() string {
	var b strings.Builder
	// minwidth: 0, tabwidth: 0, padding: 4, padchar: ' ', flags: 0
	w := tabwriter.NewWriter(&b, 0, 0, 4, ' ', 0)

	fmt.Fprintln(w, "ADDON")
	fmt.Fprintln(w, "-----")

	for _, addon := range r.Addons {
		fmt.Fprintln(w, addon)
	}

	w.Flush()
	return b.String()
}

func (r AddonsListResult) Data() interface{} {
	return r
}
