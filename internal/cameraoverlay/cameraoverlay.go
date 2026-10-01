// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

// Package cameraoverlay builds VentunoQ's per-CSI-port camera overlays on the
// board from a template, instead of shipping one pre-built dtbo per combination.
package cameraoverlay

import (
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-linux-config/internal/executor"
)

// Sensor describes a camera sensor this package knows how to wire up.
type Sensor struct {
	Name            string
	Compatible      string
	I2CAddress      uint8
	LinkFrequencyHz uint64
}

// Port describes one of VentunoQ's built-in CSI camera connectors.
type Port struct {
	Index     int
	ResetGPIO int
}

// Option is one sensor/lane-count combination available on a port.
type Option struct {
	Name     string
	DtboFile string
}

// Sensors is the hardcoded catalog of supported camera sensors.
var Sensors = []Sensor{
	{Name: "imx219", Compatible: "sony,imx219", I2CAddress: 0x10, LinkFrequencyHz: 456000000},
}

// Ports is VentunoQ's 3 built-in CSI connectors, in cci0/1/2 order.
var Ports = []Port{
	{Index: 0, ResetGPIO: 64},
	{Index: 1, ResetGPIO: 75},
	{Index: 2, ResetGPIO: 82},
}

// Lanes is every lane count the template supports.
var Lanes = []int{2, 4}

// OptionsForPort enumerates every sensor/lane combination available on a port.
func OptionsForPort(port Port) []Option {
	options := make([]Option, 0, len(Sensors)*len(Lanes))
	for _, sensor := range Sensors {
		for _, lanes := range Lanes {
			name := fmt.Sprintf("%s-%dlanes", sensor.Name, lanes)
			options = append(options, Option{Name: name, DtboFile: dtboFilename(port, sensor, lanes)})
		}
	}
	return options
}

// Filenames lists every dtbo filename this package can build, across every
// port/sensor/lanes combination.
func Filenames() []string {
	filenames := make([]string, 0, len(Ports)*len(Sensors)*len(Lanes))
	for _, port := range Ports {
		for _, option := range OptionsForPort(port) {
			filenames = append(filenames, option.DtboFile)
		}
	}
	return filenames
}

// EnsureBuilt compiles, into overlaysDir, whichever of the requested overlays
// this package recognizes. Any other filename is left untouched: it is
// assumed to be a pre-shipped, static dtbo.
func EnsureBuilt(ctx context.Context, exec executor.Executor, overlaysDir *paths.Path, overlays []string) error {
	specs := specsByFilename()
	for _, overlay := range overlays {
		spec, ok := specs[overlay]
		if !ok {
			continue
		}
		if err := build(ctx, exec, overlaysDir, overlay, spec); err != nil {
			return fmt.Errorf("failed to build %s: %w", overlay, err)
		}
	}
	return nil
}

// spec holds the (port, sensor, lanes) combination behind a generated dtbo filename.
type spec struct {
	port   Port
	sensor Sensor
	lanes  int
}

// specsByFilename maps every buildable dtbo filename back to its spec.
func specsByFilename() map[string]spec {
	specs := make(map[string]spec, len(Ports)*len(Sensors)*len(Lanes))
	for _, port := range Ports {
		for _, sensor := range Sensors {
			for _, lanes := range Lanes {
				specs[dtboFilename(port, sensor, lanes)] = spec{port: port, sensor: sensor, lanes: lanes}
			}
		}
	}
	return specs
}

func dtboFilename(port Port, sensor Sensor, lanes int) string {
	return fmt.Sprintf("monaco-monza-camera-csi%d-%s-%dlane.dtbo", port.Index, sensor.Name, lanes)
}

