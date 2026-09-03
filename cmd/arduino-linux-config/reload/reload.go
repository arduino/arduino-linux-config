// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package reload

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/executor"
	"github.com/arduino/arduino-linux-config/internal/overlay"
	"github.com/arduino/arduino-linux-config/internal/registry"
)

// Re-applies the currently persisted carrier configuration so the
// generated device tree is regenerated. The saved state remains unchanged.
func NewReloadCmd() *cobra.Command {
	cfg := config.New()
	reg := registry.New()

	var dryRun bool
	cmd := &cobra.Command{
		Use:   "reload",
		Short: "Reload the current configuration and regenerate the device tree",
		Long:  "Re-apply the currently persisted carrier configuration without changing the saved state.",
		Example: `  # Reload every configured carrier:
  arduino-linux-config reload`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if os.Geteuid() != 0 && !dryRun {
				feedback.Fatal("Command 'reload' must be run as root", feedback.ErrPermissionDenied)
			}

			reloadHandler(cmd.Context(), reg, cfg, dryRun)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate the command without applying overlays or writing state")
	return cmd
}

// Re-applies to the device tree the persisted carrier configuration
func reloadHandler(ctx context.Context, reg registry.Registry, cfg config.Configuration, dryRun bool) {
	board, err := config.GetBoard()
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	result := reloadResult{
		BoardID:          config.GetBoardID(),
		DryRun:           dryRun,
		ReloadedCarriers: make([]string, 0),
		ReloadedAddons:   make([]string, 0),
	}

	carriersOverlay, carriers := overlay.GetConfiguredCarriersOverlay(cfg, reg)
	addonOverlays, overlayName := overlay.GetConfiguredAddonsOverlay(cfg, reg)

	overlays := make([]string, 0, len(carriersOverlay)+len(addonOverlays))
	overlays = append(overlays, carriersOverlay...)
	overlays = append(overlays, addonOverlays...)

	result.ReloadedCarriers = append(result.ReloadedCarriers, carriers...)
	result.ReloadedAddons = append(result.ReloadedAddons, overlayName)

	exec, recorder := executor.Real(), executor.NewRecorder()
	if dryRun {
		exec = recorder
	}

	if err := board.Apply(ctx, exec, overlays); err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	result.Effects = recorder.Effects()
	feedback.PrintResult(result)
}

type reloadResult struct {
	BoardID          string   `json:"board_id"`
	DryRun           bool     `json:"dry_run"`
	ReloadedCarriers []string `json:"reloaded_carriers"`
	ReloadedAddons   []string `json:"reloaded_addons"`
	Effects          []string `json:"effects,omitempty"`
}

func (r reloadResult) String() string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "Board:\t%s\n", r.BoardID)
	if len(r.ReloadedCarriers) == 0 {
		fmt.Fprintln(w, "No carriers to reload")
	}
	for _, name := range r.ReloadedCarriers {
		fmt.Fprintf(w, "Reloaded carriers:\t%s\n", name)
	}
	for _, name := range r.ReloadedAddons {
		fmt.Fprintf(w, "Reloaded addons:\t%s\n", name)
	}

	if r.DryRun {
		fmt.Fprintln(w, "Dry-run: no changes applied")
		for _, effect := range r.Effects {
			fmt.Fprintln(w, effect)
		}
	}

	w.Flush()
	return b.String()
}

func (r reloadResult) Data() any {
	return r
}
