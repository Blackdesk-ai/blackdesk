package application

type QuoteCenterSelectionInput struct {
	Target        QuoteCenterMode
	HasStatements bool
	HasInsiders   bool
	HasOwners     bool
	HasAnalyst    bool
	HasFilings    bool
}

type QuoteCenterSelectionResult struct {
	Allowed       bool
	Status        string
	LoadTechnical bool
	LoadStatement bool
	LoadInsiders  bool
	LoadOwners    bool
	LoadAnalyst   bool
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
	case QuoteCenterOwners:
		if !input.HasOwners {
			return QuoteCenterSelectionResult{Status: "Owners unavailable for active provider"}
		}
		return QuoteCenterSelectionResult{
			Allowed:    true,
			Status:     "Quote center: owners",
			LoadOwners: true,
		}
	case QuoteCenterAnalyst:
		if !input.HasAnalyst {
			return QuoteCenterSelectionResult{Status: "Analyst recommendations unavailable for active provider"}
		}
		return QuoteCenterSelectionResult{
			Allowed:     true,
			Status:      "Quote center: analyst recommendations",
			LoadAnalyst: true,
		}
	case QuoteCenterTechnicals:
		return QuoteCenterSelectionResult{
			Allowed:       true,
			Status:        "Quote center: technicals",
			LoadTechnical: true,
		}
	case QuoteCenterFilings:
		if !input.HasFilings {
			return QuoteCenterSelectionResult{Status: "Filings unavailable"}
		}
		return QuoteCenterSelectionResult{
			Allowed: true,
			Status:  "Quote center: filings",
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
