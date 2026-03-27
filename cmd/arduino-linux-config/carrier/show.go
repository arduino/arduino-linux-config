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
				return fmt.Errorf("missing <carrier-name>. Usage: arduino-linux-config carrier configure <name>")
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
	if !registry.CarrierExists(carrierName) {
		feedback.Fatal(fmt.Sprintf("carrier %s not supported", carrierName), feedback.ErrBadArgument)
	}

	current, next := registry.GetStatus(cfg, carrierName)

	feedback.PrintResult(showResult{
		CarrierName:    carrierName,
		CurrentDevices: current,
		NextDevices:    next,
	})
}

type showResult struct {
	CarrierName    string                  `json:"carrier_name"`
	CurrentDevices []registry.StatusDevice `json:"current"`
	NextDevices    []registry.StatusDevice `json:"next"`
}

func (r showResult) String() string {
	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)
	fmt.Fprintf(&sb, "%s\n", r.CarrierName)

	nextMap := make(map[registry.CarrierDeviceName]string)

	if len(r.NextDevices) > 0 {
		for _, deviceName := range registry.GetDevicesNames(r.CarrierName) {
			if device, found := hasDevice(r.NextDevices, deviceName); found {
				nextMap[deviceName] = device.Option
			}
		}
	}

	for _, deviceName := range registry.GetDevicesNames(r.CarrierName) {
		c, _ := hasDevice(r.CurrentDevices, deviceName)

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

func hasDevice(devices []registry.StatusDevice, deviceName registry.CarrierDeviceName) (registry.StatusDevice, bool) {
	for _, d := range devices {
		if d.Device == string(deviceName) {
			return d, true
		}
	}
	return registry.StatusDevice{}, false
}
