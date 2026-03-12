package carrier

import (
	"github.com/arduino/arduino-linux-config/cmd/config"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/spf13/cobra"
)

func NewCarrierCmd() *cobra.Command {
	carrierCmd := &cobra.Command{
		Use:   "carrier",
		Short: "Manage Arduino Carriers",
		Long:  "Manage Arduino Carriers, including listing, enabling, and disabling.",
	}

	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	carrierCmd.AddCommand(newListCmd())
	carrierCmd.AddCommand(newEnableCmd(cfg))
	carrierCmd.AddCommand(newDisableCmd(cfg))
	carrierCmd.AddCommand(newShowCmd(cfg))

	return carrierCmd
}
