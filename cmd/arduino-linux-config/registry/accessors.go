package registry

func GetDevices(carrierName string) ([]Device, bool) {
	for _, c := range Registry.Carriers {
		if string(c.Name) == carrierName {
			return c.Devices, true
		}
	}
	return nil, false
}

func GetDevicesNames(carrierName string) []CarrierDeviceName {
	for _, c := range Registry.Carriers {
		if string(c.Name) == carrierName {
			result := make([]CarrierDeviceName, 0, len(c.Devices))
			for _, device := range c.Devices {
				result = append(result, CarrierDeviceName(device.Name))
			}
			return result
		}
	}
	return []CarrierDeviceName{}
}

func FindDevice(carrierName string, deviceName CarrierDeviceName) (*Device, bool) {
	for _, c := range Registry.Carriers {
		if string(c.Name) == carrierName {
			for i := range c.Devices {
				if c.Devices[i].Name == string(deviceName) {
					return &c.Devices[i], true
				}
			}
		}
	}
	return nil, false
}

func GetDeviceOptions(device Device, optionName string) []string {
	for _, opt := range device.Options {
		if opt.Name == optionName {
			return opt.DtboFiles
		}
	}
	return nil
}

func CarrierExists(carrierName string) bool {
	for _, c := range Registry.Carriers {
		if string(c.Name) == carrierName {
			return true
		}
	}
	return false
}
