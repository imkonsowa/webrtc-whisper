package audio

import "math"

type VAD struct {
	threshold    float64
	minSpeechMs  int
	minSilenceMs int
	sampleRate   int

	speechSamples  int
	silenceSamples int
	isSpeaking     bool
}

func NewVAD(sampleRate int) *VAD {
	return &VAD{
		threshold:    500,
		minSpeechMs:  100,
		minSilenceMs: 300,
		sampleRate:   sampleRate,
	}
}

func (v *VAD) Process(samples []int16) bool {
	energy := v.calculateEnergy(samples)

	minSpeechSamples := v.sampleRate * v.minSpeechMs / 1000
	minSilenceSamples := v.sampleRate * v.minSilenceMs / 1000

	if energy > v.threshold {
		v.speechSamples += len(samples)
		v.silenceSamples = 0

		if v.speechSamples >= minSpeechSamples {
			v.isSpeaking = true
		}
	} else {
		v.silenceSamples += len(samples)

		if v.isSpeaking && v.silenceSamples >= minSilenceSamples {
			v.isSpeaking = false
			v.speechSamples = 0
		}
	}

	return v.isSpeaking
}

func (v *VAD) calculateEnergy(samples []int16) float64 {
	if len(samples) == 0 {
		return 0
	}

	var sum float64
	for _, s := range samples {
		sum += float64(s) * float64(s)
	}

	return math.Sqrt(sum / float64(len(samples)))
}

func (v *VAD) Reset() {
	v.speechSamples = 0
	v.silenceSamples = 0
	v.isSpeaking = false
}
