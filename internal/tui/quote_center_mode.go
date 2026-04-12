package tui

type quoteCenterMode int

const (
	quoteCenterChart quoteCenterMode = iota
	quoteCenterFundamentals
	quoteCenterTechnicals
	quoteCenterStatements
	quoteCenterInsiders
	quoteCenterOwners
	quoteCenterAnalyst
	quoteCenterFilings
	quoteCenterEarnings
	quoteCenterNews
)
