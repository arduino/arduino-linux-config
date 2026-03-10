package carrier

import (
	"context"
	"fmt"
	"os/exec"

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

	// 2. Remove wanted_* markers
	cmd := exec.Command("sh", "-c", "rm -f wanted_*")
	cmd.Dir = stateDir
	if err := cmd.Run(); err != nil {
		feedback.Fatal(fmt.Sprintf("failed to clear markers: %v", err), feedback.ErrGeneric)
	}

	cmd = exec.Command("cp", baseDTB, actualDTB)
	cmd.Dir = overlaysDir
	if err := cmd.Run(); err != nil {
		feedback.Fatal(fmt.Sprintf("failed to restore base DTB: %v", err), feedback.ErrGeneric)
	}
}
