package carrier

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/registry"
	"github.com/arduino/arduino-linux-config/cmd/config"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/spf13/cobra"
)

func newShowCmd(cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "show <carrier-name>",
		Short: "Show the current configuration for a carrier",

		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("missing <carrier-name>. Usage: arduino-linux-config carrier config <carrier-name>")
			}
			return nil
		},
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			showHandler(cfg, args[0])
		},
	}
}

func showHandler(cfg config.Configuration, carrierName string) {
	carrier := registry.Registry.FindByName(carrierName)
	if carrier == nil {
		feedback.Fatal(fmt.Sprintf("carrier %s not supported", carrierName), feedback.ErrBadArgument)
	}
	current, next := registry.GetStatus(cfg, *carrier)
	feedback.PrintResult(populateShowResult(*carrier, current, next))
}

func populateShowResult(carrier registry.Carrier, current []registry.StatusDevice, next []registry.StatusDevice) showResult {
	currentResult := make([]StatusDeviceResult, 0, len(current))
	for _, device := range current {
		currentResult = append(currentResult, StatusDeviceResult{
			Device: device.Device,
			Option: device.Option,
		})
	}
	nextResult := make([]StatusDeviceResult, 0, len(next))
	for _, device := range next {
		nextResult = append(nextResult, StatusDeviceResult{
			Device: device.Device,
			Option: device.Option,
		})
	}
	return showResult{
		CarrierName:    string(carrier.Name),
		CurrentDevices: currentResult,
		NextDevices:    nextResult,
		carrier:        carrier,
	}
}

type showResult struct {
	CarrierName    string               `json:"carrier_name"`
	CurrentDevices []StatusDeviceResult `json:"current"`
	NextDevices    []StatusDeviceResult `json:"next"`

	carrier registry.Carrier `json:"-"`
}

type StatusDeviceResult struct {
	Device     string `json:"device"`
	Option     string `json:"option"`
	DeviceType string `json:"device_type"`
}

func (r showResult) String() string {
	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)
	fmt.Fprintf(&sb, "%s\n", r.CarrierName)

	nextMap := make(map[registry.CarrierDeviceName]string)

	if len(r.NextDevices) > 0 {
		for _, d := range r.carrier.Devices {
			if device, found := hasDevice(r.NextDevices, d.Name); found {
				nextMap[d.Name] = device.Option
			}
		}
	}

	for _, device := range r.carrier.Devices {
		c, _ := hasDevice(r.CurrentDevices, device.Name)

		if len(r.NextDevices) > 0 {
			fmt.Fprintf(w, "  %s:\t[current: %s]\t[next boot: %s]\n", c.Device, c.Option, nextMap[registry.CarrierDeviceName(c.Device)])
		} else {
			fmt.Fprintf(w, "  %s:\t[current: %s]\t\n", c.Device, c.Option)
		}

	}

	w.Flush()
	return sb.String()
}

func (r showResult) Data() interface{} {
	return r
}

func hasDevice(devices []StatusDeviceResult, deviceName registry.CarrierDeviceName) (StatusDeviceResult, bool) {
	for _, d := range devices {
		if d.Device == string(deviceName) {
			return d, true
		}
	}
	return StatusDeviceResult{}, false
}
