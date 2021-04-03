package max3010x

// An Option configures a device.
type Option interface {
	Apply(*Device)
}

// OptionFunc is a function that configures a device.
type OptionFunc func(device *Device)

// Apply calls OptionFunc on device instance
func (f OptionFunc) Apply(device *Device) {
	f(device)
}

// WithSpecificBus can be used to specify I²C bus name
// ("/dev/i2c-2", "I2C2", "2").
func WithSpecificBus(name string) Option {
	return OptionFunc(func(d *Device) {
		d.bus = name
	})
}

// WithAddress can be used to specify alternative I²C name.
// if default (0x57) is unavailable and changed
func WithAddress(addr uint16) Option {
	return OptionFunc(func(d *Device) {
		d.addr = addr
	})
}

// WithSensor can be used to mock alternative sensor implementation.
func WithSensor(sensor sensor) Option {
	return OptionFunc(func(d *Device) {
		d.sensor = sensor
	})
}
