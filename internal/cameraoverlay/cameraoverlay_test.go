// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package cameraoverlay

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-linux-config/internal/executor"
)

// totalLanes sums every sensor's supported lane count, for sizing expectations
// that shouldn't need updating every time the catalog grows.
func totalLanes() int {
	total := 0
	for _, sensor := range Sensors {
		total += len(sensor.Lanes)
	}
	return total
}

func TestOptionsForPort(t *testing.T) {
	options := OptionsForPort(Port{Index: 1, ResetGPIO: 75})
	require.Len(t, options, totalLanes())
	require.Contains(t, options, Option{Name: "imx219-2lanes", DtboFile: "monaco-monza-camera-csi1-imx219-2lane.dtbo"})
	require.Contains(t, options, Option{Name: "imx219-4lanes", DtboFile: "monaco-monza-camera-csi1-imx219-4lane.dtbo"})
	require.Contains(t, options, Option{Name: "ov5647-2lanes", DtboFile: "monaco-monza-camera-csi1-ov5647-2lane.dtbo"})
}

func TestFilenamesCoverEveryPort(t *testing.T) {
	filenames := Filenames()
	require.Len(t, filenames, len(Ports)*totalLanes())
	require.Contains(t, filenames, "monaco-monza-camera-csi0-imx219-2lane.dtbo")
	require.Contains(t, filenames, "monaco-monza-camera-csi2-imx219-4lane.dtbo")
}

func TestRenderDts(t *testing.T) {
	got := renderDts(spec{
		port:   Port{Index: 0, ResetGPIO: 64},
		sensor: Sensor{Name: "imx219", Compatible: "sony,imx219", I2CAddress: 0x10, LinkFrequencyHz: 456000000},
		lanes:  2,
	})

	require.Contains(t, got, "&cci0 {")
	require.Contains(t, got, "&cci0_i2c0 {")
	require.Contains(t, got, `sensor@10 {`)
	require.Contains(t, got, `compatible = "sony,imx219";`)
	require.Contains(t, got, "reg = <0x10>;")
	require.Contains(t, got, "reset-gpios = <&tlmm 64 0>;")
	require.Contains(t, got, "data-lanes = <1 2>;")
	require.Contains(t, got, "link-frequencies = /bits/ 64 <456000000>;")
	require.Contains(t, got, "port@0 {")
	require.Contains(t, got, "data-lanes = <0 1>;")
}

func TestRenderDtsFourLanes(t *testing.T) {
	got := renderDts(spec{
		port:   Port{Index: 2, ResetGPIO: 82},
		sensor: Sensor{Name: "imx219", Compatible: "sony,imx219", I2CAddress: 0x10, LinkFrequencyHz: 456000000},
		lanes:  4,
	})

	require.Contains(t, got, "data-lanes = <1 2 3 4>;")
	require.Contains(t, got, "data-lanes = <0 1 2 3>;")
}

// If dtc is available, every rendered overlay must actually compile as a
// standalone plugin, with &cci0/&camss left as unresolved fixups.
func TestRenderedDtsCompiles(t *testing.T) {
	dtcPath, err := exec.LookPath("dtc")
	if err != nil {
		t.Skip("dtc not found in PATH")
	}

	for _, port := range Ports {
		for _, sensor := range Sensors {
			for _, lanes := range sensor.Lanes {
				source := renderDts(spec{port: port, sensor: sensor, lanes: lanes})

				dir := paths.New(t.TempDir())
				dts := dir.Join("overlay.dts")
				dtbo := dir.Join("overlay.dtbo")
				require.NoError(t, dts.WriteFile([]byte(source)))

				cmd := exec.Command(dtcPath, "-@", "-I", "dts", "-O", "dtb", "-o", dtbo.String(), dts.String())
				output, err := cmd.CombinedOutput()
				require.NoError(t, err, "dtc failed for csi%d/%s/%dlanes: %s", port.Index, sensor.Name, lanes, output)
			}
		}
	}
}

func TestEnsureBuiltOnlyBuildsKnownOverlays(t *testing.T) {
	recorder := executor.NewRecorder()
	overlaysDir := paths.New("/var/lib/arduino-linux-config/overlays")

	const staticOverlay = "monaco-monza-dsi-waveshare,8.0-dsi-touch-a.dtbo"
	err := EnsureBuilt(t.Context(), recorder, overlaysDir, []string{
		"monaco-monza-camera-csi0-imx219-2lane.dtbo",
		staticOverlay, // pre-shipped: must be left untouched
	})
	require.NoError(t, err)

	var ranDtc, movedStatic bool
	for _, effect := range recorder.Effects() {
		if strings.HasPrefix(effect, "dtc ") {
			ranDtc = true
		}
		if strings.Contains(effect, staticOverlay) {
			movedStatic = true
		}
	}
	require.True(t, ranDtc, "expected a dtc invocation, got effects: %v", recorder.Effects())
	require.False(t, movedStatic, "static overlay must not be touched, got effects: %v", recorder.Effects())
}
