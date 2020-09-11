package max30102

import "fmt"

// Option defines a functional option for the device.
type Option func(d *Device) (Option, error)

// Options set different configuration options and returns the previous value
// of the last option passed.
func (d *Device) Options(options ...Option) (Option, error) {
	var old Option
	var err error
	for _, opt := range options {
		old, err = opt(d)
		if err != nil {
			return nil, err
		}
	}

	return old, nil
}

func (d *Device) config(reg, mask, flag byte) (byte, error) {
	cfg, err := d.Read(reg)
	if err != nil {
		return 0, fmt.Errorf("could not get %v from %v: %w", mask, reg, err)
	}
	old := cfg &^ mask
	cfg &= mask
	cfg |= flag
	if err := d.Write(reg, cfg); err != nil {
		return 0, fmt.Errorf("could not set %v in %v: %w", flag, reg, err)
	}

	return old, nil
}

// Mode sets the operation mode of the device.
func Mode(mode byte) Option {
	return func(d *Device) (Option, error) {
		old, err := d.config(ModeCfg, modeMask, mode)
		if err != nil {
			return nil, fmt.Errorf("max30102: could not configure mode: %w", err)
		}

		if err = d.Write(FIFOWrPtr, 0); err != nil {
			return nil, fmt.Errorf("max30102: could not configure mode: %w", err)
		}
		if err = d.Write(OvfCount, 0); err != nil {
			return nil, fmt.Errorf("max30102: could not configure mode: %w", err)
		}
		if err = d.Write(FIFORdPtr, 0); err != nil {
			return nil, fmt.Errorf("max30102: could not configure mode: %w", err)
		}

		return Mode(old), nil
	}
}

// RedPulseAmp sets the pulse amplitude of the red LED. It accepts values
// from 0.0 to 51.0 mA and the value is rounded down to the nearest multiple of 0.2.
func RedPulseAmp(current float64) Option {
	return func(d *Device) (Option, error) {
		if current > 51 {
			current = 51
		}
		if current < 0 {
			current = 0
		}
		b := byte(current * 5)

		old, err := d.config(Led1PA, 0, b)
		if err != nil {
			return nil, fmt.Errorf("max30102: could not configure red LED pulse amplitud: %w", err)
		}

		return RedPulseAmp(float64(old) / 5), nil
	}
}

// IRPulseAmp sets the pulse amplitude of the red LED. It accepts values
// from 0.0 to 51.0 mA and the value is rounded down to the nearest multiple of 0.2.
func IRPulseAmp(current float64) Option {
	return func(d *Device) (Option, error) {
		if current > 51 {
			current = 51
		}
		if current < 0 {
			current = 0
		}
		b := byte(current * 5)

		old, err := d.config(Led2PA, 0, b)
		if err != nil {
			return nil, fmt.Errorf("max30102: could not configure IR LED pulse amplitud: %w", err)
		}

		return IRPulseAmp(float64(old) / 5), nil
	}
}

// PulseWidth sets the pulse width of the device.
func PulseWidth(pw byte) Option {
	return func(d *Device) (Option, error) {
		old, err := d.config(SpO2Cfg, pwMask, pw)
		if err != nil {
			return nil, fmt.Errorf("max30102: could not configure pulse width: %w", err)
		}

		return PulseWidth(old), nil
	}
}

// SampleRate sets the SpO2 sample rate control of the device.
func SampleRate(sr byte) Option {
	return func(d *Device) (Option, error) {
		old, err := d.config(SpO2Cfg, srMask, sr)
		if err != nil {
			return nil, fmt.Errorf("max30102: could not configure sample rate: %w", err)
		}

		return SampleRate(old), nil
	}
}

// InterruptEnable enables interrupts.
func InterruptEnable(i byte) Option {
	return func(d *Device) (Option, error) {
		old, err := d.config(IntEna1, ^i, i)
		if err != nil {
			return nil, fmt.Errorf("max30102: could not configure interrupt flags: %w", err)
		}

		return InterruptEnable(old), nil
	}
}

// AlmostFullValue sets when the AlmostFull interrupt should be triggered. It
// can take values from 0 to 15.
func AlmostFullValue(left byte) Option {
	return func(d *Device) (Option, error) {
		left &= ^fifoFullMask
		old, err := d.config(FIFOCfg, fifoFullMask, left)
		if err != nil {
			return nil, fmt.Errorf("max30102: could not configure almost full value to %d: %w", left, err)
		}

		return AlmostFullValue(old), nil
	}
}
