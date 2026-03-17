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
			showHandler(cfg, cmd.Context(), args[0])
		},
	}
}

func showHandler(cfg config.Configuration, _ context.Context, carrierName string) {
	if carrierName != registry.MediaCarrierRegistry.Name {
		feedback.Fatal(fmt.Sprintf("carrier %q not supported", carrierName), feedback.ErrGeneric)
	}

	statuses, err := loadStatus(cfg)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to read carrier status: %v", err), feedback.ErrGeneric)
	}

	bootTime, err := getBootTime()
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get boot time: %v", err), feedback.ErrGeneric)
	}

	current := []deviceResult{}
	next := []deviceResult{}

	for _, f := range statuses {
		if f.CreatedAt.After(bootTime) {
			next = append(next, deviceResult{Device: f.DeviceName, Option: f.Option})
		} else {
			current = append(current, deviceResult{Device: f.DeviceName, Option: f.Option})
		}
	}

	feedback.PrintResult(showResult{
		CarrierName:    carrierName,
		CurrentDevices: current,
		NextDevices:    next,
	})
}

type showResult struct {
	CarrierName    string         `json:"carrier_name"`
	CurrentDevices []deviceResult `json:"current"`
	NextDevices    []deviceResult `json:"next"`
}

type deviceResult struct {
	Device string `json:"device"`
	Option string `json:"option"`
}

func (r showResult) String() string {
	var sb strings.Builder

	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)

	fmt.Fprintf(&sb, "%s\n", r.CarrierName)

	nextMap := make(map[string]string)
	for _, n := range r.NextDevices {
		if n.Device != "" {
			nextMap[n.Device] = n.Option
		}
	}

	processed := make(map[string]bool)
	for _, c := range r.CurrentDevices {
		if c.Device == "" || processed[c.Device] {
			continue
		}

		nextOpt, exists := nextMap[c.Device]
		if !exists {
			nextOpt = "none"
		}

		fmt.Fprintf(w, "  %s:\t[current: %s]\t[next boot: %s]\n", c.Device, c.Option, nextOpt)
		processed[c.Device] = true
	}

	w.Flush()
	return sb.String()
}

func (r showResult) Data() interface{} {
	return r
}
