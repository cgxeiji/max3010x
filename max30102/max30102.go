package max30102

import (
	"errors"
	"fmt"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
)

var (
	// ErrNotDevice throws an error when the device part ID does not match a
	// MAX30102 signature (0x15).
	ErrNotDevice error = errors.New("max30102: part ID does not match (0x15)")
)

// Device defines a MAX30102 device.
type Device struct {
	dev *i2c.Dev
	bus i2c.BusCloser
}

// New returns a new MAX30102 device.
func New() (*Device, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("max30102: could not initialize host: %w", err)
	}

	bus, err := i2creg.Open("")
	if err != nil {
		return nil, fmt.Errorf("max30102: could not open i2c bus: %w", err)
	}
	dev := &i2c.Dev{
		Addr: Addr,
		Bus:  bus,
	}

	d := &Device{
		dev: dev,
		bus: bus,
	}

	part, err := d.Read(RegPartID)
	if err != nil {
		return nil, fmt.Errorf("max30102: could not get part ID: %w", err)
	}
	if part != PartID {
		return nil, ErrNotDevice
	}

	return d, nil
}

// Close closes the devices and cleans after itself.
func (d *Device) Close() {
	d.bus.Close()
}

// RevID returns the revision ID of the device.
func (d *Device) RevID() (byte, error) {
	rev, err := d.Read(RegRevID)
	if err != nil {
		return 0, fmt.Errorf("max30102: could not get revision ID: %w", err)
	}
	return rev, nil
}

func (d *Device) tempEnable() error {
	if err := d.Write(TempCfg, TempEna); err != nil {
		return fmt.Errorf("max30102: could not enable temperature: %w", err)
	}
	return nil
}

func (d *Device) tempReady() (bool, error) {
	state, err := d.Read(TempCfg)
	if err != nil {
		return false, fmt.Errorf("max30102: could not read temperature state: %w", err)
	}
	return (state & TempEna) == 0, nil
}

// Temperature returns the current temperature of the device.
func (d *Device) Temperature() (float64, error) {
	if err := d.tempEnable(); err != nil {
		return 0, err
	}
	for {
		if r, err := d.tempReady(); r == true {
			break
		} else if err != nil {
			return 0, err
		}
	}

	i, err := d.Read(TempInt)
	if err != nil {
		return 0, fmt.Errorf("max30102: could not read integer part of temperature: %w", err)
	}

	f, err := d.Read(TempFrac)
	if err != nil {
		return 0, fmt.Errorf("max30102: could not read fractional part of temperature: %w", err)
	}

	return float64(i) + (float64(f) * 0.0625), nil
}

// Read reads a single byte from a register.
func (d *Device) Read(reg byte) (byte, error) {
	b := make([]byte, 1)
	if err := d.dev.Tx([]byte{reg}, b); err != nil {
		return 0, fmt.Errorf("max30102: could not read byte: %w", err)
	}

	return b[0], nil
}

// ReadBytes read n bytes from a register.
func (d *Device) ReadBytes(reg byte, n int) ([]byte, error) {
	b := make([]byte, n)
	if err := d.dev.Tx([]byte{reg}, b); err != nil {
		return nil, fmt.Errorf("max30102: could not read %d bytes: %w", n, err)
	}

	return b, nil
}

// Write writes a byte to a register.
func (d *Device) Write(reg, data byte) error {
	n, err := d.dev.Write([]byte{reg, data})
	if err != nil {
		return err
	}
	n-- // remove register write
	if n != 1 {
		return fmt.Errorf("write: wrong number of bytes written: want %d, got %d", 1, n)
	}

	return nil
}
