package tui

import (
	"strings"

	"blackdesk/internal/domain"
)

type ownerListItem struct {
	kind   string
	holder domain.OwnershipHolder
}

func (m *Model) cacheOwners(snapshot domain.OwnershipSnapshot) {
	key := strings.ToUpper(strings.TrimSpace(snapshot.Symbol))
	if key == "" {
		return
	}
	m.ownersCache[key] = snapshot
}

func (m Model) cachedOwners(symbol string) (domain.OwnershipSnapshot, bool) {
	key := strings.ToUpper(strings.TrimSpace(symbol))
	if key == "" {
		return domain.OwnershipSnapshot{}, false
	}
	if snapshot, ok := m.ownersCache[key]; ok {
		return snapshot, true
	}
	if strings.EqualFold(m.owners.Symbol, key) {
		return m.owners, true
	}
	return domain.OwnershipSnapshot{}, false
}

func (m Model) ownersForSymbol(symbol string) domain.OwnershipSnapshot {
	snapshot, _ := m.cachedOwners(symbol)
	return snapshot
}

func (m Model) ownerItemsForSymbol(symbol string) []ownerListItem {
	snapshot := m.ownersForSymbol(symbol)
	items := make([]ownerListItem, 0, len(snapshot.Institutions)+len(snapshot.Funds))
	for _, holder := range snapshot.Institutions {
		items = append(items, ownerListItem{kind: "Institution", holder: holder})
	}
	for _, holder := range snapshot.Funds {
		items = append(items, ownerListItem{kind: "Fund", holder: holder})
	}
	return items
}

func (m *Model) cycleOwnersSelection(step int) {
	items := m.ownerItemsForSymbol(m.activeSymbol())
	if len(items) == 0 {
		m.ownersSel = 0
		return
	}
	m.ownersSel = (m.ownersSel + step + len(items)) % len(items)
}

func (m Model) currentOwnerItem() (ownerListItem, bool) {
	items := m.ownerItemsForSymbol(m.activeSymbol())
	if len(items) == 0 || m.ownersSel < 0 || m.ownersSel >= len(items) {
		return ownerListItem{}, false
	}
	return items[m.ownersSel], true
}
