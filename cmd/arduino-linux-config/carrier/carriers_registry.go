package carrier

type DeviceOption struct {
	Name     string
	DtboFile string
}

// Device represents a configurable hardware device on a carrier.
type Device struct {
	Name    string
	Options []DeviceOption
}

// Carrier represents a hardware carrier board with its configurable devices.
type Carrier struct {
	Name    string
	Devices []Device
}

// getCarrierRegistry returns all carriers available on this hardware.
// TODO: migrate to a config file loader when needed.
func getCarrierRegistry() []Carrier {
	return []Carrier{
		mediaCarrier(),
		builtinCarrier(),
	}
}

func getMediaCarrierRegistry() Carrier {
	return mediaCarrier()
}

func getBuiltinCarrierRegistry() Carrier {
	return builtinCarrier()
}

func mediaCarrier() Carrier {
	return Carrier{
		Name: "media-carrier",
		Devices: []Device{
			{
				Name: "camera1",
				Options: []DeviceOption{
					{Name: "none", DtboFile: ""},
					{Name: "type1-2lane", DtboFile: "qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-2lanes.dtbo"},
					{Name: "type1-4lane", DtboFile: "qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-4lanes.dtbo"},
				},
			},
			{
				Name: "camera2",
				Options: []DeviceOption{
					{Name: "none", DtboFile: ""},
					{Name: "type1-2lane", DtboFile: "qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-2lanes.dtbo"},
					{Name: "type1-4lane", DtboFile: "qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-4lanes.dtbo"},
				},
			},
			{
				Name: "display1",
				Options: []DeviceOption{
					{Name: "none", DtboFile: ""},
					{Name: "8-dsi-touch-a", DtboFile: "qrb2210-arduino-imola-carrier-media-panel-8in-touch-a-dsi.dtbo"},
				},
			},
		},
	}
}

func builtinCarrier() Carrier {
	return Carrier{
		Name: "builtin",
		Devices: []Device{
			{
				Name: "camera1",
				Options: []DeviceOption{
					{Name: "none", DtboFile: ""},
					{Name: "type1-2lane", DtboFile: "qrb2210-arduino-monza-camera-imx219-csi0-2lanes.dtbo"},
					{Name: "type1-4lane", DtboFile: "qrb2210-arduino-monza-camera-imx219-csi0-4lanes.dtbo"},
				},
			},
			{
				Name: "camera2",
				Options: []DeviceOption{
					{Name: "none", DtboFile: ""},
					{Name: "type1-2lane", DtboFile: "qrb2210-arduino-monza-camera-imx219-csi1-2lanes.dtbo"},
					{Name: "type1-4lane", DtboFile: "qrb2210-arduino-monza-camera-imx219-csi1-4lanes.dtbo"},
				},
			},
			{
				Name: "camera3",
				Options: []DeviceOption{
					{Name: "none", DtboFile: ""},
					{Name: "type1-2lane", DtboFile: "qrb2210-arduino-monza-camera-imx219-csi2-2lanes.dtbo"},
					{Name: "type1-4lane", DtboFile: "qrb2210-arduino-monza-camera-imx219-csi2-4lanes.dtbo"},
				},
			},
			{
				Name: "display1",
				Options: []DeviceOption{
					{Name: "none", DtboFile: ""},
					{Name: "8-dsi-touch-a", DtboFile: "qrb2210-arduino-monza-panel-8in-touch-a-dsi.dtbo"},
				},
			},
		},
	}
}
