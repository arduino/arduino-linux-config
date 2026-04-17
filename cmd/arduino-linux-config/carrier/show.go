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
	carrier, exist := registry.Registry.FindByName(carrierName)
	if !exist {
		feedback.Fatal(fmt.Sprintf("carrier %s not supported", carrierName), feedback.ErrBadArgument)
	}
	current, next, err := registry.GetStatus(cfg, carrier)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get status for carrier %s: %v", carrierName, err), feedback.ErrGeneric)
	}

	feedback.PrintResult(populateShowResult(carrier, current, next))
}

func populateShowResult(carrier registry.Carrier, current registry.CarrierStatus, next registry.CarrierStatus) showResult {
	currentResult := make([]StatusDeviceResult, 0, len(current.StatusDevices))
	for _, device := range current.StatusDevices {
		currentResult = append(currentResult, StatusDeviceResult{
			Device: device.Device,
			Option: device.Option,
		})
	}
	nextResult := make([]StatusDeviceResult, 0, len(next.StatusDevices))
	for _, device := range next.StatusDevices {
		nextResult = append(nextResult, StatusDeviceResult{
			Device: device.Device,
			Option: device.Option,
		})
	}
	return showResult{
		CarrierName:    string(carrier.Name),
		CurrentEnabled: current.Enable,
		NextEnabled:    next.Enable,
		CurrentDevices: currentResult,
		NextDevices:    nextResult,
		carrier:        carrier,
	}
}

type showResult struct {
	CarrierName    string               `json:"carrier_name"`
	CurrentEnabled bool                 `json:"current_enabled"`
	NextEnabled    bool                 `json:"next_enabled"`
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

	statusNext := "disabled"
	if r.NextEnabled {
		statusNext = "enabled"
	}
	statusCurrent := "disabled"
	if r.CurrentEnabled {
		statusCurrent = "enabled"
	}
	fmt.Fprintf(&sb, "%s [current: %s]\t[next: %s]\n", r.CarrierName, statusCurrent, statusNext)

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

func (r showResult) Data() any {
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
