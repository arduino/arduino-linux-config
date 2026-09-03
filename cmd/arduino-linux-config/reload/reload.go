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
	"github.com/arduino/arduino-linux-config/internal/devicetree"
	"github.com/arduino/arduino-linux-config/internal/executor"
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

// Re-applies to the device tree the persisted configuration of every mount
func reloadHandler(ctx context.Context, reg registry.Registry, cfg config.Configuration, dryRun bool) {
	result := reloadResult{
		BoardID:          config.GetBoardID(),
		DryRun:           dryRun,
		ReloadedCarriers: make([]string, 0),
		ReloadedHats:     make([]string, 0),
	}

	// Only the enabled mounts are reported: a disabled one adds no overlay.
	for _, mount := range reg.Mounts {
		_, next, err := status.Get(cfg, mount)
		if err != nil {
			feedback.Fatal(fmt.Sprintf("failed to get status for %s: %v", mount.Name, err), feedback.ErrGeneric)
		}
		if !next.Enable {
			continue
		}
		if mount.Kind == registry.KindHat {
			result.ReloadedHats = append(result.ReloadedHats, string(mount.Name))
		} else {
			result.ReloadedCarriers = append(result.ReloadedCarriers, string(mount.Name))
		}
	}

	exec, recorder := executor.Real(), executor.NewRecorder()
	if dryRun {
		exec = recorder
	}

	// The incompatible overlays are not reported: reload applies a status that
	// enable and disable already accepted.
	if _, err := devicetree.Rebuild(ctx, exec, reg, cfg, nil); err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	result.Effects = recorder.Effects()
	feedback.PrintResult(result)
}

type reloadResult struct {
	BoardID          string   `json:"board_id"`
	DryRun           bool     `json:"dry_run"`
	ReloadedCarriers []string `json:"reloaded_carriers"`
	ReloadedHats     []string `json:"reloaded_hats"`
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
	for _, name := range r.ReloadedHats {
		fmt.Fprintf(w, "Reloaded hats:\t%s\n", name)
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
