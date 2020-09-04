package max3010x

import (
	"fmt"

	"github.com/cgxeiji/max3010x/max30102"
)

// Device defines a MAX3010x device.
type Device struct {
	sensor sensor

	PartID byte
	RevID  byte
}

type sensor interface {
	Temperature() (float64, error)
	RevID() (byte, error)

	Close()
}

// New returns a new MAX3010x device.
func New() (*Device, error) {
	sensor, err := max30102.New()
	if err != nil {
		return nil, err
	}

	d := &Device{
		sensor: sensor,
	}

	d.PartID = max30102.PartID

	if d.RevID, err = d.sensor.RevID(); err != nil {
		return nil, fmt.Errorf("max3010x: could not get revision ID: %w", err)
	}

	return d, nil
}

// Close closes the devices and cleans after itself.
func (d *Device) Close() {
	d.sensor.Close()
}

// Temperature returns the current temperature of the device.
func (d *Device) Temperature() (float64, error) {
	return d.sensor.Temperature()
}
