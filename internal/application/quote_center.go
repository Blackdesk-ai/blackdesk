package application

type QuoteCenterSelectionInput struct {
	Target        QuoteCenterMode
	HasStatements bool
	HasInsiders   bool
}

type QuoteCenterSelectionResult struct {
	Allowed       bool
	Status        string
	LoadTechnical bool
	LoadStatement bool
	LoadInsiders  bool
}

func PlanQuoteCenterSelection(input QuoteCenterSelectionInput) QuoteCenterSelectionResult {
	switch input.Target {
	case QuoteCenterStatements:
		if !input.HasStatements {
			return QuoteCenterSelectionResult{Status: "Statements unavailable for active provider"}
		}
		return QuoteCenterSelectionResult{
			Allowed:       true,
			Status:        "Quote center: statements",
			LoadStatement: true,
		}
	case QuoteCenterInsiders:
		if !input.HasInsiders {
			return QuoteCenterSelectionResult{Status: "Insiders unavailable for active provider"}
		}
		return QuoteCenterSelectionResult{
			Allowed:      true,
			Status:       "Quote center: insiders",
			LoadInsiders: true,
		}
	case QuoteCenterTechnicals:
		return QuoteCenterSelectionResult{
			Allowed:       true,
			Status:        "Quote center: technicals",
			LoadTechnical: true,
		}
	case QuoteCenterFundamentals:
		return QuoteCenterSelectionResult{
			Allowed: true,
			Status:  "Quote center: fundamentals",
		}
	default:
		return QuoteCenterSelectionResult{
			Allowed: true,
			Status:  "Quote center: chart",
		}
	}
}
