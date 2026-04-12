package application

import "testing"

func TestPlanQuoteCenterSelectionHonorsCapabilities(t *testing.T) {
	statements := PlanQuoteCenterSelection(QuoteCenterSelectionInput{
		Target:        QuoteCenterStatements,
		HasStatements: true,
	})
	if !statements.Allowed || !statements.LoadStatement || statements.Status != "Quote center: statements" {
		t.Fatalf("unexpected statements center result: %+v", statements)
	}

	insiders := PlanQuoteCenterSelection(QuoteCenterSelectionInput{
		Target:      QuoteCenterInsiders,
		HasInsiders: false,
	})
	if insiders.Allowed || insiders.Status != "Insiders unavailable for active provider" {
		t.Fatalf("unexpected insiders center result: %+v", insiders)
	}

	filings := PlanQuoteCenterSelection(QuoteCenterSelectionInput{
		Target:     QuoteCenterFilings,
		HasFilings: true,
	})
	if !filings.Allowed || filings.Status != "Quote center: filings" {
		t.Fatalf("unexpected filings center result: %+v", filings)
	}
}
