package tui

import (
	"math"

	"blackdesk/internal/domain"
)

type technicalSnapshot struct {
	price         float64
	lastVolume    int64
	closeCount    int
	sma20         float64
	sma50         float64
	sma200        float64
	ema12         float64
	ema26         float64
	adx14         float64
	atr14         float64
	macd          float64
	macdSignal    float64
	macdHist      float64
	rsi14         float64
	ret1M         float64
	ret3M         float64
	ret12M        float64
	hv21          float64
	hv63          float64
	hv252         float64
	hvRank21      float64
	hvRankReady   bool
	hvPct21       float64
	hvPctReady    bool
	avgVolume20   float64
	volumeRatio20 float64
	priceZ21      float64
	priceZTail    float64
	priceZSamples int
	ret1MOverHV   float64
	ret3MOverHV   float64
	ret12MOverHV  float64
}

func buildTechnicalSnapshot(quote domain.QuoteSnapshot, series domain.PriceSeries) technicalSnapshot {
	closes := extractCloses(series.Candles)
	price := quote.Price
	if price == 0 && len(closes) > 0 {
		price = closes[len(closes)-1]
	}
	lastVolume := quote.Volume
	if lastVolume == 0 && len(series.Candles) > 0 {
		lastVolume = series.Candles[len(series.Candles)-1].Volume
	}
	avgVolume20 := averageVolume(series.Candles, 20)
	volumeRatio20 := 0.0
	if avgVolume20 > 0 && lastVolume > 0 {
		volumeRatio20 = float64(lastVolume) / avgVolume20
	}
	adx14 := averageDirectionalIndex(series.Candles, 14)
	atr14 := averageTrueRange(series.Candles, 14)
	macd, signal, hist := macd(closes)
	priceZ21, priceZTail, priceZSamples := rollingReturnZScore(closes, 21, 252)
	hv21 := historicalVolatility(closes, 21)
	hvRank21, hvRankReady := historicalVolatilityRank(closes, 21, 252)
	hvPct21, hvPctReady := historicalVolatilityPercentile(closes, 21, 252)
	return technicalSnapshot{
		price:         price,
		lastVolume:    lastVolume,
		closeCount:    len(closes),
		sma20:         simpleMovingAverage(closes, 20),
		sma50:         simpleMovingAverage(closes, 50),
		sma200:        simpleMovingAverage(closes, 200),
		ema12:         exponentialMovingAverage(closes, 12),
		ema26:         exponentialMovingAverage(closes, 26),
		adx14:         adx14,
		atr14:         atr14,
		macd:          macd,
		macdSignal:    signal,
		macdHist:      hist,
		rsi14:         rsi(closes, 14),
		ret1M:         lookbackReturn(closes, 21),
		ret3M:         lookbackReturn(closes, 63),
		ret12M:        lookbackReturn(closes, 252),
		hv21:          hv21,
		hv63:          historicalVolatility(closes, 63),
		hv252:         historicalVolatility(closes, 252),
		hvRank21:      hvRank21,
		hvRankReady:   hvRankReady,
		hvPct21:       hvPct21,
		hvPctReady:    hvPctReady,
		avgVolume20:   avgVolume20,
		volumeRatio20: volumeRatio20,
		priceZ21:      priceZ21,
		priceZTail:    priceZTail,
		priceZSamples: priceZSamples,
		ret1MOverHV:   returnOverVol(lookbackReturn(closes, 21), hv21, math.Sqrt(12)),
		ret3MOverHV:   returnOverVol(lookbackReturn(closes, 63), hv21, 2),
		ret12MOverHV:  returnOverVol(lookbackReturn(closes, 252), historicalVolatility(closes, 252), 1),
	}
}
