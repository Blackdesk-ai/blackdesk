package application

import (
	"fmt"
	"strings"

	"blackdesk/internal/domain"
)

type ScreenerEntryInput struct {
	Available   bool
	WasActive   bool
	HasItems    bool
	CurrentName string
}

type ScreenerEntryResult struct {
	MarkLoaded  bool
	ShouldLoad  bool
	Status      string
	ApplyStatus bool
}

func PlanScreenerEntry(input ScreenerEntryInput) ScreenerEntryResult {
	if !input.Available {
		return ScreenerEntryResult{
			Status:      "Screeners unavailable for active provider",
			ApplyStatus: true,
		}
	}
	result := ScreenerEntryResult{MarkLoaded: true}
	if input.HasItems && input.WasActive {
		return result
	}
	if input.HasItems && !input.WasActive {
		result.Status = "Screener: " + strings.TrimSpace(input.CurrentName)
		result.ApplyStatus = true
		return result
	}
	result.Status = "Loading screener…"
	result.ApplyStatus = true
	result.ShouldLoad = true
	return result
}

type ScreenerAdvanceInput struct {
	Available    bool
	Definitions  []domain.ScreenerDefinition
	CurrentIndex int
	Step         int
}

type ScreenerAdvanceResult struct {
	NextIndex   int
	Definition  domain.ScreenerDefinition
	ShouldLoad  bool
	Status      string
	ApplyStatus bool
}

func PlanScreenerAdvance(input ScreenerAdvanceInput) ScreenerAdvanceResult {
	result := ScreenerAdvanceResult{NextIndex: input.CurrentIndex}
	if !input.Available || len(input.Definitions) == 0 {
		return result
	}
	nextIndex := (input.CurrentIndex + input.Step + len(input.Definitions)) % len(input.Definitions)
	definition := input.Definitions[nextIndex]
	result.NextIndex = nextIndex
	result.Definition = definition
	result.ShouldLoad = strings.TrimSpace(definition.ID) != ""
	if result.ShouldLoad {
		result.Status = "Loading screener: " + definition.Name
		result.ApplyStatus = true
	}
	return result
}

type ScreenerLoadInput struct {
	CurrentResult domain.ScreenerResult
	SelectedIndex int
	UserTriggered bool
	Data          domain.ScreenerResult
	Err           error
}

type ScreenerLoadResult struct {
	Result        domain.ScreenerResult
	SelectedIndex int
	Loaded        bool
	Status        string
}

func ApplyScreenerLoad(input ScreenerLoadInput) ScreenerLoadResult {
	result := ScreenerLoadResult{
		Result:        input.CurrentResult,
		SelectedIndex: input.SelectedIndex,
		Loaded:        input.UserTriggered,
	}
	if input.Err == nil {
		result.Result = input.Data
		if result.SelectedIndex >= len(result.Result.Items) {
			result.SelectedIndex = 0
		}
		if len(result.Result.Items) > 0 {
			result.Status = fmt.Sprintf("Loaded %s (%d names)", result.Result.Definition.Name, len(result.Result.Items))
		} else {
			result.Status = fmt.Sprintf("%s returned no names", result.Result.Definition.Name)
		}
		return result
	}
	if strings.TrimSpace(input.CurrentResult.Definition.Name) != "" {
		result.Status = "Screener unavailable: " + input.CurrentResult.Definition.Name
		return result
	}
	result.Status = "Screener unavailable"
	return result
}
