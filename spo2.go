package max3010x

import (
	"errors"
	"fmt"
)

// SpO2 returns the SpO2 value in 100%.
func (d *Device) SpO2() (float64, error) {
	r, err := d.rValue()
	if errors.Is(err, errLowValue) {
		d.spo2.reset()
		return 0, fmt.Errorf("max3010x: could not get SpO2: %w", ErrNotDetected)
	} else if err != nil {
		d.spo2.reset()
		return 0, fmt.Errorf("max3010x: could not get R value: %w", err)
	}

	spo2 := 104 - 17*r
	if spo2 <= 0 {
		return 0, nil
	}

	// if first measurement, pre-fill values.
	if d.spo2.mean == 0 {
		d.spo2.mean = spo2
	}
	d.spo2.add(spo2)

	return d.spo2.mean, nil
}

func (d *Device) rValue() (float64, error) {
	err := d.leds()
	if err != nil {
		return 0, err
	}

	if d.redLED.last() < threshold || d.irLED.last() < threshold {
		return 0, errLowValue
	}

	irACDC := d.irLED.acdc()
	if irACDC == 0 {
		return 0, nil
	}

	return d.redLED.acdc() / irACDC, nil
}
