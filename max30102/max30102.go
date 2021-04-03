package max30102

import (
	"errors"
	"fmt"
	"time"

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

// New returns a new MAX30102 device. By default, this sets the LED pulse
// amplitude to 2.4mA, with a pulse width of 411us and a sample rate of 100
// samples/s.
//
// Argument "busName" can be used to specify the exact bus to use ("/dev/i2c-2", "I2C2", "2").
// Argument "addr" can be used to specify alternative address if default (0x57) is unavailable and changed.
// If "busName" argument is specified as an empty string "" the first available bus will be used.
func New(busName string, addr uint16) (*Device, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("max30102: could not initialize host: %w", err)
	}

	bus, err := i2creg.Open(busName)
	if err != nil {
		return nil, fmt.Errorf("max30102: could not open I2C bus: %w", err)
	}

	if addr == 0 {
		addr = Addr
	}

	dev := &i2c.Dev{
		Addr: addr,
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

	err = d.Reset()
	if err != nil {
		return nil, fmt.Errorf("max30102: could not reset device: %w", err)
	}
	if _, err = d.Options(
		RedPulseAmp(2.8),
		IRPulseAmp(2.8),
		PulseWidth(PW411),
		SampleRate(SR100),
		InterruptEnable(NewFIFOData|AlmostFull),
		AlmostFullValue(0),
		Mode(ModeSpO2),
	); err != nil {
		return nil, fmt.Errorf("max30102: could not initialize device: %w", err)
	}
	d.drain()

	return d, nil
}

// Close closes the devices and cleans after itself.
func (d *Device) Close() {
	d.Shutdown()
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

func (d *Device) waitUntil(reg, flag byte, bit byte) error {
	switch bit {
	case 1:
		for {
			state, err := d.Read(reg)
			if err != nil {
				return fmt.Errorf("could not wait for %v in %v to be %v", flag, reg, bit)
			} else if state&flag != 0 {
				//fmt.Printf("%#x = %#b\n", reg, state)
				return nil
			}
		}
	case 0:
		for {
			if state, err := d.Read(reg); err != nil {
				return fmt.Errorf("could not wait for %v in %v to be %v", flag, reg, bit)
			} else if state&flag == 0 {
				//fmt.Printf("%#x = %#b\n", reg, state)
				return nil
			}
		}
	}

	return fmt.Errorf("invalid bit %v, it should be 1 or 0", bit)
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
	if err := d.waitUntil(TempCfg, TempEna, 0); err != nil {
		return 0, err
	}

	i, err := d.Read(TempInt)
	if err != nil {
		return 0, fmt.Errorf("max30102: could not read integer part of temperature: %w", err)
	}

	f, err := d.Read(TempFrac)
	if err != nil {
		return 0, fmt.Errorf("max30102: could not read fractional part of temperature: %w", err)
	}

	return float64(int8(i)) + (float64(f) * 0.0625), nil
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

// Reset resets the device. All configurations, thresholds, and data registers
// are reset to their power-on state.
func (d *Device) Reset() error {
	if err := d.Write(ModeCfg, ResetControl); err != nil {
		return fmt.Errorf("max30102: could not reset: %w", err)
	}
	if err := d.waitUntil(ModeCfg, ResetControl, 0); err != nil {
		return fmt.Errorf("max30102: could not reset: %w", err)
	}

	return nil
}

// IRRed returns the value of the red LED and IR LED. The values are normalized
// from 0.0 to 1.0.
func (d *Device) IRRed() (ir, red float64, err error) {
	const msbMask byte = 0b0000_0011

	err = d.waitUntil(IntStat1, NewFIFOData, 1)
	if err != nil {
		return 0, 0, err
	}

	bytes, err := d.ReadBytes(FIFOData, 6)
	if err != nil {
		return 0, 0, err
	}

	ir = float64(
		int(bytes[3]&msbMask)<<16|
			int(bytes[4])<<8|
			int(bytes[5])) / maxADC
	red = float64(
		int(bytes[0]&msbMask)<<16|
			int(bytes[1])<<8|
			int(bytes[2])) / maxADC

	return ir, red, nil
}

// IRRedBatch returns a batch of IR and red LED values based on the AlmostFull
// flag. The amount of data returned can be configured by setting the
// AlmostFullValue leftover value, which is set to 0 by default. Therefore,
// this function returns 32 samples by default.
func (d *Device) IRRedBatch() (ir, red []float64, err error) {
	const maxADC = 262143
	const msbMask byte = 0b0000_0011

	err = d.drain()
	if err != nil {
		return nil, nil, fmt.Errorf("max30102: could not empty FIFO: %w", err)
	}
	err = d.waitUntil(IntStat1, AlmostFull, 1)
	if err != nil {
		return nil, nil, fmt.Errorf("max30102: error waiting for almost full interrupt: %w", err)
	}

	n, err := d.available()
	if err != nil {
		return nil, nil, fmt.Errorf("max30102: error reading available data: %w", err)
	}

	ir = make([]float64, n)
	red = make([]float64, n)
	for i := 0; i < n; i++ {
		bytes, err := d.ReadBytes(FIFOData, 6)
		if err != nil {
			return nil, nil, err
		}

		irData := float64(
			int(bytes[3]&msbMask)<<16|
				int(bytes[4])<<8|
				int(bytes[5])) / maxADC
		redData := float64(
			int(bytes[0]&msbMask)<<16|
				int(bytes[1])<<8|
				int(bytes[2])) / maxADC

		ir[i] = irData
		red[i] = redData
	}

	return ir, red, nil
}

func (d *Device) drain() error {
	n, err := d.available()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		_, err := d.ReadBytes(FIFOData, 6)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) available() (int, error) {
	wr, err := d.Read(FIFOWrPtr)
	if err != nil {
		return 0, nil
	}
	rd, err := d.Read(FIFORdPtr)
	if err != nil {
		return 0, nil
	}

	if wr == rd {
		return 32, nil
	}
	return (int(wr) + 32 - int(rd)) % 32, nil
}

// Calibrate auto-calibrates the current of each LED.
func (d *Device) Calibrate() error {
	var ir []float64
	var red []float64
	var err error

	irAmp := 0.0
	redAmp := 0.0

	if _, err = d.Options(
		IRPulseAmp(irAmp),
		RedPulseAmp(redAmp),
	); err != nil {
		return fmt.Errorf("max30102: could not calibrate sensor: %w", err)
	}

	for mean(ir) < 0.4 {
		if irAmp >= 5 {
			break
		}
		irAmp += 0.5

		if _, err = d.Options(
			IRPulseAmp(irAmp),
		); err != nil {
			return fmt.Errorf("max30102: could not calibrate sensor: %w", err)
		}
		time.Sleep(40 * time.Millisecond)

		ir, red, err = d.IRRedBatch()
		if err != nil {
			return fmt.Errorf("max30102: could not calibrate sensor: %w", err)
		}
	}

	for mean(red) < 0.4 {
		if redAmp >= 5 {
			break
		}
		redAmp += 0.5

		if _, err = d.Options(
			RedPulseAmp(redAmp),
		); err != nil {
			return fmt.Errorf("max30102: could not calibrate sensor: %w", err)
		}
		time.Sleep(40 * time.Millisecond)

		ir, red, err = d.IRRedBatch()
		if err != nil {
			return fmt.Errorf("max30102: could not calibrate sensor: %w", err)
		}
	}

	fmt.Println("calibration:")
	fmt.Printf("  irAmp = %.1fmA\n", irAmp)
	fmt.Printf("  redAmp = %.1fmA\n", redAmp)

	return nil
}

func mean(a []float64) float64 {
	if len(a) == 0 {
		return 0
	}

	r := 0.0
	for _, v := range a {
		r += v
	}

	return r / float64(len(a))
}

// Shutdown sets the device into power-save mode.
func (d *Device) Shutdown() error {
	_, err := d.config(ModeCfg, ^modeSHDN, modeSHDN)

	return err
}

// Startup wakes the device from power-save mode.
func (d *Device) Startup() error {
	_, err := d.config(ModeCfg, ^modeSHDN, ^modeSHDN)

	return err
}

func (d *Device) debugRegister(reg byte) {
	b, _ := d.Read(reg)
	fmt.Printf("%#x = %#x (%#b)\n", reg, b, b)
}
