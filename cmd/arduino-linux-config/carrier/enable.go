package carrier

import (
	"context"

	"github.com/spf13/cobra"
)

func newEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enables the specified carrier",
		Run: func(cmd *cobra.Command, args []string) {
			enableHandler(cmd.Context())
		},
	}

	return cmd
}

func enableHandler(_ context.Context) {
	// Implementation for enabling carrier
}
