package registry

import (
	"time"

	"github.com/arduino/go-paths-helper"
)

const DeviceOptionNone = "none"

// UnoQ specific media carrier
type MediaCarrierDeviceName string

const (
	Leds    MediaCarrierDeviceName = "leds"
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

// TODO Read from files
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

type StatusFile struct {
	DeviceName string
	Option     string
	CreatedAt  time.Time
	Path       *paths.Path
}
