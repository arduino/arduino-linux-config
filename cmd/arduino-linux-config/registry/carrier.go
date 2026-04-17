// This file is part of arduino-linux-config.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-linux-config.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package registry

type DeviceType string

const (
	DeviceTypeCamera  DeviceType = "camera"
	DeviceTypeDisplay DeviceType = "display"
)

type CarrierRegistry struct {
	Carriers []Carrier
}

func (r CarrierRegistry) FindByName(carrier string) (Carrier, bool) {
	for _, c := range r.Carriers {
		if string(c.Name) == carrier {
			return c, true
		}
	}
	return Carrier{}, false
}

type CarrierDeviceName string

const (
	None    CarrierDeviceName = "none"
	Camera0 CarrierDeviceName = "camera0"
	Camera1 CarrierDeviceName = "camera1"
	Display CarrierDeviceName = "display"
)

type CarrierName string

const (
	MediaCarrier CarrierName = "media-carrier"
)

type Carrier struct {
	Name         CarrierName
	EnabledDtbo  []string
	DisabledDtbo []string
	Devices      []Device
}

func (c Carrier) FindDeviceByName(deviceName CarrierDeviceName) (Device, bool) {
	for _, d := range c.Devices {
		if d.Name == deviceName {
			return d, true
		}
	}
	return Device{}, false
}

// Device represents a configurable hardware device on a carrier
type Device struct {
	Name       CarrierDeviceName
	DeviceType DeviceType
	Options    []DeviceOption
}

// DeviceOption represents a configuration option for a device
type DeviceOption struct {
	Name             string
	DtboFiles        []string
	IncompatibleDtbo []string
}

var Registry = CarrierRegistry{
	Carriers: []Carrier{{
		Name: MediaCarrier,
		EnabledDtbo: []string{
			"qrb2210-arduino-imola-carrier-media.dtbo",
			"qrb2210-arduino-imola-video_sound-usbc.dtbo",
		},
		DisabledDtbo: []string{
			"qrb2210-arduino-imola-video_sound-usbc.dtbo",
		},
		Devices: []Device{
			{
				Name:       "camera0",
				DeviceType: DeviceTypeCamera,
				Options: []DeviceOption{
					{
						Name:      "none",
						DtboFiles: []string{"qrb2210-arduino-imola-video_sound-usbc.dtbo"},
					},
					{
						Name: "type1-2lane",
						DtboFiles: []string{
							"qrb2210-arduino-imola-video_sound-usbc.dtbo",
							"qrb2210-arduino-imola-carrier-media.dtbo",
							"qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-2lanes.dtbo",
						},
					},
					{
						Name: "type1-4lane",
						DtboFiles: []string{
							"qrb2210-arduino-imola-video_sound-usbc.dtbo",
							"qrb2210-arduino-imola-carrier-media.dtbo",
							"qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-4lanes.dtbo",
						},
					},
				},
			},
			{
				Name:       "camera1",
				DeviceType: DeviceTypeCamera,
				Options: []DeviceOption{
					{
						Name:      "none",
						DtboFiles: []string{"qrb2210-arduino-imola-video_sound-usbc.dtbo"},
					},
					{
						Name: "type1-2lane",
						DtboFiles: []string{
							"qrb2210-arduino-imola-video_sound-usbc.dtbo",
							"qrb2210-arduino-imola-carrier-media.dtbo",
							"qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-2lanes.dtbo",
						},
					}, {
						Name: "type1-4lane",
						DtboFiles: []string{
							"qrb2210-arduino-imola-video_sound-usbc.dtbo",
							"qrb2210-arduino-imola-carrier-media.dtbo",
							"qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-4lanes.dtbo",
						},
					},
				},
			},
			{
				Name:       "display",
				DeviceType: DeviceTypeDisplay,
				Options: []DeviceOption{
					{
						Name:      "none",
						DtboFiles: []string{"qrb2210-arduino-imola-video_sound-usbc.dtbo"},
					},
					{
						Name: "8-dsi-touch-a",
						DtboFiles: []string{
							"qrb2210-arduino-imola-carrier-media.dtbo",
							"qrb2210-arduino-imola-carrier-media-panel-8in_touch_a-dsi.dtbo",
						},
						IncompatibleDtbo: []string{
							"qrb2210-arduino-imola-video_sound-usbc.dtbo",
						},
					},
				},
			},
		},
	}},
}
