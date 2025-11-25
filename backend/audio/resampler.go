package audio

func Resample(input []int16, inputRate, outputRate int) []int16 {
	if inputRate == outputRate {
		return input
	}

	ratio := float64(inputRate) / float64(outputRate)
	outputLen := int(float64(len(input)) / ratio)
	output := make([]int16, outputLen)

	for i := 0; i < outputLen; i++ {
		srcIdx := float64(i) * ratio
		idx := int(srcIdx)
		frac := srcIdx - float64(idx)

		if idx+1 < len(input) {
			output[i] = int16(float64(input[idx])*(1-frac) + float64(input[idx+1])*frac)
		} else if idx < len(input) {
			output[i] = input[idx]
		}
	}

	return output
}

func ResampleBytes(input []byte, inputRate, outputRate int) []byte {
	samples := make([]int16, len(input)/2)
	for i := 0; i < len(samples); i++ {
		samples[i] = int16(input[i*2]) | int16(input[i*2+1])<<8
	}

	resampled := Resample(samples, inputRate, outputRate)

	output := make([]byte, len(resampled)*2)
	for i, s := range resampled {
		output[i*2] = byte(s)
		output[i*2+1] = byte(s >> 8)
	}

	return output
}
