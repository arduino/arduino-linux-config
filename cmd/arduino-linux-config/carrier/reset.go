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

func newResetCmd(cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "reset <carrier-name>",
		Short: "Reset a carrier and restore the base DTB",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("missing <carrier-name>. Usage: arduino-linux-config carrier reset <carrier-name>")
			}
			return nil
		},
		SilenceUsage: true,

		Run: func(cmd *cobra.Command, args []string) {
			resetHandler(cfg, cmd.Context(), args[0])
		},
	}
}

func resetHandler(cfg config.Configuration, _ context.Context, carrierName string) {
	if carrierName != registry.MediaCarrierRegistry.Name {
		feedback.Fatal(fmt.Sprintf("carrier %s not supported", carrierName), feedback.ErrBadArgument)
	}

	Reset(cfg)

	feedback.PrintResult(cmdResult{CarrierName: carrierName})
	current, next := registry.GetStatus(cfg)
	feedback.PrintResult(showResult{
		CarrierName:    carrierName,
		CurrentDevices: current,
		NextDevices:    next,
	})
}

func Reset(cfg config.Configuration) {
	if err := restoreFactoryDTB(cfg); err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}
	selection := make(map[registry.MediaCarrierDeviceName]string)
	registry.StatusUpdate(cfg, selection)
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

type cmdResult struct {
	CarrierName string `json:"carrier_name"`
}

func (r cmdResult) String() string {
	return fmt.Sprintf("Carrier %s reset (will take effect on next boot)\n", r.CarrierName)
}

func (r cmdResult) Data() interface{} {
	return r
}
