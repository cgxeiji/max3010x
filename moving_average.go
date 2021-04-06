package max3010x

// movingAverage stores an estimated moving average of the last 4 values.
type movingAverage struct {
	mean float64
}

func (m *movingAverage) add(n float64) {
	m.mean += (n - m.mean) / 4
}

func (m *movingAverage) reset() {
	m.mean = 0
}
