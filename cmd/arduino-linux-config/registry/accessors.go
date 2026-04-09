package registry

func GetDevices(carrierName string) ([]Device, bool) {
	carrier, ok := Registry.Carriers[CarrierName(carrierName)]
	return carrier.Devices, ok
}

func GetDevicesNames(carrierName string) []CarrierDeviceName {
	carrier, ok := Registry.Carriers[CarrierName(carrierName)]
	if !ok {
		return []CarrierDeviceName{}
	}

	result := make([]CarrierDeviceName, len(carrier.Devices))
	for i, device := range carrier.Devices {
		result[i] = CarrierDeviceName(device.Name)
	}
	return result
}

func FindDevice(carrierName string, deviceName CarrierDeviceName) (*Device, bool) {
	carrier, ok := Registry.Carriers[CarrierName(carrierName)]
	if !ok {
		return nil, false
	}

	for i := range carrier.Devices {
		if carrier.Devices[i].Name == string(deviceName) {
			return &carrier.Devices[i], true
		}
	}

	return nil, false
}

func CarrierExists(carrierName string) bool {
	_, exists := Registry.Carriers[CarrierName(carrierName)]
	return exists
}
