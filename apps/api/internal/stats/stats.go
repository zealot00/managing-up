package stats

import (
	"math"
)

func ComputeMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func ComputeVariance(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sumSq := 0.0
	for _, v := range values {
		diff := v - mean
		sumSq += diff * diff
	}
	return sumSq / float64(len(values))
}

func ComputeStdDev(values []float64, mean float64) float64 {
	return math.Sqrt(ComputeVariance(values, mean))
}

func TDistCDF(t, df float64) float64 {
	x := df / (df + t*t)
	return 1 - 0.5*math.Pow(x, df/2)
}

func TInv(p, df float64) float64 {
	if df <= 0 {
		return 1.96
	}
	switch {
	case df == 1:
		return 12.706
	case df == 2:
		return 4.303
	case df == 3:
		return 3.182
	case df == 4:
		return 2.776
	case df == 5:
		return 2.571
	case df == 10:
		return 2.228
	case df == 20:
		return 2.086
	case df >= 30:
		return 1.96
	default:
		return 1.96 + (2.576-1.96)*math.Exp(-0.05*df)
	}
}

func Round3(f float64) float64 {
	return float64(int(f*1000+0.5)) / 1000
}

func Round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}

// ComputePValue performs a two-sample t-test and returns the p-value.
func ComputePValue(sampleA, sampleB []float64) float64 {
	if len(sampleA) < 2 || len(sampleB) < 2 {
		return 1.0
	}

	meanA := ComputeMean(sampleA)
	meanB := ComputeMean(sampleB)
	varA := ComputeVariance(sampleA, meanA)
	varB := ComputeVariance(sampleB, meanB)

	nA := float64(len(sampleA))
	nB := float64(len(sampleB))

	if nA == 0 || nB == 0 {
		return 1.0
	}

	se := math.Sqrt(varA/nA + varB/nB)
	if se == 0 {
		return 1.0
	}

	tStat := (meanA - meanB) / se

	df := nA + nB - 2
	if df <= 0 {
		df = 1
	}

	pValue := 2 * (1 - TDistCDF(math.Abs(tStat), df))
	if pValue < 0 {
		pValue = 0
	}
	if pValue > 1 {
		pValue = 1
	}

	return pValue
}

// ComputeConfidenceInterval computes the 95% CI for the difference between two samples.
func ComputeConfidenceInterval(sampleA, sampleB []float64) [2]float64 {
	if len(sampleA) < 2 || len(sampleB) < 2 {
		return [2]float64{0, 0}
	}

	meanA := ComputeMean(sampleA)
	meanB := ComputeMean(sampleB)
	varA := ComputeVariance(sampleA, meanA)
	varB := ComputeVariance(sampleB, meanB)

	nA := float64(len(sampleA))
	nB := float64(len(sampleB))

	if nA == 0 || nB == 0 {
		return [2]float64{0, 0}
	}

	se := math.Sqrt(varA/nA + varB/nB)
	if se == 0 {
		diff := meanA - meanB
		return [2]float64{diff, diff}
	}

	tCrit := TInv(0.975, nA+nB-2)
	margin := tCrit * se
	diff := meanA - meanB

	return [2]float64{Round2(diff - margin), Round2(diff + margin)}
}
