package carrier

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/registry"
	"github.com/arduino/arduino-linux-config/cmd/config"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/spf13/cobra"
)

func newDisableCmd(cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "disable <carrier-name>",
		Short: "Disable a carrier and restore the base DTB",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			disableHandler(cfg, cmd.Context(), args[0])
		},
	}
}

func disableHandler(cfg config.Configuration, _ context.Context, carrierName string) {
	// 1. Validate carrier name
	if carrierName != registry.MediaCarrierRegistry.Name {
		feedback.Fatal("carrier "+carrierName+" not supported", feedback.ErrGeneric)
	}

	// 2. Restore base DTB by copying it over the actual DTB
	tmp := cfg.ActualDTB().String() + ".tmp"

	data, err := os.ReadFile(cfg.BaseDTB().String())
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to read base DTB: %v", err), feedback.ErrGeneric)
	}
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		feedback.Fatal(fmt.Sprintf("failed to write DTB: %v", err), feedback.ErrGeneric)
	}
	if err := os.Rename(tmp, cfg.ActualDTB().String()); err != nil {
		feedback.Fatal(fmt.Sprintf("failed to rename DTB: %v", err), feedback.ErrGeneric)
	}

	// 3. Remove wanted_* markers
	files, _ := filepath.Glob(filepath.Join(cfg.StatusDir().String(), "wanted_*"))
	for _, f := range files {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			feedback.Fatal(fmt.Sprintf("failed to remove marker %q: %v", f, err), feedback.ErrGeneric)
		}
	}
}
