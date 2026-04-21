package carrier

import (
	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/spf13/cobra"
)

func NewCarrierCmd() *cobra.Command {
	carrierCmd := &cobra.Command{
		Use:   "carrier",
		Short: "Manage Arduino Carriers",
		Long:  "Manage Arduino Carriers, including listing, configuring, and resetting.",
	}

	cfg := config.New()
	reg := registry.New()

	carrierCmd.AddCommand(newListCmd(reg))
	carrierCmd.AddCommand(newEnableCmd(reg, cfg))
	carrierCmd.AddCommand(newDisableCmd(reg, cfg))
	carrierCmd.AddCommand(newShowCmd(reg, cfg))

	return carrierCmd
}
