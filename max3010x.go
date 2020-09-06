package max3010x

import (
	"errors"
	"fmt"
	"time"

	"github.com/cgxeiji/max3010x/max30102"
)

// Errors
var (
	ErrWrongDevice = errors.New("wrong device")
	ErrNotDetected = errors.New("nothing detect on the sensor")
	errLowValue    = errors.New("low value")
)

// Device defines a MAX3010x device.
type Device struct {
	sensor sensor
	redLED tSeries
	irLED  tSeries

	PartID byte
	RevID  byte
}

type sensor interface {
	Temperature() (float64, error)
	RevID() (byte, error)
	Reset() error
	RedIR() (float64, float64, error)

	Close()
}

const threshold = 0.10

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

	d.redLED.init(16, 16)
	d.irLED.init(16, 16)

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

// HeartRate returns the current heart rate. Heart rate is expected to be
// between 10 to 250 beats per minute. Values outside that range are considered
// invalid and the function will continue to sample until a valid bpm is found.
// If no contact is detect on the sensor, this function returns 0 with an
// ErrNotDetected error.
func (d *Device) HeartRate() (float64, error) {
	var span time.Duration

	for {
		if err := d.detectFall(); errors.Is(err, errLowValue) {
			return 0, fmt.Errorf("max3010x: could not get heart rate: %w", ErrNotDetected)
		} else if err != nil {
			return 0, fmt.Errorf("max3010x: could not get heart rate: %w", err)
		}
		timer := time.Now()

		if err := d.detectFall(); errors.Is(err, errLowValue) {
			return 0, fmt.Errorf("max3010x: could not get heart rate: %w", ErrNotDetected)
		} else if err != nil {
			return 0, fmt.Errorf("max3010x: could not get heart rate: %w", err)
		}
		span = time.Since(timer)

		if span > 6*time.Second { // less than 10 bpm
			continue // invalid
		}
		if span < 238*time.Millisecond { // more than 250 bpm
			continue // invalid
		}
		break
	}

	return 60000 / (float64(span.Milliseconds())), nil
}

func (d *Device) detectFall() error {
	var r1, r2 float64
	var err error
	const iter = 8
	count := iter
	err = d.leds()
	if err != nil {
		return fmt.Errorf("detectFall: %w", err)
	}
	r1 = d.redLED.mean
	for {
		err = d.leds()
		if err != nil {
			return fmt.Errorf("detectFall: %w", err)
		}
		r2 = d.redLED.mean
		if r2 < threshold {
			return errLowValue
		}

		delta := r2 - r1
		if delta < 0 {
			count--
			if count <= 0 {
				return nil
			}
		} else {
			count = iter
		}
		r1 = r2
	}
}

func (d *Device) leds() error {
	r, ir, err := d.sensor.RedIR()
	if err != nil {
		return fmt.Errorf("could not get LEDs: %w", err)
	}
	d.redLED.add(r)
	d.irLED.add(ir)
	return nil
}

// Leds is fantastic.
func (d *Device) Leds() (float64, float64, error) {
	err := d.leds()
	if err != nil {
		return 0, 0, err
	}
	return d.redLED.mean, d.irLED.mean, nil
}

// SpO2 returns the SpO2 value in 100%.
func (d *Device) SpO2() (float64, error) {
	r, err := d.rValue()
	if errors.Is(err, errLowValue) {
		return 0, fmt.Errorf("max3010x: could not get SpO2: %w", ErrNotDetected)
	} else if err != nil {
		return 0, fmt.Errorf("max3010x: could not get R value: %w", err)
	}

	return 104 - 17*r, nil
}

func (d *Device) rValue() (float64, error) {
	err := d.leds()
	if err != nil {
		return 0, err
	}

	if d.redLED.mean < 0.01 || d.irLED.mean < 0.01 {
		return 0, errLowValue
	}

	return d.redLED.acdc() / d.irLED.acdc(), nil
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
