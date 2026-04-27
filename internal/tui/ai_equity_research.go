package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const aiEquityResearchPrompt = `Analyze this company as an investment opportunity with the goal of making a clear capital allocation decision.

Use the app context fully: market regime, quote, fundamentals, technicals, statements, news, and derived metrics.
Distinguish clearly between:
- Reported
- Normalized
- Market-implied
- Your estimate

0. Company Type (CRITICAL FIRST STEP)
- Classify the company as one of:
  • Mature / cash-generating
  • High-growth / reinvesting
  • Cyclical / commodity-driven
  • Financial (bank / insurer / asset manager / lender)
- Adjust all analysis accordingly.

1. Business Quality
- Evaluate ROIC, margins, and durability of the business model.
- Is this a high-quality compounder, cyclical business, financial franchise, or average company?
- Does the company have a durable competitive advantage?

2. Earnings Quality
- Are current earnings (EBITDA, FCF, margins, or financial-sector equivalents) sustainable?
- Are we at peak, normal, or depressed earnings?
- Explain what is driving any distortion (cycle, capex, one-offs, provisioning, rates, etc.).

3. Normalization
- Estimate normalized EBITDA and normalized free cash flow where applicable.
- For cyclicals: normalize across the cycle.
- For growth companies: focus on margin trajectory instead of penalizing low current FCF.
- For financials: normalize returns on equity / tangible equity, credit losses, underwriting margins, or spread earnings as appropriate.
- Recalculate EV/EBITDA and FCF yield using normalized values where applicable.

4. Capital Allocation Quality
- How has management historically allocated capital?
- Evaluate buybacks, dilution, stock-based comp, M&A, debt, dividends, and reinvestment returns.
- Has capital allocation added or destroyed per-share value?

5. Balance Sheet / Survival
- Evaluate leverage, liquidity, refinancing risk, and balance sheet resilience.
- For cyclicals, ask whether the company can survive a full downturn without permanent impairment.
- For financials, focus on capital adequacy, asset quality, reserve quality, and funding stability.

6. Valuation (ADAPTIVE)
- For mature companies:
  • Use earnings yield, FCF yield, EV/EBITDA
- For growth companies:
  • Focus on revenue growth, margin expansion, and long-term profitability
  • Do not penalize low FCF if reinvestment has high returns
- For financials:
  • Use ROTCE / ROE, P/TBV, capital quality, and earnings power
- Conclude: cheap / fair / expensive

7. Growth vs Expectations
- What growth or profitability is implied by the current price?
- Is the market too optimistic, too pessimistic, or roughly correct?
- What does the market believe today, and where is that belief likely wrong?

8. Per-Share Economics
- Focus on value creation per share, not just aggregate growth.
- Is revenue, earnings, and free cash flow growth translating into higher per-share value?

9. Implied Return
- Estimate long-term implied return:
  Implied Return ≈ Earnings Yield + N5Y Growth
- For growth companies: base this on normalized future margins.
- For financials: base this on sustainable returns on capital and valuation re-rating potential.
- Adjust for potential multiple expansion or compression where relevant.
- Provide a range:
  • Bear
  • Base
  • Bull

10. Risk
- Identify the 2–3 key risks that could break the thesis.

11. Alpha / Edge
- Compare implied return to a 10% baseline (index).
- Is this clearly better, similar, or worse?
- What is the market likely mispricing?

12. Decision (FIRST LEVEL)
- Give a clear verdict:
  BUY / HOLD / PASS
- Explain briefly why.
- Specify investor type (value, growth, defensive, cyclical, financial, etc.).

13. Price Sensitivity (CRITICAL)
- At what price does this become:
  • BUY
  • STRONG BUY
- Provide clear zones:
  • AVOID / HOLD / BUY / STRONG BUY

14. Capital Allocation Decision (FINAL)
- You must choose between this stock and a broad index (~10% return).
- Choose ONLY ONE:
  • BUY (clear alpha)
  • HOLD (keep if owned, not attractive for new capital)
  • PASS (no edge vs market)
- No mixed answers. Be decisive.
- If evidence quality is weak, still choose one, but explicitly say confidence is low and why.

15. Upside Trigger
- What must happen for this to deliver 15%+ returns?

16. Confidence
- State confidence level:
  • Low
  • Medium
  • High
- Briefly explain what would increase or reduce confidence.

17. Timing / Market Context (CRITICAL)
- Is the current market regime supportive (risk-on / risk-off)?
- Is the stock in an uptrend or downtrend based on the technical context?
- Is this the right time to enter, or should capital wait for a better setup?
- Use timing to refine entry and sizing, not to override core business quality.

18. Relative Opportunity
- Compared with other opportunities visible in the desk context, watchlist, or obvious alternatives, is this among the best uses of capital?
- Would you prioritize this now, keep it as a secondary idea, or defer capital elsewhere?

19. Asymmetry
- Estimate upside vs downside from the current price.
- Is the risk/reward asymmetric in a favorable or unfavorable way?

Be concise, analytical, and avoid generic explanations.
Focus on actionable insight, normalized reality, per-share value creation, and clear decisions.`

func (m *Model) launchAIPrompt(prompt, status string) tea.Cmd {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil
	}

	tabCmd := m.setActiveTab(tabAI)
	m.helpOpen = false
	m.searchMode = false
	m.searchInput.Blur()
	m.commandPaletteOpen = false
	m.commandInput.Blur()
	m.aiPickerOpen = false
	m.aiFocused = false
	m.aiInput.Blur()
	m.aiInput.SetValue("")

	if m.aiRunning {
		m.status = "AI already running…"
		return tabCmd
	}

	m.pushAIUserMessage(prompt)
	m.aiRunning = true
	m.aiErr = nil
	m.aiLastRequestTruncation = aiRequestTruncation{}
	if strings.TrimSpace(status) == "" {
		m.status = "Refreshing AI context…"
	} else {
		m.status = status
	}
	return tea.Batch(tabCmd, m.prepareAIContextCmd(prompt))
}
