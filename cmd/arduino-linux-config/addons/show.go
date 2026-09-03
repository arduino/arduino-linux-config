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
	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/status"
)

func newShowCmd(reg registry.Registry, cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "show [addon-name]",
		Short: "Show the current configuration",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			addonName := ""
			if len(args) > 0 {
				addonName = args[0]
			}
			showHandler(reg, cfg, addonName)
		},
	}
}

func showHandler(reg registry.Registry, cfg config.Configuration, addonName string) {
	states := getAddonStates(reg, cfg)

	found := false
	result := AddonsShowResult{Addons: make([]AddonStatusResult, 0, len(states))}
	for _, st := range states {
		if addonName != "" && string(st.Name) != addonName {
			continue
		}
		result.Addons = append(result.Addons, toAddonStatusResult(st))
		if addonName != "" {
			found = true
			break
		}
	}

	if addonName != "" && !found {
		feedback.Warnf("addon %s not found", addonName)
	}
	feedback.PrintResult(result)
}

// getAddonStates reads the resolved state for every addon or aborts on error.
func getAddonStates(reg registry.Registry, cfg config.Configuration) []status.AddonState {
	states, err := status.GetAllAddons(cfg, reg)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get addons status: %v", err), feedback.ErrGeneric)
	}
	return states
}

// printAllAddons prints the resolved state for every addon.
func printAllAddons(reg registry.Registry, cfg config.Configuration) {
	states := getAddonStates(reg, cfg)
	result := AddonsShowResult{Addons: make([]AddonStatusResult, 0, len(states))}
	for _, st := range states {
		result.Addons = append(result.Addons, toAddonStatusResult(st))
	}
	feedback.PrintResult(result)
}

func toAddonStatusResult(st status.AddonState) AddonStatusResult {
	return AddonStatusResult{
		Name:           string(st.Name),
		CurrentEnabled: st.Current,
		NextEnabled:    st.Next,
	}
}

type AddonsShowResult struct {
	Addons []AddonStatusResult `json:"addons"`
}

type AddonStatusResult struct {
	Name           string `json:"name"`
	CurrentEnabled bool   `json:"current_enabled"`
	NextEnabled    bool   `json:"next_enabled"`
}

func (r AddonsShowResult) String() string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	for _, addon := range r.Addons {
		fmt.Fprintf(w, "%s\t[current: %s]\t[next: %s]\n",
			addon.Name,
			enabledLabel(addon.CurrentEnabled),
			enabledLabel(addon.NextEnabled),
		)
	}

	w.Flush()
	return b.String()
}

func (r AddonsShowResult) Data() interface{} {
	return r
}

func enabledLabel(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}
