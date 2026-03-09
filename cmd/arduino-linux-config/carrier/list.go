package carrier

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/carrierinfo"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists the available carriers and devices for the current hardware",
		Run: func(cmd *cobra.Command, args []string) {
			listHandler(cmd.Context())
		},
	}

	return cmd
}

func listHandler(_ context.Context) {
	boardCarrierInfo := carrierinfo.GetAvailableDeviceList()

	carrier := extractCarrierResult(boardCarrierInfo.MediaCarrier)

	feedback.PrintResult(carriersResult{
		MediaCarrier:   carrier,
		BuiltInCarrier: CarrierResult{},
	})
}

// User result structures
type carriersResult struct {
	MediaCarrier   CarrierResult `json:"media_carrier"`
	BuiltInCarrier CarrierResult `json:"builtin_carrier"`
}

type CarrierResult struct {
	Name    string   `json:"name"`
	Devices []Device `json:"devices"`
}

type Device struct {
	Name             string   `json:"name"`
	AvailableDevices []string `json:"available_devices"`
}

func (deviceList carriersResult) String() string {
	var sb strings.Builder

	w := tabwriter.NewWriter(&sb, 0, 8, 2, ' ', 0)
	fmt.Fprintf(&sb, "- %s\n", deviceList.MediaCarrier.Name)
	for _, dev := range deviceList.MediaCarrier.Devices {
		options := strings.Join(dev.AvailableDevices, ", ")
		fmt.Fprintf(w, "\t%s\t%s\n", dev.Name, options)
	}
	w.Flush()
	return sb.String()
}

func (r carriersResult) Data() interface{} {
	return r
}

func extractCarrierResult(input carrierinfo.Carrier) CarrierResult {

	// group hardware data by DeviceName
	grouping := make(map[carrierinfo.MediaCarrierDevice][]string)
	for _, overlay := range input.Overlays {
		device := overlay.DeviceName
		if _, exist := grouping[device]; !exist {
			grouping[device] = append(grouping[device], "none")
		}
		grouping[device] = append(grouping[device], overlay.HardwareData)
	}

	// build the final Devices slice
	devices := make([]Device, 0, len(carrierinfo.MediaCarrierDeviceList))
	for _, device := range carrierinfo.MediaCarrierDeviceList {
		devices = append(devices, Device{
			Name:             string(device),
			AvailableDevices: grouping[device],
		})
	}

	return CarrierResult{
		Name:    input.Name,
		Devices: devices,
	}
}
