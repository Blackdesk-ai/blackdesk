package tui

type quoteCenterMode int

const (
	quoteCenterChart quoteCenterMode = iota
	quoteCenterFundamentals
	quoteCenterTechnicals
	quoteCenterSharpe
	quoteCenterStatistics
	quoteCenterStatements
	quoteCenterInsiders
	quoteCenterOwners
	quoteCenterAnalyst
	quoteCenterFilings
	quoteCenterEarnings
	quoteCenterNews
)
