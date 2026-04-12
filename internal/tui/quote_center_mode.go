package tui

type quoteCenterMode int

const (
	quoteCenterChart quoteCenterMode = iota
	quoteCenterFundamentals
	quoteCenterTechnicals
	quoteCenterStatements
	quoteCenterInsiders
	quoteCenterAnalyst
	quoteCenterFilings
	quoteCenterEarnings
	quoteCenterNews
)
