package completion

import (
	"fmt"
	"strings"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/registry"
	"github.com/spf13/cobra"
)

// CompleteCarrierName provides completion for carrier names on the first argument.
func CompleteCarrierName(registry registry.CarrierRegistry, args []string, partial string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	completions := make([]cobra.Completion, 0, len(registry.Carriers))
	for _, carrier := range registry.Carriers {
		completions = append(completions, cobra.Completion(carrier.Name))
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// CompleteDeviceOption provides completion for device options in the format "device=option".
func CompleteDeviceOption(carrier registry.Carrier, partial string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if !strings.Contains(partial, "=") {
		// Suggest device names for the first part of the argument
		completions := make([]cobra.Completion, 0, len(carrier.Devices))
		for _, device := range carrier.Devices {
			completions = append(completions, cobra.Completion(device.Name+"="))
		}
		return completions, cobra.ShellCompDirectiveNoSpace
	}

	// Suggest options for the specified device.
	completions := make([]cobra.Completion, 0, len(carrier.Devices)*4)
	for _, device := range carrier.Devices {
		for _, option := range device.Options {
			completions = append(completions, fmt.Sprintf("%s=%s", device.Name, option.Name))
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
