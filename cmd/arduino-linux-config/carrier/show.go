package carrier

import (
	"context"
	"fmt"
	"strings"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/registry"
	"github.com/arduino/arduino-linux-config/cmd/config"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/spf13/cobra"
)

func newShowCmd(cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "show <carrier-name>",
		Short: "Show the current configuration for a carrier",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			showHandler(cfg, cmd.Context(), args[0])
		},
	}
}

func showHandler(cfg config.Configuration, _ context.Context, carrierName string) {
	if carrierName != registry.MediaCarrierRegistry.Name {
		feedback.Fatal(fmt.Sprintf("carrier %q not supported", carrierName), feedback.ErrGeneric)
	}

	markers, err := loadStateMarkers(cfg)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to read carrier status: %v", err), feedback.ErrGeneric)
	}

	bootTime, err := getBootTime()
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get boot time: %v", err), feedback.ErrGeneric)
	}

	states := make(map[string]deviceState)
	for _, device := range registry.MediaCarrierRegistry.Devices {
		states[string(device.Name)] = deviceState{Current: deviceOptionNone, Next: ""}
	}

	for _, f := range markers {
		state := states[f.Device]
		if f.CreatedAt.After(bootTime) {
			state.Next = f.Option
		} else {
			state.Current = f.Option
		}
		states[f.Device] = state
	}

	feedback.PrintResult(showResult{
		CarrierName: carrierName,
		States:      states,
	})
}

type deviceState struct {
	Current string `json:"current"`
	Next    string `json:"next,omitempty"`
}

type showResult struct {
	CarrierName string                 `json:"carrier_name"`
	States      map[string]deviceState `json:"states"`
}

func (r showResult) String() string {
	var sb strings.Builder
	sb.WriteString(r.CarrierName + "\n")

	for _, device := range registry.MediaCarrierRegistry.Devices {
		state, ok := r.States[string(device.Name)]
		if !ok {
			state = deviceState{Current: deviceOptionNone}
		}

		line := fmt.Sprintf("    %s: [current: %s]", device.Name, state.Current)
		if state.Next != "" {
			line += fmt.Sprintf(" [next boot: %s]", state.Next)
		}
		sb.WriteString(line + "\n")
	}

	return sb.String()
}

func (r showResult) Data() interface{} {
	return r
}
