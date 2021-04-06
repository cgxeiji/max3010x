package max3010x

type tSeries struct {
	buffer []float64
	idx    int

	max float64
	min float64
}

func newTSeries(size int) *tSeries {
	return &tSeries{
		buffer: make([]float64, size),
	}
}

func (t *tSeries) add(entries ...float64) {
	for _, e := range entries {
		t.idx++
		t.idx %= len(t.buffer)

		old := t.buffer[t.idx]
		t.buffer[t.idx] = e

		if old == t.max || old == t.min {
			t.max = e
			t.min = e
			for _, b := range t.buffer {
				t.minmax(b)
			}
		} else {
			t.minmax(e)
		}
	}

}

func (t *tSeries) minmax(v float64) {
	if v > t.max {
		t.max = v
	}
	if v < t.min {
		t.min = v
	}
}

func (t *tSeries) last() float64 {
	return t.buffer[t.idx]
}

func (t *tSeries) acdc() float64 {
	if t.min == 0 {
		return 0
	}

	return (t.max - t.min) / t.min
}
