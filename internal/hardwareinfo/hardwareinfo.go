package hardwareinfo

import (
	"sync"
)

type HardwareInfo struct {
	Carrier        Carrier
	BuiltInCarrier Carrier // Monza only
}

type Carrier struct {
	Name     string
	Overlays []Overlay
}

type Overlay struct {
	DeviceName   string
	HardwareData string
	FileName     string
}

var (
	instance *HardwareInfo
	once     sync.Once
)

func GetAvailableDeviceList() (*HardwareInfo, error) {
	once.Do(func() {
		instance = &HardwareInfo{}
		carrier, _ := initCarrier()
		instance.Carrier = carrier
	})
	return instance, nil
}

func initCarrier() (Carrier, error) {
	return Carrier{
		Name: "media-carrier",
		Overlays: []Overlay{
			{
				DeviceName:   "leds",
				HardwareData: "carrier-leds",
				FileName:     "qrb2210-arduino-imola-carrier-media.dtbo",
			},
			{
				DeviceName:   "camera1",
				HardwareData: "imx219-csi0-2lanes",
				FileName:     "qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-2lanes.dtbo",
			},
			{
				DeviceName:   "camera1",
				HardwareData: "imx219-csi0-4lanes",
				FileName:     "qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-4lanes.dtbo",
			},
			{
				DeviceName:   "camera2",
				HardwareData: "imx219-csi1-2lanes",
				FileName:     "qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-2lanes.dtbo",
			},
			{
				DeviceName:   "camera2",
				HardwareData: "imx219-csi1-4lanes",
				FileName:     "qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-4lanes.dtbo",
			},
			{
				DeviceName:   "display1",
				HardwareData: "8in-touch-a-dsi",
				FileName:     "qrb2210-arduino-imola-carrier-media-panel-8in-touch-a-dsi.dtbo",
			},
		},
	}, nil
}
