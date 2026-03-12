package carrier

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/spf13/cobra"
)

func newDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <carrier-name>",
		Short: "Disable a carrier and restore the base DTB",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			disableHandler(cmd.Context(), args[0])
		},
	}
}

func disableHandler(_ context.Context, carrierName string) {
	// 1. Validate carrier name
	if carrierName != MediaCarrierRegistry.Name {
		feedback.Fatal("carrier "+carrierName+" not supported", feedback.ErrGeneric)
	}

	// 2. Restore base DTB by copying it over the actual DTB
	tmp := actualDTB + ".tmp"

	data, err := os.ReadFile(baseDTB)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to read base DTB: %v", err), feedback.ErrGeneric)
	}
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		feedback.Fatal(fmt.Sprintf("failed to write DTB: %v", err), feedback.ErrGeneric)
	}
	if err := os.Rename(tmp, actualDTB); err != nil {
		feedback.Fatal(fmt.Sprintf("failed to rename DTB: %v", err), feedback.ErrGeneric)
	}

	// 3. Remove wanted_* markers
	files, _ := filepath.Glob(filepath.Join(stateDir, "wanted_*"))
	for _, f := range files {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			feedback.Fatal(fmt.Sprintf("failed to remove marker %q: %v", f, err), feedback.ErrGeneric)
		}
	}
}
