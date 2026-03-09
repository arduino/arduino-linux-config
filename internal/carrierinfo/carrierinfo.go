package carrierinfo

import (
	"sync"
)

type BoardCarriers struct {
	MediaCarrier   Carrier
	BuiltInCarrier Carrier // Monza only
}

// UnoQ specific media carrier
type MediaCarrierDevice string

const (
	Leds     MediaCarrierDevice = "leds"
	Camera1  MediaCarrierDevice = "camera1"
	Camera2  MediaCarrierDevice = "camera2"
	Display1 MediaCarrierDevice = "display1"
)

var MediaCarrierDeviceList = []MediaCarrierDevice{Leds, Camera1, Camera2, Display1}

// Overlay relate structures
type Carrier struct {
	Name     string
	Overlays []Overlay
}

type Overlay struct {
	DeviceName   MediaCarrierDevice
	HardwareData string
	FileName     string
}

var (
	instance *BoardCarriers
	once     sync.Once
)

func GetAvailableDeviceList() *BoardCarriers {
	once.Do(func() {
		instance = &BoardCarriers{}
		carrier := initCarrier()
		instance.MediaCarrier = carrier
	})
	return instance
}

func initCarrier() Carrier {
	return Carrier{
		Name: "media-carrier",
		Overlays: []Overlay{
			{
				DeviceName:   Leds,
				HardwareData: "carrier-leds",
				FileName:     "qrb2210-arduino-imola-carrier-media.dtbo",
			},
			{
				DeviceName:   Camera1,
				HardwareData: "imx219-csi0-2lanes",
				FileName:     "qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-2lanes.dtbo",
			},
			{
				DeviceName:   Camera1,
				HardwareData: "imx219-csi0-4lanes",
				FileName:     "qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-4lanes.dtbo",
			},
			{
				DeviceName:   Camera2,
				HardwareData: "imx219-csi1-2lanes",
				FileName:     "qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-2lanes.dtbo",
			},
			{
				DeviceName:   Camera2,
				HardwareData: "imx219-csi1-4lanes",
				FileName:     "qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-4lanes.dtbo",
			},
			{
				DeviceName:   Display1,
				HardwareData: "8in-touch-a-dsi",
				FileName:     "qrb2210-arduino-imola-carrier-media-panel-8in-touch-a-dsi.dtbo",
			},
		},
	}
}
