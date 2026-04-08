package application

import "testing"

func TestPlanQuoteWorkspaceLoadEnablesCenterSpecificLoads(t *testing.T) {
	plan := PlanQuoteWorkspaceLoad(QuoteCenterStatements, true, true, true)

	if !plan.LoadQuote || !plan.LoadHistory || !plan.LoadNews || !plan.LoadFundamentals {
		t.Fatalf("expected base quote loads to be enabled, got %+v", plan)
	}
	if !plan.LoadStatement {
		t.Fatalf("expected statement load for statements mode, got %+v", plan)
	}
	if plan.LoadTechnical || plan.LoadInsiders {
		t.Fatalf("expected unrelated center loads to stay disabled, got %+v", plan)
	}
}

func TestPlanQuoteWorkspaceLoadSkipsUnavailableCenterLoads(t *testing.T) {
	plan := PlanQuoteWorkspaceLoad(QuoteCenterInsiders, true, false, false)

	if plan.LoadInsiders || plan.LoadStatement {
		t.Fatalf("expected unavailable center loads to stay disabled, got %+v", plan)
	}
}
