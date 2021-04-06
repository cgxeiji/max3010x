package max3010x

import (
	"errors"
	"fmt"

	"github.com/cgxeiji/max3010x/max30102"
)

var (
	// ErrWrongDevice is thrown when trying to convert a max3010x.Device
	// interface to the underlying *Device struct and the device does not match
	// the PartID.
	ErrWrongDevice = errors.New("wrong device")
	// ErrNotDetected is thrown when trying to read a heart rate or SpO2 level
	// and nothing is detected on the sensor (e.g. no finger is placed on the
	// sensor when the function is called).
	ErrNotDetected = errors.New("nothing detected on the sensor")
	// ErrTooNoisy is thrown when trying to read data and has too much
	// variation, therefore consistent measurements cannot be done (e.g.
	// ambient light, moving finger, etc.).
	ErrTooNoisy = errors.New("data has too much noise")

	errLowValue = errors.New("low value")
)

// Device defines a MAX3010x device.
type Device struct {
	sensor sensor
	redLED *tSeries
	irLED  *tSeries
	readCh chan struct{}

	hr   movingAverage
	spo2 movingAverage

	bus  string
	addr uint16

	beat *beat

	// PartID is the byte part ID as set by the manufacturer.
	// MAX30100: 0x11 or max30100.PartID
	// MAX30102: 0x15 or max30102.PartID
	PartID byte
	RevID  byte
}

type sensor interface {
	Temperature() (float64, error)
	RevID() (byte, error)
	Reset() error
	Calibrate() error

	IRRed() (float64, float64, error)
	IRRedBatch() ([]float64, []float64, error)

	Shutdown() error
	Startup() error

	Close()
}

const threshold = 0.10

// New returns a new MAX3010x device.
func New(options ...Option) (*Device, error) {
	d := &Device{
		readCh: make(chan struct{}, 1),
		beat:   newBeat(),
		irLED:  newTSeries(64),
		redLED: newTSeries(64),
	}

	for _, option := range options {
		option(d)
	}

	sensor, err := max30102.New(d.bus, d.addr)
	if err != nil {
		return nil, err
	}
	d.sensor = sensor

	d.PartID = max30102.PartID

	if d.RevID, err = d.sensor.RevID(); err != nil {
		return nil, fmt.Errorf("max3010x: could not get revision ID: %w", err)
	}

	d.readCh <- struct{}{}

	return d, nil
}

// Close closes the devices and cleans after itself.
func (d *Device) Close() {
	d.sensor.Close()
}

// Calibrate calibrates the power of each LED.
func (d *Device) Calibrate() error {
	return d.sensor.Calibrate()
}

// Temperature returns the current temperature of the device.
func (d *Device) Temperature() (float64, error) {
	return d.sensor.Temperature()
}

// ToMax30102 converts a max3010x device to a max30102 device to access low
// level functions. Check the package max3010x/max30102 for detailed behavior.
func (d *Device) ToMax30102() (*max30102.Device, error) {
	device, ok := d.sensor.(*max30102.Device)
	if !ok {
		return nil, ErrWrongDevice
	}

	return device, nil
}

// Shutdown sets the device into power-save mode.
func (d *Device) Shutdown() error {
	return d.sensor.Shutdown()
}

// Startup wakes the device from power-save mode.
func (d *Device) Startup() error {
	return d.sensor.Startup()
}

func (d *Device) leds() error {
	select {
	case <-d.readCh:
		r, ir, err := d.sensor.IRRedBatch()
		if err != nil {
			return fmt.Errorf("could not get LEDs: %w", err)
		}
		d.redLED.add(r...)
		d.irLED.add(ir...)
		d.readCh <- struct{}{}

	default:
		select {
		case <-d.readCh:
			d.readCh <- struct{}{}
		}
	}
	return nil
}

func (d *Device) ledsSingle() error {
	select {
	case <-d.readCh:
		r, ir, err := d.sensor.IRRed()
		if err != nil {
			return fmt.Errorf("could not get LEDs: %w", err)
		}
		d.redLED.add(r)
		d.irLED.add(ir)
		d.readCh <- struct{}{}
	}
	return nil
}
