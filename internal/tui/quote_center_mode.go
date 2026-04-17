package tui

type quoteCenterMode int

const (
	quoteCenterChart quoteCenterMode = iota
	quoteCenterFundamentals
	quoteCenterTechnicals
	quoteCenterSharpe
	quoteCenterStatements
	quoteCenterInsiders
	quoteCenterOwners
	quoteCenterAnalyst
	quoteCenterFilings
	quoteCenterEarnings
	quoteCenterNews
)
