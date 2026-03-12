package carrier

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/registry"
	"github.com/arduino/arduino-linux-config/cmd/config"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
)

type CarrierStatus struct {
	Configuration map[registry.MediaCarrierDevice]string `json:"configuration,omitempty"`
}

// Since a board reboot can occur asynchronously with the carrier configuration,
// we must track both the current and desired states.
//
// This is managed by maintaining two status files,
// representing the actual and desired configurations,
// that will be synchronized during the next boot sequence.
//
// At boot time we:
// mv wanted.json actual.json

// TODO: update these paths with the real ones
/*
how to create test environment for this code:
 cd /tmp/
 mkdir test_media_carrier
 cd test_media_carrier/
 cp /usr/lib/linux-image-6.16.7-gd1b1a80fb764/qcom/qrb2210-arduino-imola*.dt* /tmp/test_media_carrier/
 mkdir status

should have now:
-rw-r--r-- 1 arduino arduino 71925 Mar 10 16:18 qrb2210-arduino-imola-base.dtb
-rw-r--r-- 1 arduino arduino  2230 Mar 10 16:18 qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-2lanes.dtbo
-rw-r--r-- 1 arduino arduino  2246 Mar 10 16:18 qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-4lanes.dtbo
-rw-r--r-- 1 arduino arduino  2255 Mar 10 16:18 qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-2lanes.dtbo
-rw-r--r-- 1 arduino arduino  2271 Mar 10 16:18 qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-4lanes.dtbo
-rw-r--r-- 1 arduino arduino  4118 Mar 10 16:18 qrb2210-arduino-imola-carrier-media.dtbo
-rw-r--r-- 1 arduino arduino  2623 Mar 10 16:18 qrb2210-arduino-imola-carrier-media-panel-8in-touch-a-dsi.dtbo
-rw-r--r-- 1 arduino arduino 72384 Mar 10 16:18 qrb2210-arduino-imola.dtb
drwxrwxr-x 2 arduino arduino    40 Mar 10 16:21 status


const (
	statusDir   = "/tmp/test_media_carrier/status"
	baseDTB     = "/tmp/test_media_carrier/qrb2210-arduino-imola-base.dtb"
	actualDTB   = "/tmp/test_media_carrier/qrb2210-arduino-imola.dtb"
	overlaysDir = "/tmp/test_media_carrier"
)
*/
func newEnableCmd(cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "enable <carrier-name> [device=option...]",
		Short: "Enable a carrier with the specified device options",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			enableHandler(cfg, cmd.Context(), args[0], args[1:])
		},
	}
}

func enableHandler(cfg config.Configuration, ctx context.Context, carrierName string, deviceArgs []string) {
	wantedDevicesList, err := parseAndValidateDeviceArgs(carrierName, deviceArgs)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	wantedDtboFiles := collectDtboFiles(cfg, wantedDevicesList)

	if len(wantedDtboFiles) > 0 {
		if err := applyOverlays(cfg, ctx, wantedDtboFiles); err != nil {
			feedback.Fatal(err.Error(), feedback.ErrGeneric)
		}
	}

	createWantedMarkers(cfg, wantedDevicesList)
}

