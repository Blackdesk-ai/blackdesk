package tui

import "blackdesk/internal/domain"

func simpleMovingAverage(values []float64, period int) float64 {
	if period <= 0 || len(values) < period {
		return 0
	}
	sum := 0.0
	for _, v := range values[len(values)-period:] {
		sum += v
	}
	return sum / float64(period)
}

func exponentialMovingAverage(values []float64, period int) float64 {
	if period <= 0 || len(values) < period {
		return 0
	}
	ema := simpleMovingAverage(values[:period], period)
	multiplier := 2.0 / float64(period+1)
	for _, v := range values[period:] {
		ema = ((v - ema) * multiplier) + ema
	}
	return ema
}

func macd(values []float64) (line, signal, hist float64) {
	if len(values) < 26 {
		return 0, 0, 0
	}
	macdSeries := make([]float64, 0, len(values)-25)
	for i := 26; i <= len(values); i++ {
		window := values[:i]
		macdSeries = append(macdSeries, exponentialMovingAverage(window, 12)-exponentialMovingAverage(window, 26))
	}
	if len(macdSeries) == 0 {
		return 0, 0, 0
	}
	line = macdSeries[len(macdSeries)-1]
	signal = exponentialMovingAverage(macdSeries, 9)
	if signal == 0 && len(macdSeries) < 9 {
		return line, 0, 0
	}
	hist = line - signal
	return line, signal, hist
}

func rsi(values []float64, period int) float64 {
	if period <= 0 || len(values) <= period {
		return 0
	}
	gains := 0.0
	losses := 0.0
	for i := 1; i <= period; i++ {
		change := values[i] - values[i-1]
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}
	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)
	for i := period + 1; i < len(values); i++ {
		change := values[i] - values[i-1]
		gain := 0.0
		loss := 0.0
		if change > 0 {
			gain = change
		} else {
			loss = -change
		}
		avgGain = ((avgGain * float64(period-1)) + gain) / float64(period)
		avgLoss = ((avgLoss * float64(period-1)) + loss) / float64(period)
	}
	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

func lookbackReturn(values []float64, periods int) float64 {
	if periods <= 0 || len(values) <= periods {
		return 0
	}
	base := values[len(values)-1-periods]
	last := values[len(values)-1]
	if base == 0 {
		return 0
	}
	return (last - base) / base
}

func averageVolume(candles []domain.Candle, period int) float64 {
	if period <= 0 || len(candles) < period {
		return 0
	}
	sum := int64(0)
	for _, candle := range candles[len(candles)-period:] {
		sum += candle.Volume
	}
	return float64(sum) / float64(period)
}
