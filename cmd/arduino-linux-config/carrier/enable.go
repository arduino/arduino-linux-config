package carrier

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

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

	disableHandler(cfg, ctx, carrierName)
	wantedDtboFiles := collectDtboFiles(cfg, wantedDevicesList)

	if len(wantedDtboFiles) > 0 {
		if err := applyOverlays(cfg, ctx, wantedDtboFiles); err != nil {
			feedback.Fatal(err.Error(), feedback.ErrGeneric)
		}
	}

	fileStatusList, err := loadStateMarkers(cfg)
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}

	for device, optionValue := range wantedDevicesList {
		overlayFileName := fmt.Sprintf("%s-%s", device, optionValue)

		// Pulisci SEMPRE i file post-boot per questo device,
		// indipendentemente dall'optionValue
		if err := pruneOldPostBootStateMarker(string(device), fileStatusList); err != nil {
			feedback.Fatal(err.Error(), feedback.ErrGeneric)
		}

		// Poi, solo se non è "none", crea il nuovo file
		if optionValue != "none" {
			if err := createStateMarker(overlayFileName, cfg); err != nil {
				feedback.Fatal(err.Error(), feedback.ErrGeneric)
			}
		}
	}

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

func loadStateMarkers(cfg config.Configuration) ([]registry.OverlayFile, error) {
	entries, err := cfg.StatusDir().ReadDir()
	if err != nil {
		return nil, fmt.Errorf("failed to read overlays dir: %w", err)
	}

	const layout = "20060102-150405"
	files := make([]registry.OverlayFile, 0, len(entries))

	for _, entry := range entries {
		name := entry.Base() // e.g. "camera1_20260313-150945.dtbo"

		// Rimuovi estensione
		nameNoExt := strings.TrimSuffix(name, filepath.Ext(name))

		parts := strings.SplitN(nameNoExt, "_", 2)
		if len(parts) != 2 {
			continue
		}

		// Splitta device e option sul primo "-"
		deviceParts := strings.SplitN(parts[0], "-", 2)
		if len(deviceParts) != 2 {
			continue
		}

		device := deviceParts[0] // "camera1"
		option := deviceParts[1] // "type1-2lane"

		createdAt, err := time.Parse(layout, parts[1])
		if err != nil {
			continue
		}

		files = append(files, registry.OverlayFile{
			Device:    device,
			Option:    option,
			CreatedAt: createdAt,
			Path:      entry,
		})
	}

	return files, nil
}

func createStateMarker(deviceName string, cfg config.Configuration) error {
	const layout = "20060102-150405"
	timestamp := time.Now().UTC().Format(layout)
	filename := fmt.Sprintf("%s_%s.dtbo", deviceName, timestamp)
	return cfg.StatusDir().Join(filename).WriteFile([]byte{})
}

// pruneOldPostBootStateMarker removes all overlay files for the specified device that were created after the last boot time.
func pruneOldPostBootStateMarker(deviceName string, files []registry.OverlayFile) error {
	bootTime, err := getBootTime()
	if err != nil {
		return fmt.Errorf("failed to get boot time: %w", err)
	}
	// Filtra solo i file del device
	var deviceFiles []registry.OverlayFile
	for _, f := range files {
		if f.Device == deviceName {
			deviceFiles = append(deviceFiles, f)
		}
	}

	// Cancella i file post-boot (verranno rimpiazzati dal nuovo)
	var surviving []registry.OverlayFile
	for _, f := range deviceFiles {
		if f.CreatedAt.After(bootTime) {
			if err := f.Path.Remove(); err != nil {
				return fmt.Errorf("failed to remove post-boot overlay file %s: %w", f.Path, err)
			}
		} else {
			surviving = append(surviving, f)
		}
	}

	// Tra i file pre-boot, tieni solo il più recente (ne basta 1 pre-boot + 1 nuovo post-boot = 2 totali)
	if len(surviving) > 1 {
		// Ordina dal più recente al più vecchio
		sort.Slice(surviving, func(i, j int) bool {
			return surviving[i].CreatedAt.After(surviving[j].CreatedAt)
		})
		// Cancella tutti i pre-boot tranne il più recente
		for _, f := range surviving[1:] {
			if err := f.Path.Remove(); err != nil {
				return fmt.Errorf("failed to remove old overlay file %s: %w", f.Path, err)
			}
		}
	}

	return nil
}

func getBootTime() (time.Time, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read /proc/uptime: %w", err)
	}

	// /proc/uptime format: "12345.67 23456.78"
	// il primo valore è i secondi di uptime
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return time.Time{}, fmt.Errorf("unexpected /proc/uptime format")
	}

	uptimeSeconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse uptime: %w", err)
	}

	bootTime := time.Now().UTC().Add(-time.Duration(uptimeSeconds * float64(time.Second)))
	return bootTime, nil
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

func ParseMediaCarrierDevice(configuredDeviceName string) (registry.MediaCarrierDevice, error) {
	for _, deviceName := range registry.MediaCarrierDeviceList {
		if string(deviceName) == configuredDeviceName {
			return deviceName, nil
		}
	}
	return "", fmt.Errorf("unknown MediaCarrierDevice: %q", configuredDeviceName)
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
