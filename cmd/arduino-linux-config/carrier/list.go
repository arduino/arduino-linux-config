package carrier

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/registry"
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
	carrier := extractCarrierResult(registry.MediaCarrierRegistry)

	feedback.PrintResult(carriersResult{
		MediaCarrier: carrier,
	})
}

type carriersResult struct {
	MediaCarrier CarrierResult `json:"media_carrier"`
}

type CarrierResult struct {
	Name    string         `json:"name"`
	Devices []DeviceResult `json:"devices"`
}

type DeviceResult struct {
	Name             string   `json:"name"`
	AvailableDevices []string `json:"available_devices"`
}

func extractCarrierResult(input registry.MediaCarrier) CarrierResult {
	devices := make([]DeviceResult, 0, len(input.Devices))

	for _, device := range input.Devices {
		options := make([]string, 0, len(device.Options))
		for _, opt := range device.Options {
			options = append(options, opt.Name)
		}

		devices = append(devices, DeviceResult{
			Name:             string(device.Name),
			AvailableDevices: options,
		})
	}

	return CarrierResult{
		Name:    input.Name,
		Devices: devices,
	}
}

func (r carriersResult) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("- %s\n", r.MediaCarrier.Name))

	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)
	for _, dev := range r.MediaCarrier.Devices {
		options := strings.Join(dev.AvailableDevices, ", ")
		fmt.Fprintf(w, "    %s\t: %s\n", dev.Name, options)
	}

	w.Flush()
	return sb.String()
}

func (r carriersResult) Data() interface{} {
	return r
}
