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

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/arduino/arduino-linux-config/cmd/config"
	"github.com/arduino/go-paths-helper"
	"go.bug.st/f"
)

var (
	GetCarriers = sync.OnceValue(func() *Carriers {
		cfg, _ := config.NewConfigFromEnv()
		return f.Must(LoadConfigs(cfg.CarriersConfig()))
	})
)

type CarrierDeviceName string

const (
	None    CarrierDeviceName = "none"
	Camera0 CarrierDeviceName = "camera0"
	Camera1 CarrierDeviceName = "camera1"
	Display CarrierDeviceName = "display"
)

type Carriers struct {
	Carriers []Carrier
}

type Carrier struct {
	Name    string   `json:"name"`
	Devices []Device `json:"devices"`
}

// Device represents a configurable hardware device on a carrier
type Device struct {
	Name    string         `json:"name"`
	Options []DeviceOption `json:"options"`
}

// DeviceOption represents a configuration option for a device
type DeviceOption struct {
	Name             string   `json:"name"`
	DtboFiles        []string `json:"dtboFiles"`
	IncompatibleDtbo []string `json:"incompatibleDtboFiles,omitempty"`
}

// used to read the json
type deviceWrapper struct {
	Devices []Device `json:"devices"`
}

// LoadConfigs scans the config directory and populates the Carriers struct
func LoadConfigs(configPath *paths.Path) (*Carriers, error) {
	carriers := &Carriers{
		Carriers: make([]Carrier, 0),
	}

	entries, err := os.ReadDir(configPath.String())
	if err != nil {
		return nil, fmt.Errorf("failed to read config dir: %w", err)
	}

	for _, entry := range entries {
		// skip non-JSON
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		fullPath := filepath.Join(configPath.String(), entry.Name())

		devices, err := ReadCarrierConfig(fullPath)

		if err != nil {
			slog.Warn("Warning: skipping invalid config", slog.String("filename", entry.Name()))
			continue
		}

		carrier := Carrier{
			Name:    strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())),
			Devices: devices,
		}
		carriers.Carriers = append(carriers.Carriers, carrier)
	}

	slog.Info("Loaded configuration files", slog.Int("carriers", len(carriers.Carriers)))
	return carriers, nil
}

func ReadCarrierConfig(filePath string) ([]Device, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file error: %w", err)
	}

	var wrapper deviceWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	if wrapper.Devices == nil {
		return nil, fmt.Errorf("no 'devices' key found in JSON or tag mismatch")
	}
	return wrapper.Devices, nil
}
