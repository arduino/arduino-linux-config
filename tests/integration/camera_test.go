// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Stands in for the real, hand-patched monza base: just enough labeled nodes
// (cci1, cci1_i2c0, camss) for the generated camera overlay to resolve
// against, compiled with -@ so fdtoverlay can see their symbols.
const cameraBaseFixtureDts = `/dts-v1/;
/ {
	compatible = "arduino,monza";

	cci1: cci@ac14000 {
		cci1_i2c0: i2c-bus@0 {
		};
	};

	camss: isp@ac78000 {
	};
};
`

// Seeds combined-dtb-base.dtb on the dtb_a partition: on real Debian hardware
// this is a one-time manual step (see BaseDtbFileName in dto.VentunoQ), since
// the kernel package doesn't yet ship a camera-capable monza dtb.
func seedVentunoqDebianCameraBase(t *testing.T) {
	t.Helper()
	inDtbPartition(t, strings.Join([]string{
		"cat <<'EOF' > /tmp/camera-base.dts",
		cameraBaseFixtureDts,
		"EOF",
		"dtc -@ -I dts -O dtb -o /mnt/dtb/combined-dtb-base.dtb /tmp/camera-base.dts",
	}, "\n"))
}

// Camera overlays are built on the fly, so enabling one must both compile and
// apply correctly: the generated combined-dtb.dtb should contain the sensor,
// wired to the requested port.
func TestCameraGeneratesDeviceTree(t *testing.T) {
	startVentunoqDebianPrivilegedDockerContainer(t)
	t.Cleanup(func() { stopVentunoqDockerContainer(t) })
	seedVentunoqDebianCameraBase(t)

	execInVentunoqContainer(t, "arduino-linux-config", "hw", "enable", "cameras", "camera1=imx219-2lanes")

	out := inDtbPartition(t, "dtc -I dtb -O dts /mnt/dtb/"+ventunoqGeneratedDtb)
	require.Contains(t, out, `compatible = "sony,imx219"`, "sensor node must be merged into the device tree")
	require.Contains(t, out, "port@1 {", "the overlay must target csiphy1, not csiphy0 or csiphy2")
}
