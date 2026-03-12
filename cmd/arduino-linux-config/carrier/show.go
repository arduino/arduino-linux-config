package carrier

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	// main "github.com/arduino/arduino-linux-config/cmd/arduino-linux-config"
	// "github.com/arduino/arduino-linux-config/cmd/feedback"
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

	configuration, err := readWantedMarkers(cfg.StatusDir().String())
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to read carrier status: %v", err), feedback.ErrGeneric)
	}

	feedback.PrintResult(showResult{
		CarrierName:   carrierName,
		Configuration: configuration,
	})
}

// readWantedMarkers reads the wanted_* files from statusDir and returns
// a map of device -> option for each configured device.
// Devices with no marker file are reported as "none".
func readWantedMarkers(statusDir string) (map[string]string, error) {
	result := make(map[string]string)
	for _, device := range registry.MediaCarrierRegistry.Devices {
		result[string(device.Name)] = deviceOptionNone
	}

	pattern := filepath.Join(statusDir, "wanted_*")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Each marker file is named: wanted_<device>_<option>
	for _, f := range files {
		base := filepath.Base(f)
		rest := strings.TrimPrefix(base, "wanted_")

		idx := strings.Index(rest, "_")
		if idx < 0 {
			continue // malformed marker, skip
		}
		device := rest[:idx]
		option := rest[idx+1:]

		if _, ok := result[device]; ok {
			result[device] = option
		}
	}

	return result, nil
}

// showResult holds the data returned by the show command.
type showResult struct {
	CarrierName   string            `json:"carrier_name"`
	Configuration map[string]string `json:"configuration"`
}

func (r showResult) String() string {
	var sb strings.Builder
	sb.WriteString(r.CarrierName + "\n")

	for _, device := range registry.MediaCarrierRegistry.Devices {
		option, ok := r.Configuration[string(device.Name)]
		if !ok {
			option = deviceOptionNone
		}

		opts := make([]string, 0, len(device.Options))
		for _, opt := range device.Options {
			if opt.Name == option {
				opts = append(opts, "*"+opt.Name)
			} else {
				opts = append(opts, opt.Name)
			}
		}

		sb.WriteString(fmt.Sprintf("    %s: %s\n", device.Name, strings.Join(opts, ", ")))
	}

	return sb.String()
}

func (r showResult) Data() interface{} {
	return r
}
