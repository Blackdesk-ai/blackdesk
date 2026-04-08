package tui

import (
	"strings"

	"blackdesk/internal/application"
)

func (m Model) buildPrepareAIContextRequest(symbol string, missingQuotes []string) application.PrepareAIContextRequest {
	needQuote := !strings.EqualFold(m.quote.Symbol, symbol) || m.errQuote != nil
	needHistory := !strings.EqualFold(m.series.Symbol, symbol) || len(m.series.Candles) == 0 || m.errHistory != nil
	needTechnical := m.needsTechnicalHistory(symbol)
	needInsiders := m.needsInsiders(symbol)
	needNews := m.news == nil || m.errNews != nil
	needFundamentals := !strings.EqualFold(m.fundamentals.Symbol, symbol) || aiFundamentalsMissing(m.fundamentals) || m.errFundamentals != nil
	missingStatements := m.missingAIStatements(symbol)
	current := ranges[m.rangeIdx]

	return application.PrepareAIContextRequest{
		Symbol:            symbol,
		MissingQuotes:     missingQuotes,
		NeedQuote:         needQuote,
		NeedHistory:       needHistory,
		NeedTechnical:     needTechnical,
		NeedInsiders:      needInsiders,
		NeedNews:          needNews,
		NeedFundamentals:  needFundamentals,
		StatementRequests: toApplicationStatementRequests(missingStatements),
		SelectedStatement: application.StatementRequest{Kind: m.statementKind, Frequency: m.statementFreq},
		Range:             current.Range,
		Interval:          current.Interval,
		TechnicalRange:    "2y",
		TechnicalInterval: "1d",
	}
}

func toApplicationStatementRequests(items []statementRequest) []application.StatementRequest {
	if len(items) == 0 {
		return nil
	}
	out := make([]application.StatementRequest, 0, len(items))
	for _, item := range items {
		out = append(out, application.StatementRequest{
			Kind:      item.kind,
			Frequency: item.frequency,
		})
	}
	return out
}

func aiContextPreparedMsgFromResult(prompt, symbol string, result application.PreparedAIContext) aiContextPreparedMsg {
	return aiContextPreparedMsg{
		prompt:          prompt,
		symbol:          symbol,
		quote:           result.Quote,
		quoteErr:        result.QuoteErr,
		quotes:          result.Quotes,
		quotesErr:       result.QuotesErr,
		history:         result.History,
		historyErr:      result.HistoryErr,
		technical:       result.Technical,
		technicalErr:    result.TechnicalErr,
		statementBundle: result.StatementBundle,
		statement:       result.Statement,
		statementLoaded: result.StatementLoaded,
		statementErr:    result.StatementErr,
		insiders:        result.Insiders,
		insidersLoaded:  result.InsidersLoaded,
		insidersErr:     result.InsidersErr,
		news:            result.News,
		newsLoaded:      result.NewsLoaded,
		newsErr:         result.NewsErr,
		fundamentals:    result.Fundamentals,
		fundErr:         result.FundErr,
	}
}

func aiQuoteInsightPreparedMsgFromResult(symbol string, result application.PreparedAIContext) aiQuoteInsightPreparedMsg {
	return aiQuoteInsightPreparedMsg{
		symbol:          symbol,
		quote:           result.Quote,
		quoteErr:        result.QuoteErr,
		history:         result.History,
		historyErr:      result.HistoryErr,
		technical:       result.Technical,
		technicalErr:    result.TechnicalErr,
		statementBundle: result.StatementBundle,
		statement:       result.Statement,
		statementLoaded: result.StatementLoaded,
		statementErr:    result.StatementErr,
		insiders:        result.Insiders,
		insidersLoaded:  result.InsidersLoaded,
		insidersErr:     result.InsidersErr,
		news:            result.News,
		newsLoaded:      result.NewsLoaded,
		newsErr:         result.NewsErr,
		fundamentals:    result.Fundamentals,
		fundErr:         result.FundErr,
	}
}