// build renders, compiles and installs the overlay for one spec, via
// temporary files so a failed dtc run never leaves a partial dtbo behind.
func build(ctx context.Context, exec executor.Executor, overlaysDir *paths.Path, filename string, s spec) error {
	if err := exec.MkdirAll(overlaysDir); err != nil {
		return fmt.Errorf("failed to create %s: %w", overlaysDir, err)
	}

	tmp := fmt.Sprintf("camera-overlay.%d", time.Now().UnixNano())
	tmpDts := overlaysDir.Join(tmp + ".dts")
	tmpDtbo := overlaysDir.Join(tmp + ".dtbo")
	defer func() { _ = exec.Remove(tmpDts) }()

	if err := exec.WriteFile(tmpDts, []byte(renderDts(s)), 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", tmpDts, err)
	}

	if err := exec.Run(ctx, "dtc", "-@", "-I", "dts", "-O", "dtb", "-o", tmpDtbo.String(), tmpDts.String()); err != nil {
		return err
	}

	return exec.Rename(tmpDtbo, overlaysDir.Join(filename))
}

// laneList renders a MIPI CSI-2 data-lanes property counting up from "from".
func laneList(from, count int) string {
	lanes := make([]string, count)
	for i := range lanes {
		lanes[i] = fmt.Sprintf("%d", from+i)
	}
	return "<" + strings.Join(lanes, " ") + ">"
}

type templateData struct {
	Port            int
	ResetGPIO       int
	SensorNode      string
	Compatible      string
	I2CAddress      string
	SensorDataLanes string
	LinkFrequencyHz uint64
	CamssDataLanes  string
}

func renderDts(s spec) string {
	data := templateData{
		Port:            s.port.Index,
		ResetGPIO:       s.port.ResetGPIO,
		SensorNode:      fmt.Sprintf("%x", s.sensor.I2CAddress),
		Compatible:      s.sensor.Compatible,
		I2CAddress:      fmt.Sprintf("0x%x", s.sensor.I2CAddress),
		SensorDataLanes: laneList(1, s.lanes),
		LinkFrequencyHz: s.sensor.LinkFrequencyHz,
		CamssDataLanes:  laneList(0, s.lanes),
	}

	var buf strings.Builder
	// The template is fixed and every field above is produced by this package,
	// so execution against it cannot fail.
	if err := dtsTemplate.Execute(&buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

// dtsTemplate inlines GPIO_ACTIVE_HIGH as 0 (its dt-bindings/gpio/gpio.h
// value) so compiling it needs no C preprocessor, only dtc.
var dtsTemplate = template.Must(template.New("camera-overlay").Parse(`/dts-v1/;
/plugin/;

&clocks {
	cam24m: cam-clk {
		compatible = "fixed-clock";
		#clock-cells = <0>;
		clock-frequency = <24000000>;
		clock-output-names = "cam24m";
	};
};

&cci{{.Port}} {
	status = "okay";
};

&cci{{.Port}}_i2c0 {
	#address-cells = <1>;
	#size-cells = <0>;

	sensor@{{.SensorNode}} {
		compatible = "{{.Compatible}}";
		reg = <{{.I2CAddress}}>;
		clocks = <&cam24m>;

		reset-gpios = <&tlmm {{.ResetGPIO}} 0>; /* GPIO_ACTIVE_HIGH */

		port {
			/* MIPI CSI-2 bus endpoint */
			cam{{.Port}}_ep: endpoint {
				remote-endpoint = <&csiphy{{.Port}}_ep>;
				clock-lanes = <0>;
				data-lanes = {{.SensorDataLanes}};
				link-frequencies = /bits/ 64 <{{.LinkFrequencyHz}}>;
			};
		};
	};
};

&camss {
	vdda-phy-supply = <&vreg_l4a>;
	vdda-pll-supply = <&vreg_l5a>;

	status = "okay";

	ports {
		#address-cells = <1>;
		#size-cells = <0>;

		port@{{.Port}} {
			reg = <{{.Port}}>;

			csiphy{{.Port}}_ep: endpoint {
				data-lanes = {{.CamssDataLanes}};
				remote-endpoint = <&cam{{.Port}}_ep>;
			};
		};
	};
};
`))
