package application

import (
	"context"
	"strings"

	"blackdesk/internal/domain"
)

type StatementRequest struct {
	Kind      domain.StatementKind
	Frequency domain.StatementFrequency
}

type PrepareAIContextRequest struct {
	Symbol            string
	MissingQuotes     []string
	NeedQuote         bool
	NeedHistory       bool
	NeedTechnical     bool
	NeedInsiders      bool
	NeedNews          bool
	NeedFundamentals  bool
	StatementRequests []StatementRequest
	SelectedStatement StatementRequest
	Range             string
	Interval          string
	TechnicalRange    string
	TechnicalInterval string
}

type PreparedAIContext struct {
	Quote           *domain.QuoteSnapshot
	QuoteErr        error
	Quotes          []domain.QuoteSnapshot
	QuotesErr       error
	History         *domain.PriceSeries
	HistoryErr      error
	Technical       *domain.PriceSeries
	TechnicalErr    error
	StatementBundle []domain.FinancialStatement
	Statement       *domain.FinancialStatement
	StatementLoaded bool
	StatementErr    error
	Insiders        *domain.InsiderSnapshot
	InsidersLoaded  bool
	InsidersErr     error
	News            []domain.NewsItem
	NewsLoaded      bool
	NewsErr         error
	Fundamentals    *domain.FundamentalsSnapshot
	FundErr         error
}

func (s *Services) PrepareAIContext(ctx context.Context, req PrepareAIContextRequest) PreparedAIContext {
	var out PreparedAIContext
	symbol := strings.TrimSpace(req.Symbol)
	if symbol == "" {
		return out
	}

	if req.NeedQuote {
		quote, err := s.GetQuote(ctx, symbol)
		if err == nil {
			out.Quote = &quote
		}
		out.QuoteErr = err
	}

	if len(req.MissingQuotes) > 0 {
		quotes, err := s.GetQuotes(ctx, req.MissingQuotes)
		out.Quotes = quotes
		out.QuotesErr = err
	}

	if req.NeedHistory {
		series, err := s.GetHistory(ctx, symbol, req.Range, req.Interval)
		if err == nil {
			out.History = &series
		}
		out.HistoryErr = err
	}

	if req.NeedTechnical {
		series, err := s.GetHistory(ctx, symbol, req.TechnicalRange, req.TechnicalInterval)
		if err == nil {
			out.Technical = &series
		}
		out.TechnicalErr = err
	}

	if len(req.StatementRequests) > 0 && s.HasStatements() {
		for _, statementReq := range req.StatementRequests {
			data, err := s.GetStatement(ctx, symbol, statementReq.Kind, statementReq.Frequency)
			if statementReq.Kind == req.SelectedStatement.Kind && statementReq.Frequency == req.SelectedStatement.Frequency {
				out.StatementLoaded = true
				if err == nil {
					stmt := data
					out.Statement = &stmt
				}
				out.StatementErr = err
			}
			if err == nil {
				out.StatementBundle = append(out.StatementBundle, data)
			}
		}
	}

	if req.NeedInsiders && s.HasInsiders() {
		data, err := s.GetInsiders(ctx, symbol)
		out.InsidersLoaded = true
		if err == nil {
			out.Insiders = &data
		}
		out.InsidersErr = err
	}

	if req.NeedNews {
		items, err := s.GetNews(ctx, symbol)
		out.News = items
		out.NewsLoaded = true
		out.NewsErr = err
	}

	if req.NeedFundamentals {
		data, err := s.GetFundamentals(ctx, symbol)
		if err == nil {
			out.Fundamentals = &data
		}
		out.FundErr = err
	}

	return out
}
