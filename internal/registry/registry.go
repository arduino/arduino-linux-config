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
	Mounts []Mount
}

// Kind groups the mounts by the connector they use.
type Kind string

const (
	KindCarrier Kind = "carrier"
	KindHat     Kind = "hat"
)

// Mount names are unique over every kind, so the name alone selects a part.
func (r Registry) FindByName(name string) (Mount, bool) {
	for _, m := range r.Mounts {
		if string(m.Name) == name {
			return m, true
		}
	}
	return Mount{}, false
}

// Label is the word shown to the user for a kind.
func (k Kind) Label() string {
	if k == KindHat {
		return "addon"
	}
	return string(k)
}

// ByKind returns every mount when kind is empty.
func (r Registry) ByKind(kind Kind) []Mount {
	mounts := make([]Mount, 0, len(r.Mounts))
	for _, m := range r.Mounts {
		if kind == "" || m.Kind == kind {
			mounts = append(mounts, m)
		}
	}
	return mounts
}

func (r Registry) Has(kind Kind) bool {
	return len(r.ByKind(kind)) > 0
}

type DeviceName string

const (
	None    DeviceName = "none"
	Camera0 DeviceName = "camera0"
	Camera1 DeviceName = "camera1"
	Display DeviceName = "display"
)

type MountName string

const (
	MediaCarrier   MountName = "media-carrier"
	AudioCodecZero MountName = "audio-codec-zero"
	Automation     MountName = "automation"
	HighPrecision  MountName = "high-precision"
)

// Mount is a part that plugs into the board and adds device tree overlays.
// A carrier and a hat differ only by Kind and by the connector they use.
type Mount struct {
	Name          MountName
	Kind          Kind
	EnabledDtbos  []string
	DisabledDtbos []string
	Devices       []Device // empty for the hats available today
}

func (c Mount) FindDeviceByName(deviceName DeviceName) (Device, bool) {
	for _, d := range c.Devices {
		if d.Name == deviceName {
			return d, true
		}
	}
	return Device{}, false
}

// Device represents a configurable hardware device on a carrier
type Device struct {
	Name       DeviceName
	DeviceType DeviceType
	Options    []DeviceOption
}

// DeviceOption represents a configuration option for a device
type DeviceOption struct {
	Name             string
	DtboFiles        []string
	IncompatibleDtbo []string
}

func New() Registry {
	board := config.GetBoardID()
	boardOs := config.GetLinuxDistribution()

	switch {
	case board == "unoq":
		// unoq has no hat connector, so it declares no mount of kind hat.
		return Registry{
			Mounts: []Mount{unoqMediaCarrier},
		}
	case board == "ventunoq" && boardOs == "ubuntu":
		return Registry{
			Mounts: append([]Mount{ventunoqUbuntuMediaCarrier}, ventunoqUbuntuHats...),
		}
	default:
		return Registry{}
	}
}

var unoqMediaCarrier = Mount{
	Name: MediaCarrier,
	Kind: KindCarrier,
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

var ventunoqUbuntuHats = []Mount{
	{ // TODO update
		Name: AudioCodecZero,
		Kind: KindHat,
		EnabledDtbos: []string{
			"monaco-monza-automation-hat.dtbo",
		},
	},
	{
		Name: Automation,
		Kind: KindHat,
		EnabledDtbos: []string{
			"monaco-monza-automation-hat.dtbo",
		},
	},
	{ // TODO update
		Name: HighPrecision,
		Kind: KindHat,
		EnabledDtbos: []string{
			"monaco-monza-automation-hat.dtbo",
		},
	},
}

var ventunoqUbuntuMediaCarrier = Mount{
	Name: MediaCarrier,
	Kind: KindCarrier,
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
