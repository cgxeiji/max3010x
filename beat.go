package max3010x

type beat struct {
	filterFIR *fir
	signal    struct {
		dc movingAverage
		ac struct {
			max  float64
			min  float64
			prev float64
		}
		rising bool
	}
}

func newBeat() *beat {
	return &beat{
		filterFIR: newFIR(),
	}
}

// check receives a normalized (0.0 - 1.0) signal input and checks for
// beats. It returns true on rising edges (positive zero crossings) or false
// otherwise.
func (b *beat) check(signal float64) bool {
	beat := false

	b.signal.dc.add(signal)
	ac := b.filterFIR.lowPass(signal - b.signal.dc.mean)

	// Rising edge
	if b.signal.ac.prev < 0 && ac >= 0 {
		delta := b.signal.ac.max - b.signal.ac.min
		if delta > 1 && delta < 50 {
			beat = true
		}

		b.signal.rising = true
		b.signal.ac.max = 0
	}

	// Falling edge
	if b.signal.ac.prev > 0 && ac <= 0 {
		b.signal.rising = false
		b.signal.ac.min = 0
	}

	if b.signal.rising {
		if ac > b.signal.ac.prev {
			b.signal.ac.max = ac
		}
	} else {
		if ac < b.signal.ac.prev {
			b.signal.ac.min = ac
		}
	}

	b.signal.ac.prev = ac

	return beat
}
