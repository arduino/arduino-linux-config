package carrier

import (
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
			listHandler()
		},
	}
}

func listHandler() {
	carriersResult := extractCarriersResult()
	feedback.PrintResult(carriersResult)
}

type CarriersResult struct {
	Carriers []CarrierResult `json:"carriers"`
}

type CarrierResult struct {
	Name    string         `json:"name"`
	Devices []DeviceResult `json:"devices"`
}

type DeviceResult struct {
	Name             string   `json:"name"`
	DeviceType       string   `json:"device_type"`
	AvailableDevices []string `json:"available_devices"`
}

func extractCarriersResult() CarriersResult {
	carriersResult := CarriersResult{
		Carriers: make([]CarrierResult, 0, len(registry.Registry.Carriers)),
	}

	for _, carrier := range registry.Registry.Carriers {
		carriersResult.Carriers = append(carriersResult.Carriers, CarrierResult{
			Name:    string(carrier.Name),
			Devices: extractDeviceResult(carrier.Devices),
		})
	}
	return carriersResult
}

func extractDeviceResult(devices []registry.Device) []DeviceResult {
	devicesList := make([]DeviceResult, len(devices))

	for i, d := range devices {
		options := extractOptions(d.Options)
		devicesList[i] = DeviceResult{
			Name:             string(d.Name),
			DeviceType:       string(d.DeviceType),
			AvailableDevices: options,
		}
	}
	return devicesList
}

func extractOptions(options []registry.DeviceOption) []string {
	res := make([]string, len(options))
	for i, opt := range options {
		res[i] = opt.Name
	}
	return res
}

func (r CarriersResult) String() string {
	var b strings.Builder
	// Using a simple builder for headers, and tabwriter for the indented table

	for _, carrier := range r.Carriers {
		// 1. Print the "Header" for each carrier file
		fmt.Fprintf(&b, "%s\n", carrier.Name)

		// 2. Initialize a tabwriter for the indented section
		// We use a new writer per carrier to keep column widths consistent within that group
		w := tabwriter.NewWriter(&b, 0, 8, 2, ' ', 0)

		fmt.Fprintln(w, "  Device\tAvailable options")
		fmt.Fprintln(w, "  ------\t-----------------")

		if len(carrier.Devices) == 0 {
			fmt.Fprintln(w, "  -\t-")
		} else {
			for _, device := range carrier.Devices {
				options := strings.Join(device.AvailableDevices, ", ")
				if options == "" {
					options = "none"
				}
				// Use a leading space/tab for the "indented" look
				fmt.Fprintf(w, "  %s\t%s\n", device.Name, options)
			}
		}

		w.Flush()
		b.WriteString("\n") // Add a newline between different carrier files
	}

	return b.String()
}

func (r CarriersResult) Data() interface{} {
	return r
}
