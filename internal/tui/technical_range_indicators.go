package tui

import (
	"math"

	"blackdesk/internal/domain"
)

func averageTrueRange(candles []domain.Candle, period int) float64 {
	if period <= 0 || len(candles) <= period {
		return 0
	}
	trueRanges := make([]float64, 0, len(candles)-1)
	for i := 1; i < len(candles); i++ {
		tr := trueRange(candles[i], candles[i-1].Close)
		if tr > 0 {
			trueRanges = append(trueRanges, tr)
		}
	}
	if len(trueRanges) < period {
		return 0
	}
	atr := 0.0
	for _, tr := range trueRanges[:period] {
		atr += tr
	}
	atr /= float64(period)
	for _, tr := range trueRanges[period:] {
		atr = ((atr * float64(period-1)) + tr) / float64(period)
	}
	return atr
}

func averageDirectionalIndex(candles []domain.Candle, period int) float64 {
	if period <= 0 || len(candles) <= period*2 {
		return 0
	}
	trueRanges := make([]float64, 0, len(candles)-1)
	plusDM := make([]float64, 0, len(candles)-1)
	minusDM := make([]float64, 0, len(candles)-1)
	for i := 1; i < len(candles); i++ {
		curr := candles[i]
		prev := candles[i-1]
		upMove := curr.High - prev.High
		downMove := prev.Low - curr.Low
		plus := 0.0
		minus := 0.0
		if upMove > downMove && upMove > 0 {
			plus = upMove
		}
		if downMove > upMove && downMove > 0 {
			minus = downMove
		}
		trueRanges = append(trueRanges, trueRange(curr, prev.Close))
		plusDM = append(plusDM, plus)
		minusDM = append(minusDM, minus)
	}
	if len(trueRanges) < period*2 {
		return 0
	}
	trN := sumFloatSlice(trueRanges[:period])
	plusN := sumFloatSlice(plusDM[:period])
	minusN := sumFloatSlice(minusDM[:period])
	dxSeries := make([]float64, 0, len(trueRanges)-period+1)
	for i := period; i <= len(trueRanges); i++ {
		if i > period {
			trN = trN - (trN / float64(period)) + trueRanges[i-1]
			plusN = plusN - (plusN / float64(period)) + plusDM[i-1]
			minusN = minusN - (minusN / float64(period)) + minusDM[i-1]
		}
		if trN == 0 {
			dxSeries = append(dxSeries, 0)
			continue
		}
		plusDI := 100 * (plusN / trN)
		minusDI := 100 * (minusN / trN)
		sumDI := plusDI + minusDI
		if sumDI == 0 {
			dxSeries = append(dxSeries, 0)
			continue
		}
		dxSeries = append(dxSeries, 100*math.Abs(plusDI-minusDI)/sumDI)
	}
	if len(dxSeries) < period {
		return 0
	}
	adx := sumFloatSlice(dxSeries[:period]) / float64(period)
	for _, dx := range dxSeries[period:] {
		adx = ((adx * float64(period-1)) + dx) / float64(period)
	}
	return adx
}

func trueRange(candle domain.Candle, prevClose float64) float64 {
	rangeHighLow := candle.High - candle.Low
	if prevClose == 0 {
		return math.Max(rangeHighLow, 0)
	}
	rangeHighClose := math.Abs(candle.High - prevClose)
	rangeLowClose := math.Abs(candle.Low - prevClose)
	return math.Max(rangeHighLow, math.Max(rangeHighClose, rangeLowClose))
}

func sumFloatSlice(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum
}
