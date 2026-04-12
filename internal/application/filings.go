package application

import (
	"context"

	"blackdesk/internal/domain"
)

type FilingsProvider interface {
	GetFilings(context.Context, string) (domain.FilingsSnapshot, error)
	GetFilingDocument(context.Context, domain.FilingItem) (domain.FilingDocument, error)
}
