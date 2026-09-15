// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package registry

import (
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/arduino/arduino-linux-config/internal/config"
)

type DtboSupport struct {
	Dtbo           string
	MinRequirement string
	IsSupported    func(current string) bool
}

// SupportMatrix defines minimum software/kernel version requirements for dtbo files.
type SupportMatrix struct {
	Support []DtboSupport
}

func (sm SupportMatrix) getSupport(dtbo string) (DtboSupport, bool) {
	for _, s := range sm.Support {
		if s.Dtbo == dtbo {
			return s, true
		}
	}
	return DtboSupport{}, false
}

// Factory builds a Registry based on target board specifications and configuration.
type Factory struct {
	board         string
	boardOS       string
	dtboSuppport  string
	supportMatrix SupportMatrix
}

func NewFactory(supportMatrix SupportMatrix) *Factory {
	return &Factory{
		board:         config.GetBoardID(),
		boardOS:       config.GetLinuxDistribution(),
		dtboSuppport:  config.GetDtboSupportVersion(),
		supportMatrix: supportMatrix,
	}
}

// if the overlay is not declared or its function is missing is unsupported
func (f *Factory) isDtboSupported(dtbo string) bool {
	support, found := f.supportMatrix.getSupport(dtbo)
	if !found || support.IsSupported == nil {
		return false
	}
	return support.IsSupported(f.dtboSuppport)
}

func (f *Factory) areDtbosSupported(dtbos []string) bool {
	for _, dtbo := range dtbos {
		if !f.isDtboSupported(dtbo) {
			return false
		}
	}
	return true
}

func (f *Factory) isDeviceSupported(d Device) bool {
	for _, opt := range d.Options {
		if !f.areDtbosSupported(opt.DtboFiles) {
			return false
		}
	}
	return true
}

// A mount is supported when its enabled DTBOs are all supported;
// otherwise, a mount without mount-level DTBOs is considered supported
// if at least one device DTBO is supported.
func (f *Factory) isMountSupported(m Mount) bool {
	if len(m.EnabledDtbos) > 0 {
		return f.areDtbosSupported(m.EnabledDtbos)
	}
	if len(m.Devices) == 0 {
		return false
	}
	for _, d := range m.Devices {
		if f.isDeviceSupported(d) {
			return true
		}
	}
	return false
}

func (f *Factory) populateMountOsSupport(m Mount) Mount {
	// device level support
	populatedDevices := make([]Device, len(m.Devices))
	for i, d := range m.Devices {
		d.OsSupport = f.isDeviceSupported(d)
		populatedDevices[i] = d
	}
	m.Devices = populatedDevices

	// mount level support
	m.OsSupport = f.isMountSupported(m)
	return m
}

func (f *Factory) Create() Registry {
	var mounts []Mount
	switch {
	case f.board == "unoq":
		mounts = []Mount{unoqMediaCarrier}
	case f.board == "ventunoq" && f.boardOS == "ubuntu":
		mounts = append([]Mount{ventunoqMediaCarrier}, ventunoqUbuntuHats...)
	default:
		return Registry{}
	}

	resMounts := make([]Mount, len(mounts))
	for i, m := range mounts {
		resMounts[i] = f.populateMountOsSupport(m)
	}

	return Registry{
		Mounts: resMounts,
	}
}

// semver treats "1078-qcom" as an alphanumeric identifier and compares it lexically.
// normalizeKernelVersion turns Ubuntu/Debian style kernel versions and compare it numerically
// dpkg --compare-versions "6.8.0-999-qcom" gt "6.8.0-1000-qcom" && echo "First version is newer"
func normalizeKernelVersion(v string) string {
	idx := strings.Index(v, "-")
	if idx == -1 {
		return v
	}
	return v[:idx+1] + strings.Replace(v[idx+1:], "-", ".", 1)
}

func isVersionEqual(current, expected string) bool {
	if expected == "" || current == expected {
		return true
	}
	if current == "" {
		return false
	}

	vCurrent, err1 := semver.NewVersion(normalizeKernelVersion(current))
	vExpected, err2 := semver.NewVersion(normalizeKernelVersion(expected))
	if err1 == nil && err2 == nil {
		return vCurrent.Equal(vExpected)
	}

	return current == expected
}

