package tui

type quoteCenterMode int

const (
	quoteCenterChart quoteCenterMode = iota
	quoteCenterFundamentals
	quoteCenterTechnicals
	quoteCenterStatements
	quoteCenterInsiders
	quoteCenterFilings
	quoteCenterEarnings
	quoteCenterNews
)
