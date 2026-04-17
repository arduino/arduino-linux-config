package carrier

import (
	"fmt"

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
			resetHandler(cfg, args[0])
		},
	}
}

func resetHandler(cfg config.Configuration, carrierName string) {
	carrier, exist := registry.Registry.FindByName(carrierName)
	if !exist {
		feedback.Fatal(fmt.Sprintf("carrier %s not supported", carrierName), feedback.ErrBadArgument)
	}

	reset(cfg, carrier)
	feedback.PrintResult(cmdResult{CarrierName: carrierName})
	current, next := registry.GetStatus(cfg, carrier)
	feedback.PrintResult(populateShowResult(carrier, current, next))
}

func reset(cfg config.Configuration, carrier registry.Carrier) {
	baseFiles := make([]string, 0)

	for _, device := range carrier.Devices {
		for _, option := range device.Options {
			if option.Name == string(registry.None) {
				baseFiles = append(baseFiles, option.DtboFiles...)
			}
		}
	}

	err := mergeOverlays(cfg, baseFiles)
	if err != nil {
		feedback.Fatal(
			fmt.Sprintf("Error merging overlays: %v", err),
			feedback.ErrGeneric,
		)
	}

	selection := make(map[registry.CarrierDeviceName]string)
	registry.StatusUpdate(cfg, carrier, selection)
}

type cmdResult struct {
	CarrierName string `json:"carrier_name"`
}

func (r cmdResult) String() string {
	return fmt.Sprintf("Carrier %s reset (will take effect on next boot)\n", r.CarrierName)
}

func (r cmdResult) Data() any {
	return r
}
