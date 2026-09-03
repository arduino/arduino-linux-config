// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package registry

import (
	"github.com/arduino/arduino-linux-config/internal/config"
)

type DeviceType string

const (
	DeviceTypeCamera  DeviceType = "camera"
	DeviceTypeDisplay DeviceType = "display"
)

type Registry struct {
	Carriers []Carrier
	Addons   []Addon
}

func (r Registry) FindByName(carrier string) (Carrier, bool) {
	for _, c := range r.Carriers {
		if string(c.Name) == carrier {
			return c, true
		}
	}
	return Carrier{}, false
}

func (r Registry) FindAddonByName(addon string) (Addon, bool) {
	for _, a := range r.Addons {
		if string(a.Name) == addon {
			return a, true
		}
	}
	return Addon{}, false
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

// Start addon section
type Addon struct {
	Name         AddonName
	EnabledDtbos []string
}

type AddonName string

const (
	AudioCodecZero AddonName = "audio-codec-zero"
	Automation     AddonName = "automation"
	HighPrecision  AddonName = "high-precision"
)

func New() Registry {
	board := config.GetBoardID()
	boardOs := config.GetLinuxDistribution()

	switch {
	case board == "unoq":
		return Registry{
			Carriers: []Carrier{unoqMediaCarrier},
		}
	case board == "ventunoq" && boardOs == "ubuntu":
		return Registry{
			Carriers: []Carrier{ventunoqUbuntuMediaCarrier},
			Addons:   ventunoqUbuntuAddons,
		}
	default:
		return Registry{}
	}
}

var unoqMediaCarrier = Carrier{
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
				},
				{
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
}

var ventunoqUbuntuAddons = []Addon{
	{ // TODO update
		Name: AudioCodecZero,
		EnabledDtbos: []string{
			"monaco-addons-iqaudio-codeczero-monza.dtbo",
		},
	},
	{
		Name: Automation,
		EnabledDtbos: []string{
			"monaco-monza-automation-hat.dtbo",
		},
	},
}

var ventunoqUbuntuMediaCarrier = Carrier{
	Name: MediaCarrier,
	Devices: []Device{
		{
			Name:       "display",
			DeviceType: DeviceTypeDisplay,
			Options: []DeviceOption{
				{
					Name:      "none",
					DtboFiles: []string{},
				},
				{
					Name: "8-dsi-touch-a",
					DtboFiles: []string{
						"monaco-monza-dsi-waveshare,8.0-dsi-touch-a.dtbo",
					},
				},
			},
		},
	},
}
