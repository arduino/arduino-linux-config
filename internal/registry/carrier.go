// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

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
	Name          CarrierName
	EnabledDtbos  []string
	DisabledDtbos []string
	Devices       []Device
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

func New() CarrierRegistry {
	return CarrierRegistry{
		Carriers: []Carrier{{
			Name: MediaCarrier,
			EnabledDtbos: []string{
				"qrb2210-arduino-imola-carrier-media.dtbo",
				"qrb2210-arduino-imola-video_sound-usbc.dtbo",
			},
			DisabledDtbos: []string{
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
							Name: "type1-2lanes",
							DtboFiles: []string{
								"qrb2210-arduino-imola-video_sound-usbc.dtbo",
								"qrb2210-arduino-imola-carrier-media.dtbo",
								"qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-2lanes.dtbo",
							},
						},
						{
							Name: "type1-4lanes",
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
							Name: "type1-2lanes",
							DtboFiles: []string{
								"qrb2210-arduino-imola-video_sound-usbc.dtbo",
								"qrb2210-arduino-imola-carrier-media.dtbo",
								"qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-2lanes.dtbo",
							},
						}, {
							Name: "type1-4lanes",
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
							Name: "5-dsi-touch-a",
							DtboFiles: []string{
								"qrb2210-arduino-imola-carrier-media.dtbo",
								"qrb2210-arduino-imola-carrier-media-panel-5in_touch_a-dsi.dtbo",
							},
							IncompatibleDtbo: []string{
								"qrb2210-arduino-imola-video_sound-usbc.dtbo",
							},
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
						{
							Name: "10-dsi-touch-a",
							DtboFiles: []string{
								"qrb2210-arduino-imola-carrier-media.dtbo",
								"qrb2210-arduino-imola-carrier-media-panel-10in_touch_a-dsi.dtbo",
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
}
