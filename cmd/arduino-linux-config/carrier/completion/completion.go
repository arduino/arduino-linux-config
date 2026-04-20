package completion

import (
	"fmt"
	"strings"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/registry"
	"github.com/spf13/cobra"
)

func CompleteCarrierName(partial string) ([]cobra.Completion, cobra.ShellCompDirective) {
	// Suggest carrier names for the first argument
	completions := make([]cobra.Completion, len(registry.Registry.Carriers))
	for i, carrier := range registry.Registry.Carriers {
		completions[i] = cobra.Completion(carrier.Name)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func CompleteDeviceOption(carrier registry.Carrier, partial string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if !strings.Contains(partial, "=") {
		// Suggest device names for the first part of the argument
		var completions []cobra.Completion
		for _, device := range carrier.Devices {
			completions = append(completions, cobra.Completion(device.Name+"="))
		}
		return completions, cobra.ShellCompDirectiveNoSpace
	}

	// Suggest options for the specified device
	var completions []cobra.Completion
	for _, device := range carrier.Devices {
		for _, option := range device.Options {
			completions = append(completions, cobra.Completion(fmt.Sprintf("%s=%s", device.Name, option.Name)))
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}
