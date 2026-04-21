package carrier

import (
	"fmt"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/carrier/completion"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/spf13/cobra"
)

func newDisableCmd(cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "disable <carrier-name>",
		Short: "Disable a carrier and restore the base DTB",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			carrierName := args[0]
			disableHandler(registry.New(), cfg, carrierName)
		},

		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completion.CompleteCarrierName(registry.New(), args, toComplete)
		},
	}
}

func disableHandler(reg registry.CarrierRegistry, cfg config.Configuration, carrierName string) {
	carrier, exist := reg.FindByName(carrierName)
	if !exist {
		feedback.Fatal(fmt.Sprintf("carrier %s not supported", carrierName), feedback.ErrBadArgument)
	}

	err := disable(cfg, carrier)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to disable carrier %s: %v", carrierName, err), feedback.ErrGeneric)
	}
	feedback.PrintResult(cmdResult{CarrierName: carrierName})
	current, next, err := registry.GetStatus(cfg, carrier)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get status for carrier %s: %v", carrierName, err), feedback.ErrGeneric)
	}
	feedback.PrintResult(populateShowResult(carrier, current, next))
}

func disable(cfg config.Configuration, carrier registry.Carrier) error {
	baseFiles := make([]string, 0)

	for _, device := range carrier.Devices {
		for _, option := range device.Options {
			if option.Name == string(registry.None) {
				baseFiles = append(baseFiles, option.DtboFiles...)
			}
		}
	}

	// Add disable dtbs to restore original board configuration.
	baseFiles = append(baseFiles, carrier.DisabledDtbos...)

	err := mergeOverlays(cfg, baseFiles)
	if err != nil {
		return fmt.Errorf("cannot merge overlays: %w", err)
	}

	err = registry.StatusUpdate(cfg, carrier, registry.CarrierStatus{Enable: false})
	if err != nil {
		return fmt.Errorf("cannot update status: %w", err)
	}
	return nil
}

type cmdResult struct {
	CarrierName string `json:"carrier_name"`
}

func (r cmdResult) String() string {
	return fmt.Sprintf("Carrier %s disabled (will take effect on next boot)\n", r.CarrierName)
}

func (r cmdResult) Data() any {
	return r
}
