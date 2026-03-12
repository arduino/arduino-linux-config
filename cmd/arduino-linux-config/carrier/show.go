package carrier

import (
	"context"

	// main "github.com/arduino/arduino-linux-config/cmd/arduino-linux-config"
	// "github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	appCmd := &cobra.Command{
		Use:   "show",
		Short: "Show information about the current system carriers and devices",
		Long:  "A CLI tool to show information about the current system carriers, including devices and device options.",
		Run: func(cmd *cobra.Command, args []string) {
			showHandler(cmd.Context())
		},
	}

	return appCmd
}

// type CarriersStatusResult struct {
// 	ActualStatus main.CarrierStatus `json:"actual_status"`
// 	WantedStatus main.CarrierStatus `json:"wanted_status"`
// }

func showHandler(_ context.Context) {

	lines := []string{
		"media-carrier status",
		"",
		"- camera1:  current state: none         | next reboot: type1-4lane",
		"- camera2:  current state: type1-2lane",
		"- camera3:  current state: none         | next reboot: type2-1lane",
		"- display1: current state: none         | next reboot: 8-dsi-touch-a",
	}

	// actualStatus, _ := main.LoadCarriers(main.ActualFile)
	// wantedStatus, _ := main.LoadCarriers(main.WantedFile)

	for _, line := range lines {
		feedback.Print(line)
	}

	// feedback.PrintResult(CarriersStatusResult{ActualStatus: actualStatus, WantedStatus: wantedStatus})
}
