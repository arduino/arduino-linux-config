package carrier

import (
	"context"
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
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			showHandler(cmd.Context(), cfg, args[0])
		},
	}
}

func showHandler(_ context.Context, cfg config.Configuration, carrierName string) {
	if carrierName != registry.MediaCarrierRegistry.Name {
		feedback.Fatal(fmt.Sprintf("carrier %q not supported", carrierName), feedback.ErrGeneric)
	}

	current, next := registry.GetStatus(cfg)

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
	if len(r.CurrentDevices) == 0 && len(r.NextDevices) == 0 {
		return fmt.Sprintf("Media carrier %s not yet configured\n", r.CarrierName)
	}

	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)
	fmt.Fprintf(&sb, "%s\n", r.CarrierName)

	nextMap := make(map[registry.MediaCarrierDeviceName]string)

	if len(r.NextDevices) > 0 {
		for _, deviceName := range registry.MediaCarrierDeviceList {
			if device, found := hasDevice(r.NextDevices, deviceName); found {
				nextMap[deviceName] = device.Option
			} else {
				nextMap[deviceName] = "none"
			}
		}
	}

	for _, deviceName := range registry.MediaCarrierDeviceList {
		c, found := hasDevice(r.CurrentDevices, deviceName)
		if !found {
			c = registry.StatusDevice{Device: string(deviceName), Option: "none"}
		}

		if len(r.NextDevices) > 0 {
			fmt.Fprintf(w, "  %s:\t[current: %s]\t[next boot: %s]\n", c.Device, c.Option, nextMap[registry.MediaCarrierDeviceName(c.Device)])
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

func hasDevice(devices []registry.StatusDevice, deviceName registry.MediaCarrierDeviceName) (registry.StatusDevice, bool) {
	for _, d := range devices {
		if d.Device == string(deviceName) {
			return d, true
		}
	}
	return registry.StatusDevice{}, false
}
