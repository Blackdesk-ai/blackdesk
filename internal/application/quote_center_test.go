package application

import "testing"

func TestPlanQuoteCenterSelectionHonorsCapabilities(t *testing.T) {
	fundamentals := PlanQuoteCenterSelection(QuoteCenterSelectionInput{
		Target: QuoteCenterFundamentals,
	})
	if !fundamentals.Allowed || !fundamentals.LoadTechnical || fundamentals.Status != "Quote center: fundamentals" {
		t.Fatalf("unexpected fundamentals center result: %+v", fundamentals)
	}

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

	owners := PlanQuoteCenterSelection(QuoteCenterSelectionInput{
		Target:    QuoteCenterOwners,
		HasOwners: true,
	})
	if !owners.Allowed || !owners.LoadOwners || owners.Status != "Quote center: owners" {
		t.Fatalf("unexpected owners center result: %+v", owners)
	}

	filings := PlanQuoteCenterSelection(QuoteCenterSelectionInput{
		Target:     QuoteCenterFilings,
		HasFilings: true,
	})
	if !filings.Allowed || filings.Status != "Quote center: filings" {
		t.Fatalf("unexpected filings center result: %+v", filings)
	}

	analyst := PlanQuoteCenterSelection(QuoteCenterSelectionInput{
		Target:     QuoteCenterAnalyst,
		HasAnalyst: true,
	})
	if !analyst.Allowed || !analyst.LoadAnalyst || analyst.Status != "Quote center: analyst recommendations" {
		t.Fatalf("unexpected analyst center result: %+v", analyst)
	}
}
