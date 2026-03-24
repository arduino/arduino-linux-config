package registry

import (
	"fmt"
	"slices"
)

const DeviceOptionNone = "none"

// UnoQ specific media carrier
type MediaCarrierDeviceName string

const (
	Camera1 MediaCarrierDeviceName = "camera1"
	Camera2 MediaCarrierDeviceName = "camera2"
	Display MediaCarrierDeviceName = "display"
)

var MediaCarrierDeviceList = []MediaCarrierDeviceName{Camera1, Camera2, Display}

type MediaCarrier struct {
	Name    string
	Devices []Device
}

// Device represents a configurable hardware device on a carrier.
type Device struct {
	Name    MediaCarrierDeviceName
	Options []DeviceOption
}

type DeviceOption struct {
	Name     string
	DtboFile string
}

// TODO
// We should read the media-carrier from a configuration file
// media-carrier.json
// To collect all DTBOs files:
// 1. For each option we have:DtboFiles, CompatibleDtbo, IncompatibleDtbo
// 2. if CompatibleDtbo and IncompatibleDtbo intersectios is not null: ERROR, configuration now allowed
// 3. Add to the DTBOs collection all add DtboFiles and CompatibleDtbo, then remove IncompatibleDtbo
var MediaCarrierRegistry = MediaCarrier{
	Name: "media-carrier",
	Devices: []Device{
		{
			Name: Camera1,
			Options: []DeviceOption{
				{Name: "none", DtboFile: ""},
				{Name: "type1-2lane", DtboFile: "qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-2lanes.dtbo"},
				{Name: "type1-4lane", DtboFile: "qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-4lanes.dtbo"},
			},
		},
		{
			Name: Camera2,
			Options: []DeviceOption{
				{Name: "none", DtboFile: ""},
				{Name: "type1-2lane", DtboFile: "qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-2lanes.dtbo"},
				{Name: "type1-4lane", DtboFile: "qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-4lanes.dtbo"},
			},
		},
		{
			Name: Display,
			Options: []DeviceOption{
				{Name: "none", DtboFile: ""},
				{Name: "8-dsi-touch-a", DtboFile: "qrb2210-arduino-imola-carrier-media-panel-8in-touch-a-dsi.dtbo"},
			},
		},
	},
}

func GetMediaCarrierDeviceName(deviceName string) (MediaCarrierDeviceName, bool) {
	device := MediaCarrierDeviceName(deviceName)

	if slices.Contains(MediaCarrierDeviceList, device) {
		return device, true
	}

	return "", false
}

func IsOptionValid(deviceName MediaCarrierDeviceName, optionName string) bool {
	for _, device := range MediaCarrierRegistry.Devices {
		if device.Name != deviceName {
			continue
		}
		for _, option := range device.Options {
			if optionName == option.Name {
				return true
			}
		}
	}
	return false
}

func ValidateInput(rawDevice string, rawOption string) (MediaCarrierDeviceName, error) {
	device, found := GetMediaCarrierDeviceName(rawDevice)
	if !found {
		return "", fmt.Errorf("unknown device: %q", rawDevice)
	}

	if !IsOptionValid(device, rawOption) {
		return "", fmt.Errorf("device %q does not support option %q", device, rawOption)
	}

	return device, nil
}
