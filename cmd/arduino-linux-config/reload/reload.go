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
	"github.com/arduino/arduino-linux-config/internal/status"
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

	carriers := reg.Carriers

	result := reloadResult{
		BoardID:  config.GetBoardID(),
		DryRun:   dryRun,
		Reloaded: make([]string, 0, len(carriers)),
	}

	overlays := make([]string, 0, len(carriers))
	for _, carrier := range carriers {
		// A carrier without a persisted status is reported as disabled
		_, next, err := status.Get(cfg, carrier)
		if err != nil {
			feedback.Fatal(fmt.Sprintf("failed to get status for carrier %s: %v", carrier.Name, err), feedback.ErrGeneric)
		}
		overlays = append(overlays, overlay.CollectForStatus(carrier, next)...)
		result.Reloaded = append(result.Reloaded, string(carrier.Name))
	}

	next, err := status.GetNextConfiguredAddon(cfg, reg)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get addons status: %v", err), feedback.ErrGeneric)
	}
	overlays = append(overlays, overlay.GetDtboForAddon(reg.Addons, next)...)

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
	BoardID  string   `json:"board_id"`
	DryRun   bool     `json:"dry_run"`
	Reloaded []string `json:"reloaded"`
	Effects  []string `json:"effects,omitempty"`
}

func (r reloadResult) String() string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "Board:\t%s\n", r.BoardID)
	if len(r.Reloaded) == 0 {
		fmt.Fprintln(w, "No carriers to reload")
	}
	for _, name := range r.Reloaded {
		fmt.Fprintf(w, "Reloaded carrier:\t%s\n", name)
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
