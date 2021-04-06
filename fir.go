package max3010x

var firC = []float64{21.5, 40.125, 72.375, 115.875, 170.0, 232.25, 298.75, 364.5, 423.875, 471.0, 501.5, 512.0}

const firSize = 32

type fir struct {
	buffer []float64
	idx    int
}

func newFIR() *fir {
	return &fir{
		buffer: make([]float64, firSize),
	}
}

// lowPass applies a low pass FIR filter to a delta.
func (f *fir) lowPass(delta float64) float64 {
	f.buffer[f.idx] = delta

	z := firC[11] * f.buffer[(f.idx-11)&0x1F]

	for i := 0; i < 11; i++ {
		z += firC[i] * (f.buffer[(f.idx-i)&0x1F] + f.buffer[(f.idx-(firSize-10)+i)&0x1F])
	}

	f.idx++
	f.idx %= firSize

	return z
}
