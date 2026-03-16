package registry

import (
	"time"

	"github.com/arduino/go-paths-helper"
)

// UnoQ specific media carrier
type MediaCarrierDevice string

const (
	Leds    MediaCarrierDevice = "leds"
	Camera1 MediaCarrierDevice = "camera1"
	Camera2 MediaCarrierDevice = "camera2"
	Display MediaCarrierDevice = "display"
)

var MediaCarrierDeviceList = []MediaCarrierDevice{Leds, Camera1, Camera2, Display}

type MediaCarrier struct {
	Name    string
	Devices []Device
}

// Device represents a configurable hardware device on a carrier.
type Device struct {
	Name    MediaCarrierDevice
	Options []DeviceOption
}

type DeviceOption struct {
	Name     string
	DtboFile string
}

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

type OverlayFile struct {
	Device    string // "camera1"
	Option    string // "type1-2lane"
	CreatedAt time.Time
	Path      *paths.Path
}
