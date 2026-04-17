package tui

func technicalStatRows(s technicalSnapshot) []marketTableRow {
	priceZValue := formatMetricSigned(s.priceZ21)
	priceZSignal := priceZLabel(s.priceZ21)
	if priceZValue != "-" {
		priceZValue = rarityStyle(s.priceZ21).Render(priceZValue)
	}
	if priceZSignal != "-" {
		priceZSignal = rarityStyle(s.priceZ21).Render(priceZSignal)
	}
	rarityValue := formatProbability(s.priceZTail)
	raritySignal := rarityLabel(s.priceZ21)
	if rarityValue != "-" {
		rarityValue = rarityStyle(s.priceZ21).Render(rarityValue)
	}
	if raritySignal != "-" {
		raritySignal = rarityStyle(s.priceZ21).Render(raritySignal)
	}
	return []marketTableRow{
		{name: "PriceZ 21", price: priceZValue, chg: priceZSignal, move: 0, styled: false},
		{name: "ROC/HV", price: formatMetricSigned(s.ret12MOverHV), chg: "12M", move: s.ret12MOverHV, styled: s.ret12MOverHV != 0},
	}
}
