package tui

func aiStatRowGuide() map[string]string {
	return map[string]string{
		"name":   "Human-readable metric label exactly as shown in Blackdesk.",
		"value":  "Metric value formatted for display in Blackdesk.",
		"signal": "Optional qualitative tag, comparison label, or regime hint associated with the metric.",
	}
}

func aiRowsFromMarketRows(rows []marketTableRow) []aiStatRow {
	items := make([]aiStatRow, 0, len(rows))
	for _, row := range rows {
		items = append(items, aiStatRow{Name: row.name, Value: row.price, Signal: row.chg})
	}
	return items
}
