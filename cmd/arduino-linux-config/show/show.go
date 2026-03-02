package show

import (
	"github.com/spf13/cobra"
)

func NewShowCmd() *cobra.Command {
	appCmd := &cobra.Command{
		Use:   "show",
		Short: "Show information about the current system carriers and devices",
		Long:  "A CLI tool to show information about the current system carriers, including devices and device options.",
	}

	appCmd.AddCommand(
	//	newShowCmd()
	)

	return appCmd
}