func isVersionAtLeast(current, minReq string) bool {
	if minReq == "" || current == minReq {
		return true
	}
	if current == "" {
		return false
	}

	vCurrent, err1 := semver.NewVersion(normalizeKernelVersion(current))
	vMin, err2 := semver.NewVersion(normalizeKernelVersion(minReq))
	if err1 == nil && err2 == nil {
		return vCurrent.Compare(vMin) >= 0
	}

	return current >= minReq
}

func NewSupportMatrix() SupportMatrix {
	const unoQKernelVersion = "7.0.0-g122c2c22d838"
	const ventunoQUbuntuKernelVersion = "6.8.0-1078-qcom"

	return SupportMatrix{
		Support: []DtboSupport{
			{
				Dtbo:           "qrb2210-arduino-imola-carrier-media.dtbo",
				MinRequirement: unoQKernelVersion,
				IsSupported: func(current string) bool {
					return isVersionEqual(current, unoQKernelVersion)
				},
			},
			{
				Dtbo:           "qrb2210-arduino-imola-video_sound-usbc.dtbo",
				MinRequirement: unoQKernelVersion,
				IsSupported: func(current string) bool {
					return isVersionEqual(current, unoQKernelVersion)
				},
			},
			{
				Dtbo:           "qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-2lanes.dtbo",
				MinRequirement: unoQKernelVersion,
				IsSupported: func(current string) bool {
					return isVersionEqual(current, unoQKernelVersion)
				},
			},
			{
				Dtbo:           "qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-4lanes.dtbo",
				MinRequirement: unoQKernelVersion,
				IsSupported: func(current string) bool {
					return isVersionEqual(current, unoQKernelVersion)
				},
			},
			{
				Dtbo:           "qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-2lanes.dtbo",
				MinRequirement: unoQKernelVersion,
				IsSupported: func(current string) bool {
					return isVersionEqual(current, unoQKernelVersion)
				},
			},
			{
				Dtbo:           "qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-4lanes.dtbo",
				MinRequirement: unoQKernelVersion,
				IsSupported: func(current string) bool {
					return isVersionEqual(current, unoQKernelVersion)
				},
			},
			{
				Dtbo:           "qrb2210-arduino-imola-carrier-media-panel-5in_touch_a-dsi.dtbo",
				MinRequirement: unoQKernelVersion,
				IsSupported: func(current string) bool {
					return isVersionEqual(current, unoQKernelVersion)
				},
			},
			{
				Dtbo:           "qrb2210-arduino-imola-carrier-media-panel-8in_touch_a-dsi.dtbo",
				MinRequirement: unoQKernelVersion,
				IsSupported: func(current string) bool {
					return isVersionEqual(current, unoQKernelVersion)
				},
			},
			{
				Dtbo:           "qrb2210-arduino-imola-carrier-media-panel-10in_touch_a-dsi.dtbo",
				MinRequirement: unoQKernelVersion,
				IsSupported: func(current string) bool {
					return isVersionEqual(current, unoQKernelVersion)
				},
			},
			{
				Dtbo:           "monaco-addons-iqaudio-codeczero-monza.dtbo",
				MinRequirement: ventunoQUbuntuKernelVersion,
				IsSupported: func(current string) bool {
					return isVersionAtLeast(current, "6.8.0-1084-qcom")
				},
			},
			{
				Dtbo:           "monaco-monza-automation-hat.dtbo",
				MinRequirement: ventunoQUbuntuKernelVersion,
				IsSupported: func(current string) bool {
					return isVersionAtLeast(current, "6.8.0-1080-qcom")
				},
			},
			{
				Dtbo:           "monaco-monza-dsi-waveshare,8.0-dsi-touch-a.dtbo",
				MinRequirement: ventunoQUbuntuKernelVersion,
				IsSupported: func(current string) bool {
					return isVersionAtLeast(current, "6.8.0-1084-qcom")
				},
			},
		},
	}
}
