package carrier

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists the available carriers and devices for the current hardware",
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			listHandler(registry.New())
		},
	}
}

func listHandler(reg registry.CarrierRegistry) {
	carriersResult := extractCarriersResult(reg)
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

func extractCarriersResult(reg registry.CarrierRegistry) CarriersResult {
	carriersResult := CarriersResult{
		Carriers: make([]CarrierResult, 0, len(reg.Carriers)),
	}

	for _, carrier := range reg.Carriers {
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
	// minwidth: 0, tabwidth: 0, padding: 4, padchar: ' ', flags: 0
	w := tabwriter.NewWriter(&b, 0, 0, 4, ' ', 0)

	fmt.Fprintln(w, "CARRIER\tDEVICE\tOPTIONS")
	fmt.Fprintln(w, "-------\t------\t-------")

	for _, carrier := range r.Carriers {
		for i, device := range carrier.Devices {
			carrierName := ""
			if i == 0 {
				carrierName = carrier.Name
			}

			fmt.Fprintf(w, "%s\t%s\t%s\n",
				carrierName,
				device.Name,
				strings.Join(device.AvailableDevices, ", "),
			)
		}
		fmt.Fprintln(w, "\t\t")
	}

	w.Flush()
	return b.String()
}

func (r CarriersResult) Data() interface{} {
	return r
}
