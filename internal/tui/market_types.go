package tui

type marketSectionBlock struct {
	title      string
	valueLabel string
	items      []marketBoardItem
}

type marketTableRow struct {
	name   string
	price  string
	chg    string
	move   float64
	styled bool
}
