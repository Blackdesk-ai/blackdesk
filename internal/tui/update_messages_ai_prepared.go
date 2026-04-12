package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/domain"
)

type aiPreparedLoadInput struct {
	prompt          string
	symbol          string
	marketRisk      domain.MarketRiskSnapshot
	marketRiskErr   error
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
		marketRisk:      msg.marketRisk,
		marketRiskErr:   msg.marketRiskErr,
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
		marketRisk:      msg.marketRisk,
		marketRiskErr:   msg.marketRiskErr,
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

func (m Model) handleAIFilingAnalysisPrepared(msg aiFilingAnalysisPreparedMsg) (Model, tea.Cmd) {
	m.applyAIPreparedData(aiPreparedLoadInput{
		prompt:          msg.prompt,
		symbol:          msg.symbol,
		marketRisk:      msg.marketRisk,
		marketRiskErr:   msg.marketRiskErr,
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
	if msg.filingErr != nil {
		m.aiRunning = false
		m.aiErr = msg.filingErr
		m.pushAIAssistantMessage("", msg.filingErr, 0)
		m.status = msg.filingErr.Error()
		return m, nil
	}
	snapshot := m.filingsForSymbol(msg.symbol)
	chunks := splitFilingTextChunks(msg.filing.Text, aiFilingChunkChars, aiFilingChunkOverlapChars)
	if len(chunks) <= 1 {
		m.clearAIFilingRun()
		m.status = "Running " + m.activeAIConnectorLabel() + " filing analysis…"
		return m, m.runFilingAnalysisCmd(msg.symbol, snapshot, msg.filing, msg.prompt)
	}
	m.aiFilingRun = aiFilingRunState{
		symbol:       msg.symbol,
		snapshot:     snapshot,
		filing:       msg.filing,
		prompt:       msg.prompt,
		chunks:       chunks,
		nextChunkIdx: 0,
		truncation:   aiRequestTruncation{FilingText: msg.filing.Truncated},
	}
	m.aiFilingRunActive = true
	m.status = fmt.Sprintf("Running %s filing chunk 1/%d…", m.activeAIConnectorLabel(), len(chunks))
	return m, m.runFilingChunkAnalysisCmd(msg.symbol, snapshot, msg.filing, chunks[0])
}

func (m Model) handleAIFilingChunkLoaded(msg aiFilingChunkLoadedMsg) (Model, tea.Cmd) {
	if !m.aiFilingRunActive {
		return m, nil
	}
	m.aiFilingRun.totalDuration += msg.duration
	m.aiFilingRun.truncation = m.aiFilingRun.truncation.merge(msg.truncation)
	if msg.err != nil {
		truncation := m.aiFilingRun.truncation
		duration := m.aiFilingRun.totalDuration
		m.clearAIFilingRun()
		return m.handleAIResponseLoaded(aiResponseLoadedMsg{
			connectorID: msg.connectorID,
			output:      msg.output,
			duration:    duration,
			truncation:  truncation,
			symbol:      msg.symbol,
			err:         msg.err,
		})
	}
	chunk := m.aiFilingRun.chunks[m.aiFilingRun.nextChunkIdx]
	m.aiFilingRun.analyses = append(m.aiFilingRun.analyses, filingChunkAnalysisSummary{
		ChunkIndex: chunk.Index,
		ChunkRange: filingChunkRangeLabel(chunk),
		Analysis:   strings.TrimSpace(msg.output),
	})
	m.aiFilingRun.nextChunkIdx++
	if m.aiFilingRun.nextChunkIdx < len(m.aiFilingRun.chunks) {
		next := m.aiFilingRun.chunks[m.aiFilingRun.nextChunkIdx]
		m.status = fmt.Sprintf("Running %s filing chunk %d/%d…", m.activeAIConnectorLabel(), next.Index, next.Total)
		return m, m.runFilingChunkAnalysisCmd(m.aiFilingRun.symbol, m.aiFilingRun.snapshot, m.aiFilingRun.filing, next)
	}
	m.aiFilingRun.synthesizing = true
	m.status = fmt.Sprintf("Synthesizing %s filing report from %d chunks…", m.activeAIConnectorLabel(), len(m.aiFilingRun.chunks))
	return m, m.runFilingSynthesisCmd(m.aiFilingRun.symbol, m.aiFilingRun.snapshot, m.aiFilingRun.filing, m.aiFilingRun.prompt, m.aiFilingRun.analyses)
}

func (m Model) handleAIFilingSynthesisLoaded(msg aiFilingSynthesisLoadedMsg) (Model, tea.Cmd) {
	if !m.aiFilingRunActive {
		return m.handleAIResponseLoaded(aiResponseLoadedMsg{
			connectorID: msg.connectorID,
			output:      msg.output,
			duration:    msg.duration,
			truncation:  msg.truncation,
			symbol:      msg.symbol,
			err:         msg.err,
		})
	}
	totalDuration := m.aiFilingRun.totalDuration + msg.duration
	truncation := m.aiFilingRun.truncation.merge(msg.truncation)
	m.clearAIFilingRun()
	return m.handleAIResponseLoaded(aiResponseLoadedMsg{
		connectorID: msg.connectorID,
		output:      msg.output,
		duration:    totalDuration,
		truncation:  truncation,
		symbol:      msg.symbol,
		err:         msg.err,
	})
}

func (m *Model) applyAIPreparedData(input aiPreparedLoadInput) {
	for _, stmt := range input.statementBundle {
		m.cacheStatement(stmt)
	}
	if input.marketRiskErr != nil {
		m.marketRisk = domain.MarketRiskSnapshot{}
	} else if input.marketRisk.Available {
		m.marketRisk = input.marketRisk
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
