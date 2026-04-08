package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/domain"
)

type aiPreparedLoadInput struct {
	prompt          string
	symbol          string
	quote           *domain.QuoteSnapshot
	quoteErr        error
	quotes          []domain.QuoteSnapshot
	quotesErr       error
	history         *domain.PriceSeries
	historyErr      error
	technical       *domain.PriceSeries
	technicalErr    error
	statementBundle []domain.FinancialStatement
	statement       *domain.FinancialStatement
	statementLoaded bool
	statementErr    error
	insiders        *domain.InsiderSnapshot
	insidersLoaded  bool
	insidersErr     error
	news            []domain.NewsItem
	newsLoaded      bool
	newsErr         error
	fundamentals    *domain.FundamentalsSnapshot
	fundErr         error
}

func (m Model) handleAIQuoteInsightPrepared(msg aiQuoteInsightPreparedMsg) (Model, tea.Cmd) {
	m.applyAIPreparedData(aiPreparedLoadInput{
		symbol:          msg.symbol,
		quote:           msg.quote,
		quoteErr:        msg.quoteErr,
		history:         msg.history,
		historyErr:      msg.historyErr,
		technical:       msg.technical,
		technicalErr:    msg.technicalErr,
		statementBundle: msg.statementBundle,
		statement:       msg.statement,
		statementLoaded: msg.statementLoaded,
		statementErr:    msg.statementErr,
		insiders:        msg.insiders,
		insidersLoaded:  msg.insidersLoaded,
		insidersErr:     msg.insidersErr,
		news:            msg.news,
		newsLoaded:      msg.newsLoaded,
		newsErr:         msg.newsErr,
		fundamentals:    msg.fundamentals,
		fundErr:         msg.fundErr,
	})
	m.status = "Running " + m.activeAIConnectorLabel() + " quote insight…"
	return m, m.runQuoteInsightCmd(msg.symbol)
}

func (m Model) handleAIContextPrepared(msg aiContextPreparedMsg) (Model, tea.Cmd) {
	m.applyAIPreparedData(aiPreparedLoadInput{
		prompt:          msg.prompt,
		symbol:          msg.symbol,
		quote:           msg.quote,
		quoteErr:        msg.quoteErr,
		quotes:          msg.quotes,
		quotesErr:       msg.quotesErr,
		history:         msg.history,
		historyErr:      msg.historyErr,
		technical:       msg.technical,
		technicalErr:    msg.technicalErr,
		statementBundle: msg.statementBundle,
		statement:       msg.statement,
		statementLoaded: msg.statementLoaded,
		statementErr:    msg.statementErr,
		insiders:        msg.insiders,
		insidersLoaded:  msg.insidersLoaded,
		insidersErr:     msg.insidersErr,
		news:            msg.news,
		newsLoaded:      msg.newsLoaded,
		newsErr:         msg.newsErr,
		fundamentals:    msg.fundamentals,
		fundErr:         msg.fundErr,
	})
	m.status = "Running " + m.activeAIConnectorLabel() + "…"
	return m, m.runAICmd(msg.prompt)
}

func (m *Model) applyAIPreparedData(input aiPreparedLoadInput) {
	for _, stmt := range input.statementBundle {
		m.cacheStatement(stmt)
	}
	if input.quote != nil {
		m.quote = *input.quote
		m.errQuote = input.quoteErr
		if input.symbol != "" {
			m.watchQuotes[strings.ToUpper(input.symbol)] = *input.quote
		}
	}
	if input.quotesErr == nil {
		for _, quote := range input.quotes {
			if quote.Symbol == "" {
				continue
			}
			m.watchQuotes[strings.ToUpper(quote.Symbol)] = quote
		}
	}
	if input.history != nil {
		m.series = *input.history
		m.errHistory = input.historyErr
	}
	if input.technical != nil {
		m.technicalCache[strings.ToUpper(input.technical.Symbol)] = *input.technical
		m.errTechnicalHistory = input.technicalErr
	}
	if input.statementLoaded {
		if input.statement != nil {
			m.statement = *input.statement
		}
		m.errStatement = input.statementErr
	}
	if input.insidersLoaded {
		if input.insiders != nil {
			m.insiders = *input.insiders
			m.cacheInsiders(*input.insiders)
		}
		m.errInsiders = input.insidersErr
	}
	if input.newsLoaded {
		m.news = input.news
		m.errNews = input.newsErr
		if m.newsSelected >= len(m.news) {
			m.newsSelected = 0
		}
	}
	if input.fundamentals != nil {
		m.fundamentals = *input.fundamentals
		m.errFundamentals = input.fundErr
	}
}
