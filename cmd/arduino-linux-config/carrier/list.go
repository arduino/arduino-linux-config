package carrier

import (
	"context"
	"strings"

	"github.com/arduino/arduino-linux-config/cmd/feedback"
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

func listHandler(ctx context.Context) {
	mediaCarrier, err := stubbedMediaCarrier(ctx)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	builtInCarrier, err := stubbedBuiltInCarrier(ctx)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	feedback.PrintResult(carriersResult{
		BuiltInCarrier: builtInCarrier,
		MediaCarrier:   mediaCarrier,
	})

}

type carriersResult struct {
	BuiltInCarrier Carrier `json:"builtin_carrier"`
	MediaCarrier   Carrier `json:"media_carrier"`
}

func (deviceList carriersResult) String() string {
	var sb strings.Builder

	sb.WriteString(deviceList.BuiltInCarrier.Name + ":\n")
	for _, carrier := range deviceList.BuiltInCarrier.Devices {
		sb.WriteString("- ")
		sb.WriteString(carrier.Name)
		sb.WriteString(": ")
		sb.WriteString(strings.Join(carrier.Options, " | "))
		sb.WriteByte('\n')
	}

	sb.WriteString(deviceList.MediaCarrier.Name + ":\n")
	for _, carrier := range deviceList.MediaCarrier.Devices {
		sb.WriteString("- ")
		sb.WriteString(carrier.Name)
		sb.WriteString(": ")
		sb.WriteString(strings.Join(carrier.Options, " | "))
		sb.WriteByte('\n')
	}

	return sb.String()
}

func (r carriersResult) Data() interface{} {
	return r
}

// Stubbed data
type Carrier struct {
	Name    string   `json:"name"`
	Devices []Device `json:"devices"`
}

type Device struct {
	Name    string   `json:"name"`
	Options []string `json:"options"`
}

// nolint:unparam
func stubbedMediaCarrier(_ context.Context) (Carrier, error) {
	return Carrier{
		Name: "media-carrier",
		Devices: []Device{
			{Name: "camera1", Options: []string{"none", "type1-2lane", "type1-4lane", "other-camera"}},
			{Name: "camera2", Options: []string{"none", "type1-2lane", "type1-4lane", "other-camera"}},
			{Name: "camera3", Options: []string{"none", "type1-2lane", "type1-4lane", "other-camera"}},
			{Name: "display1", Options: []string{"none", "8-dsi-touch-a"}},
		},
	}, nil
}

// nolint:unparam
func stubbedBuiltInCarrier(_ context.Context) (Carrier, error) {
	return Carrier{
		Name: "builtin-carrier",
		Devices: []Device{
			{Name: "camera1", Options: []string{"none", "type1-2lane", "type1-4lane", "other-camera"}},
			{Name: "camera2", Options: []string{"none", "type1-2lane", "type1-4lane", "other-camera"}},
			{Name: "camera3", Options: []string{"none", "type1-2lane", "type1-4lane", "other-camera"}},
			{Name: "display1", Options: []string{"none", "8-dsi-touch-a"}},
		},
	}, nil
}
