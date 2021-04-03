package max3010x

// An Option configures a device.
type Option func(d *Device) Option

// OnBus can be used to specify I²C bus name
// ("/dev/i2c-2", "I2C2", "2"). By default, the bus name is "", which selects
// the first available bus.
func OnBus(name string) Option {
	return func(d *Device) Option {
		old := d.bus
		d.bus = name
		return OnBus(old)
	}
}

// OnAddr can be used to specify alternative I²C name.
// By default, the address is 0x57.
func OnAddr(addr uint16) Option {
	return func(d *Device) Option {
		old := d.addr
		d.addr = addr
		return OnAddr(old)
	}
}
