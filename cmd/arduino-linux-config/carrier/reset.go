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

	err := reset(cfg, carrier)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to reset carrier %s: %v", carrierName, err), feedback.ErrGeneric)
	}
	feedback.PrintResult(cmdResult{CarrierName: carrierName})
	current, next, err := registry.GetStatus(cfg, carrier)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get status for carrier %s: %v", carrierName, err), feedback.ErrGeneric)
	}
	feedback.PrintResult(populateShowResult(carrier, current, next))
}

func reset(cfg config.Configuration, carrier registry.Carrier) error {
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
		return fmt.Errorf("cannot merge overlays: %w", err)
	}

	selection := make(map[registry.CarrierDeviceName]string)
	err = registry.StatusUpdate(cfg, carrier, selection)
	if err != nil {
		return fmt.Errorf("cannot update status: %w", err)
	}
	return nil
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
