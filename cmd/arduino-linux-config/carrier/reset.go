package carrier

import (
	"context"
	"fmt"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/registry"
	"github.com/arduino/arduino-linux-config/cmd/config"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/spf13/cobra"
)

func newResetCmd(cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "reset <carrier-name>",
		Short: "Reset a carrier and restore the base DTB",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("missing <carrier-name>. Usage: arduino-linux-config carrier reset <carrier-name>")
			}
			return nil
		},
		SilenceUsage: true,

		Run: func(cmd *cobra.Command, args []string) {
			resetHandler(cfg, cmd.Context(), args[0])
		},
	}
}

func resetHandler(cfg config.Configuration, _ context.Context, carrierName string) {
	if !registry.CarrierExists(carrierName) {
		feedback.Fatal(fmt.Sprintf("carrier %s not supported", carrierName), feedback.ErrBadArgument)
	}

	reset(cfg, carrierName)

	feedback.PrintResult(cmdResult{CarrierName: carrierName})

	current, next := registry.GetStatus(cfg, carrierName)
	feedback.PrintResult(showResult{
		CarrierName:    carrierName,
		CurrentDevices: current,
		NextDevices:    next,
	})
}

func reset(cfg config.Configuration, carrierName string) {
	baseFiles := make([]string, 0)

	devices, _ := registry.GetDevices(*registry.GetCarriers(), carrierName)
	for _, device := range devices {
		for _, option := range device.Options {
			if option.Name == string(registry.None) {
				baseFiles = append(baseFiles, option.DtboFiles...)
			}
		}
	}

	applyOverlays(uniqueStrings(baseFiles))

	selection := make(map[registry.MediaCarrierDeviceName]string)
	registry.StatusUpdate(cfg, carrierName, selection)
}

type cmdResult struct {
	CarrierName string `json:"carrier_name"`
}

func (r cmdResult) String() string {
	return fmt.Sprintf("Carrier %s reset (will take effect on next boot)\n", r.CarrierName)
}

func (r cmdResult) Data() interface{} {
	return r
}