// parseAndValidateDeviceArgs does the following:
// - checks carrierName is valid
// - for each device=option argument:
//   - splits into device and option
//   - validates device exists for this carrier
//   - validates option exists for this device
//   - stores selection in map
//
// example input: "media-carrier", ["camera1=type1-2lane", "display1=8-dsi-touch-a"]
// example output: {Camera1: "type1-2lane", Display1: "8-dsi-touch-a"}
func parseAndValidateDeviceArgs(carrierName string, args []string) (map[registry.MediaCarrierDevice]string, error) {
	// we support only media-carrier for now,builtin in the future
	if carrierName != registry.MediaCarrierRegistry.Name {
		return nil, fmt.Errorf("carrier %q not supported", carrierName)
	}
	selection := make(map[registry.MediaCarrierDevice]string)
	for _, arg := range args {
		arg = strings.TrimRight(arg, ",")

		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid argument %q: expected device=option format", arg)
		}

		deviceName := parts[0]
		optionName := parts[1]

		deviceCarrierName, err := ParseMediaCarrierDevice(deviceName)
		if err != nil {
			return nil, fmt.Errorf("device %q not supported for carrier %q", deviceName, carrierName)
		}

		device, err := findDevice(registry.MediaCarrierRegistry, deviceCarrierName)
		if err != nil {
			return nil, err
		}
		option, err := findOption(device, optionName)
		if err != nil {
			return nil, err
		}
		//example: selection["camera1"] = "type1-2lane"
		selection[deviceCarrierName] = option.Name
	}

	return selection, nil
}
func findDevice(carrier registry.MediaCarrier, deviceName registry.MediaCarrierDevice) (registry.Device, error) {
	for _, d := range carrier.Devices {
		if d.Name == deviceName {
			return d, nil
		}
	}
	return registry.Device{}, fmt.Errorf("device %q not found in carrier %q", deviceName, carrier.Name)
}

func findOption(device registry.Device, optionName string) (registry.DeviceOption, error) {
	for _, o := range device.Options {
		if o.Name == optionName {
			return o, nil
		}
	}
	return registry.DeviceOption{}, fmt.Errorf("option %q not found for device %q", optionName, device.Name)
}

func ParseMediaCarrierDevice(s string) (registry.MediaCarrierDevice, error) {
	for _, d := range registry.MediaCarrierDeviceList {
		if string(d) == s {
			return d, nil
		}
	}
	return "", fmt.Errorf("unknown MediaCarrierDevice: %q", s)
}

func collectDtboFiles(cfg config.Configuration, selection map[registry.MediaCarrierDevice]string) []string {
	var dtboFiles []string

	for deviceName, optionName := range selection {
		if optionName == deviceOptionNone || optionName == "" {
			continue
		}

		for _, device := range registry.MediaCarrierRegistry.Devices {
			if device.Name != deviceName {
				continue
			}
			for _, opt := range device.Options {
				if opt.Name == optionName {
					if opt.DtboFile != "" {
						dtboFiles = append(dtboFiles, filepath.Join(cfg.OverlaysDir().String(), opt.DtboFile))
					}
					break
				}
			}
		}
	}

	return dtboFiles
}

func createWantedMarkers(cfg config.Configuration, selection map[registry.MediaCarrierDevice]string) {
	_ = deleteWanted(cfg)

	for deviceName, optionName := range selection {
		if optionName == "" || optionName == deviceOptionNone {
			continue // Skip creating status file for disabled devices
		}

		fileName := fmt.Sprintf("wanted_%s_%s", string(deviceName), optionName)
		markerPath := filepath.Join(cfg.StatusDir().String(), fileName)

		if err := touchFile(markerPath); err != nil {
			slog.Warn("Failed to create status file")
		}
	}
}

func touchFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	return f.Close()
}

func applyOverlays(cfg config.Configuration, ctx context.Context, dtboFiles []string) error {
	args := append([]string{"fdtoverlay", "-i", cfg.BaseDTB().String(), "-o", cfg.ActualDTB().String()}, dtboFiles...)

	proc, err := paths.NewProcess(nil, args...)
	if err != nil {
		return fmt.Errorf("failed to create process: %w", err)
	}
	stdout, stderr, err := proc.RunAndCaptureOutput(ctx)
	if err != nil {
		return fmt.Errorf("fdtoverlay failed: %w\n%s", err, stderr)
	}
	if len(stdout) > 0 {
		feedback.Print(string(stdout))
	}
	return nil
}

func deleteWanted(cfg config.Configuration) error {
	pattern := filepath.Join(cfg.StatusDir().String(), "wanted_*")

	files, err := filepath.Glob(pattern)
	if err != nil {
		slog.Warn("Fail to delete status files")
		return err
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			slog.Warn("Fail to delete status file")
		}
	}

	return nil
}
