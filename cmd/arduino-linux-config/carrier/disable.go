package carrier

import (
	"context"
	"fmt"
	"os"

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
	if carrierName != registry.MediaCarrierRegistry.Name {
		feedback.Fatal("carrier "+carrierName+" not supported", feedback.ErrGeneric)
	}

	statuses, err := registry.LoadStatus(cfg)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	bootTime, err := registry.GetBootTime()
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	for _, f := range statuses {
		if f.CreatedAt.After(bootTime) {
			if err := f.Path.Remove(); err != nil {
				feedback.Fatal(fmt.Sprintf("failed to remove marker: %v", err), feedback.ErrGeneric)
			}
		}
	}

	for _, deviceName := range registry.MediaCarrierDeviceList {
		overlayFileName := fmt.Sprintf("%s-%s", deviceName, "none")
		if err := registry.CreateStatusFile(cfg, string(overlayFileName)); err != nil {
			feedback.Fatal(err.Error(), feedback.ErrGeneric)
		}
	}

	if err := restoreFactoryDTB(cfg); err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}
}

func restoreFactoryDTB(cfg config.Configuration) error {
	tmp := cfg.ActualDTB().String() + ".tmp"
	data, err := os.ReadFile(cfg.FactoryDTB().String())
	if err != nil {
		return fmt.Errorf("failed to read base DTB: %w", err)
	}
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("failed to write DTB: %w", err)
	}
	if err := os.Rename(tmp, cfg.ActualDTB().String()); err != nil {
		return fmt.Errorf("failed to rename DTB: %w", err)
	}
	return nil
}
