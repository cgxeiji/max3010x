package max3010x

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// HeartRate returns the current heart rate. Heart rate is expected to be
// between 10 to 250 beats per minute. Values outside that range are considered
// invalid and the function will continue to sample until a valid bpm is found.
// If no contact is detect on the sensor, this function returns 0 with an
// ErrNotDetected error. If the sensor cannot detect a beat after 1s, it
// returns 0 with an ErrTooNoisy error.
func (d *Device) HeartRate() (float64, error) {
	type beatPkg struct {
		span float64
		err  error
	}
	beatCh := make(chan beatPkg)

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	go func(ctx context.Context) {
		if err := d.detectBeat(ctx); err != nil {
			beatCh <- beatPkg{
				err: err,
			}
			return
		}
		timer := time.Now()

		for {
			select {
			case <-ctx.Done():
				beatCh <- beatPkg{
					err: ctx.Err(),
				}
			default:
			}

			if err := d.detectBeat(ctx); err != nil {
				beatCh <- beatPkg{
					err: err,
				}
				return
			}
			t := time.Since(timer)

			if t > 6*time.Second { // less than 10 bpm
				continue // invalid
			}
			if t < 238*time.Millisecond { // more than 250 bpm
				continue // invalid
			}

			beatCh <- beatPkg{
				span: float64(t.Milliseconds()),
			}
			break
		}
	}(ctx)

	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("max3010x: could not get heart rate: %w", ErrTooNoisy)

	case b := <-beatCh:
		if errors.Is(b.err, errLowValue) {
			d.hr.reset()
			return 0, fmt.Errorf("max3010x: could not get heart rate: %w", ErrNotDetected)
		} else if b.err != nil {
			d.hr.reset()
			return 0, fmt.Errorf("max3010x: could not get heart rate: %w", b.err)
		}

		// if first measurement, pre-fill values.
		if d.hr.mean == 0 {
			d.hr.mean = b.span
		}

		d.hr.add(b.span)
	}

	return 60000 / d.hr.mean, nil
}

func (d *Device) detectBeat(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := d.ledsSingle()
		if err != nil {
			return fmt.Errorf("detectBeat: %w", err)
		}
		r := d.redLED.last()
		if r < threshold {
			return errLowValue
		}
		if d.beat.check(r) {
			break
		}
	}

	return nil
}
