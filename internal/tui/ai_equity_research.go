package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const aiEquityResearchPrompt = `Analyze this company as an investment opportunity with the goal of making a clear capital allocation decision.

0. Company Type (CRITICAL FIRST STEP)
- Classify the company as one of:
  • Mature / cash-generating
  • High-growth / reinvesting
  • Cyclical / commodity-driven
- Adjust all analysis accordingly.

1. Business Quality
- Evaluate ROIC, margins, and durability of the business model.
- Is this a high-quality compounder, cyclical business, or average company?
- Does the company have a durable competitive advantage?

2. Earnings Quality
- Are current earnings (EBITDA, FCF, margins) sustainable?
- Are we at peak, normal, or depressed earnings?
- Explain what is driving any distortion (cycle, capex, one-offs, etc.).

3. Normalization
- Estimate normalized EBITDA and normalized free cash flow.
- For cyclicals: normalize across the cycle.
- For growth companies: focus on margin trajectory instead of penalizing low current FCF.
- Recalculate EV/EBITDA and FCF yield using normalized values where applicable.

4. Valuation (ADAPTIVE)
- For mature companies:
  • Use earnings yield, FCF yield, EV/EBITDA
- For growth companies:
  • Focus on revenue growth, margin expansion, and long-term profitability
  • Do not penalize low FCF if reinvestment has high returns
- Conclude: cheap / fair / expensive

5. Growth vs Expectations
- What growth is implied by the current price?
- Is the market too optimistic, too pessimistic, or roughly correct?

6. Expected Return
- Estimate long-term expected return:
  Expected Return ≈ Earnings Yield + Sustainable Growth
- For growth companies: base this on normalized future margins
- Provide a range:
  • Bear
  • Base
  • Bull

7. Risk
- Identify the 2–3 key risks that could break the thesis.

8. Alpha / Edge
- Compare expected return to a 10% baseline (index).
- Is this clearly better, similar, or worse?
- What is the market likely mispricing?

9. Decision (FIRST LEVEL)
- Give a clear verdict:
  BUY / HOLD / PASS
- Explain briefly why.
- Specify investor type (value, growth, defensive, cyclical, etc.).

10. Price Sensitivity (CRITICAL)
- At what price does this become:
  • BUY
  • STRONG BUY
- Provide clear zones:
  • AVOID / HOLD / BUY / STRONG BUY

11. Capital Allocation Decision (FINAL)
- You must choose between this stock and a broad index (~10% return).
- Choose ONLY ONE:
  • BUY (clear alpha)
  • HOLD (keep if owned, not attractive for new capital)
  • PASS (no edge vs market)
- No mixed answers. Be decisive.

12. Upside Trigger
- What must happen for this to deliver 15%+ returns?

Be concise, analytical, and avoid generic explanations.
Focus on actionable insight, normalized reality, and clear decisions.`

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
