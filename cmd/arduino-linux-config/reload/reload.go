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

// Re-applies to the device tree the persisted configuration of every mount
func reloadHandler(ctx context.Context, reg registry.Registry, cfg config.Configuration, dryRun bool) {
	result := reloadResult{
		BoardID:  config.GetBoardID(),
		DryRun:   dryRun,
		Reloaded: make([]string, 0, len(reg.Mounts)),
	}
	for _, mount := range reg.Mounts {
		result.Reloaded = append(result.Reloaded, string(mount.Name))
	}

	command, incompatible, err := devicetree.Rebuild(ctx, reg, cfg, nil, dryRun)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}
	if len(incompatible) > 0 {
		feedback.Warnf("Incompatible overlays, removing %v", incompatible)
	}

	if dryRun {
		feedback.Print("Dry-run: no changes applied")
		feedback.Print(command)
	}
	feedback.PrintResult(result)
}

type reloadResult struct {
	BoardID  string   `json:"board_id"`
	DryRun   bool     `json:"dry_run"`
	Reloaded []string `json:"reloaded"`
}

func (r reloadResult) String() string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "Board:\t%s\n", r.BoardID)
	if len(r.Reloaded) == 0 {
		fmt.Fprintln(w, "Nothing to reload")
	}
	for _, name := range r.Reloaded {
		fmt.Fprintf(w, "Reloaded:\t%s\n", name)
	}

	w.Flush()
	return b.String()
}

func (r reloadResult) Data() any {
	return r
}
