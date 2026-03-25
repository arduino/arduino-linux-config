package registry

import (
	"fmt"
)

// TODO cleanup, change name, refactor
func ValidateInput(carrierName string, rawDevice string, rawOption string) (bool, error) {
	devices, exists := GetDevices(*GetCarriers(), carrierName)
	if !exists {
		return false, fmt.Errorf("carrier %q not supported", carrierName)
	}

	device, exists := deviceExists(rawDevice, devices)
	if !exists {
		return false, fmt.Errorf("unknown device for carrier %s: %q", carrierName, rawDevice)
	}

	if !isOptionValid(rawOption, device) {
		return false, fmt.Errorf("device %q does not support option %q", rawDevice, rawOption)
	}

	return true, nil
}

func CarrierExists(CarrierName string) bool {
	for _, carrier := range GetCarriers().Carriers {
		if string(carrier.Name) == CarrierName {
			return true
		}
	}
	return false
}

func deviceExists(deviceName string, devices []Device) (Device, bool) {
	for _, device := range devices {
		if string(device.Name) == deviceName {
			return device, true
		}
	}
	return Device{}, false
}

func isOptionValid(optionName string, device Device) bool {
	for _, option := range device.Options {
		if optionName == option.Name {
			return true
		}
	}
	return false
}

func GetDevices(carriers Carriers, carrierName string) ([]Device, bool) {
	for _, c := range carriers.Carriers {
		if c.Name == carrierName {
			return c.Devices, true
		}
	}
	return nil, false
}

func GetCarrierDevices(carrierName string) []MediaCarrierDeviceName {
	for _, c := range GetCarriers().Carriers {
		if c.Name == carrierName {
			result := make([]MediaCarrierDeviceName, 0, len(c.Devices))
			for _, device := range c.Devices {
				result = append(result, MediaCarrierDeviceName(device.Name))
			}
			return result
		}
	}
	return []MediaCarrierDeviceName{}
}

func GetDevice(carriers Carriers, carrierName string, deviceName string) (*Device, bool) {
	for _, c := range carriers.Carriers {
		if c.Name == carrierName {
			for i := range c.Devices {
				if string(c.Devices[i].Name) == deviceName {
					return &c.Devices[i], true
				}
			}
		}
	}
	return nil, false
}

func GetOptions(device Device, optionName string) []string {
	for _, opt := range device.Options {
		if opt.Name == optionName {
			return opt.DtboFiles
		}
	}
	return nil
}
