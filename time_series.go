package max3010x

type tSeries struct {
	data   []float64
	sum    float64
	mean   float64
	idxD   int
	idxS   int
	smooth []float64
}

func (t *tSeries) init(size, smooth int) {
	t.data = make([]float64, size)
	t.smooth = make([]float64, smooth)
	t.sum = 0
	t.mean = 0
	t.idxD = 0
	t.idxS = 0
}

func (t *tSeries) add(n float64) {
	t.sum -= t.smooth[t.idxS]
	t.sum += n
	t.smooth[t.idxS] = n
	t.idxS = (t.idxS + 1) % len(t.smooth)

	t.data[t.idxD] = n
	t.idxD = (t.idxD + 1) % len(t.data)

	t.mean = t.sum / float64(len(t.smooth))
}

func (t *tSeries) acdc() float64 {
	return acdc(t.data)
}

func acdc(data []float64) (ratio float64) {
	min := data[0]
	max := data[0]

	for _, d := range data {
		if d > max {
			max = d
		}
		if d < min {
			min = d
		}
	}

	return (max - min) / min
}
