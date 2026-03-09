package carrier

import (
	"context"
	"strings"

	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists the available carriers and devices for the current hardware",
		Run: func(cmd *cobra.Command, args []string) {
			listHandler(cmd.Context())
		},
	}
}

func listHandler(_ context.Context) {
	carriers := getCarrierRegistry()
	feedback.PrintResult(carriersResult{Carriers: carriers})
}

type carriersResult struct {
	Carriers []Carrier `json:"carriers"`
}

func (r carriersResult) String() string {
	var sb strings.Builder

	for _, carrier := range r.Carriers {
		sb.WriteString(carrier.Name + ":\n")
		for _, device := range carrier.Devices {
			options := make([]string, len(device.Options))
			for i, opt := range device.Options {
				options[i] = opt.Name
			}
			sb.WriteString("  - ")
			sb.WriteString(device.Name)
			sb.WriteString(": ")
			sb.WriteString(strings.Join(options, " | "))
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

func (r carriersResult) Data() interface{} {
	return r
}
