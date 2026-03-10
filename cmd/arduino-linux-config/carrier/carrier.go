package carrier

import (
	"github.com/spf13/cobra"
)

func NewCarrierCmd() *cobra.Command {
	carrierCmd := &cobra.Command{
		Use:   "carrier",
		Short: "Manage Arduino Carriers",
		Long:  "Manage Arduino Carriers, including listing, enabling, and disabling.",
	}

	carrierCmd.AddCommand(newListCmd())
	carrierCmd.AddCommand(newShowCmd())
	carrierCmd.AddCommand(newEnableCmd())
	carrierCmd.AddCommand(newDisableCmd())

	return carrierCmd
}
