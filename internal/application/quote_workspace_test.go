package application

import "testing"

func TestPlanQuoteWorkspaceLoadEnablesCenterSpecificLoads(t *testing.T) {
	plan := PlanQuoteWorkspaceLoad(QuoteCenterStatements, true, true, true, true)

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
	plan := PlanQuoteWorkspaceLoad(QuoteCenterInsiders, true, false, false, false)

	if plan.LoadInsiders || plan.LoadStatement {
		t.Fatalf("expected unavailable center loads to stay disabled, got %+v", plan)
	}
}

func TestPlanQuoteWorkspaceLoadEnablesFilingsLoad(t *testing.T) {
	plan := PlanQuoteWorkspaceLoad(QuoteCenterFilings, false, true, true, true)

	if !plan.LoadFilings {
		t.Fatalf("expected filings load for filings mode, got %+v", plan)
	}
	if plan.LoadStatement || plan.LoadInsiders || plan.LoadTechnical {
		t.Fatalf("expected unrelated center loads to stay disabled, got %+v", plan)
	}
}

func TestPlanQuoteWorkspaceLoadEnablesAnalystLoad(t *testing.T) {
	plan := PlanQuoteWorkspaceLoad(QuoteCenterAnalyst, false, true, true, true)

	if !plan.LoadAnalyst {
		t.Fatalf("expected analyst load for analyst mode, got %+v", plan)
	}
	if plan.LoadStatement || plan.LoadInsiders || plan.LoadTechnical {
		t.Fatalf("expected unrelated center loads to stay disabled, got %+v", plan)
	}
}

func TestPlanQuoteWorkspaceLoadEnablesOwnersLoad(t *testing.T) {
	plan := PlanQuoteWorkspaceLoad(QuoteCenterOwners, false, true, true, true)

	if !plan.LoadOwners {
		t.Fatalf("expected owners load for owners mode, got %+v", plan)
	}
	if plan.LoadStatement || plan.LoadInsiders || plan.LoadTechnical {
		t.Fatalf("expected unrelated center loads to stay disabled, got %+v", plan)
	}
}
