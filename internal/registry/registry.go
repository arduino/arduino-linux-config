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

type Registry struct {
	Mounts []Mount
}

// Mount names are unique over every kind, so the name alone selects a part.
func (r Registry) FindByName(name string) (Mount, bool) {
	for _, m := range r.Mounts {
		if string(m.Name) == name {
			return m, true
		}
	}
	return Mount{}, false
}

// ByKind returns every mount when kind is empty.
func (r Registry) ByKind(kind Kind) Registry {
	mounts := make([]Mount, 0, len(r.Mounts))
	for _, m := range r.Mounts {
		if kind == "" || m.Kind == kind {
			mounts = append(mounts, m)
		}
	}
	return Registry{Mounts: mounts}
}

// Supported drops the mounts whose OsSupport is false, and within the
// remaining mounts, the devices whose OsSupport is false.
func (r Registry) Supported() Registry {
	mounts := make([]Mount, 0, len(r.Mounts))
	for _, m := range r.Mounts {
		if !m.OsSupport {
			continue
		}
		devices := make([]Device, 0, len(m.Devices))
		for _, d := range m.Devices {
			if d.OsSupport {
				devices = append(devices, d)
			}
		}
		m.Devices = devices
		mounts = append(mounts, m)
	}
	return Registry{Mounts: mounts}
}

// Kind groups the mounts by the connector they use.
type Kind string

const (
	KindCarrier Kind = "carrier"
	KindHat     Kind = "hat"
)

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
)

// Mount is a part that plugs into the board and adds device tree overlays.
// A carrier and a hat differ only by Kind and by the connector they use.
type Mount struct {
	Name          MountName
	Kind          Kind
	EnabledDtbos  []string
	DisabledDtbos []string
	Devices       []Device // empty for the hats available today
	OsSupport     bool
}

func (c Mount) FindDeviceByName(deviceName DeviceName) (Device, bool) {
	for _, d := range c.Devices {
		if d.Name == deviceName {
			return d, true
		}
	}
	return Device{}, false
}

// Device represents a configurable hardware device on a mount
type Device struct {
	Name       DeviceName
	DeviceType DeviceType
	Options    []DeviceOption
	OsSupport  bool
}

// DeviceOption represents a configuration option for a device
type DeviceOption struct {
	Name             string
	DtboFiles        []string
	IncompatibleDtbo []string
}

func New() Registry {
	return NewFactory(NewSupportMatrix()).Create()
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
	{
		Name: AudioCodecZero,
		Kind: KindHat,
		EnabledDtbos: []string{
			"monaco-addons-iqaudio-codeczero-monza.dtbo",
		},
	},
	{
		Name: Automation,
		Kind: KindHat,
		EnabledDtbos: []string{
			"monaco-monza-automation-hat.dtbo",
		},
	},
}

var ventunoqMediaCarrier = Mount{
	Name:          MediaCarrier,
	Kind:          KindCarrier,
	EnabledDtbos:  []string{},
	DisabledDtbos: []string{},
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
