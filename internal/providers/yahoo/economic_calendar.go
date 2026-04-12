package yahoo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"blackdesk/internal/domain"
)

type economicCalendarResponse struct {
	Finance struct {
		Result economicCalendarResultSet `json:"result"`
	} `json:"finance"`
}

type economicCalendarResultSet struct {
	Items []economicCalendarResult
}

type economicCalendarResult struct {
	EconomicEvents []economicCalendarDay `json:"economicEvents"`
}

type economicCalendarDay struct {
	Timestamp  int64                    `json:"timestamp"`
	Timezone   string                   `json:"timezone"`
	Count      int                      `json:"count"`
	TotalCount int                      `json:"totalCount"`
	Records    []economicCalendarRecord `json:"records"`
}

type economicCalendarRecord struct {
	Event             string            `json:"event"`
	CountryCode       string            `json:"countryCode"`
	EventTime         yahooScalarString `json:"eventTime"`
	Period            yahooScalarString `json:"period"`
	Actual            yahooScalarString `json:"actual"`
	ConsensusEstimate yahooScalarString `json:"consensusEstimate"`
	Prior             yahooScalarString `json:"prior"`
	RevisedFrom       yahooScalarString `json:"revisedFrom"`
	Description       yahooScalarString `json:"description"`
}

type yahooScalarString string

func (s *economicCalendarResultSet) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		s.Items = nil
		return nil
	}
	switch trimmed[0] {
	case '[':
		return json.Unmarshal(trimmed, &s.Items)
	case '{':
		var item economicCalendarResult
		if err := json.Unmarshal(trimmed, &item); err != nil {
			return err
		}
		s.Items = []economicCalendarResult{item}
		return nil
	default:
		return fmt.Errorf("unexpected yahoo economic calendar result payload")
	}
}

func (s *yahooScalarString) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*s = ""
		return nil
	}

	var text string
	if err := json.Unmarshal(trimmed, &text); err == nil {
		*s = yahooScalarString(strings.TrimSpace(text))
		return nil
	}

	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()
	var number json.Number
	if err := decoder.Decode(&number); err == nil {
		*s = yahooScalarString(number.String())
		return nil
	}

	var boolean bool
	if err := json.Unmarshal(trimmed, &boolean); err == nil {
		if boolean {
			*s = "true"
		} else {
			*s = "false"
		}
		return nil
	}

	return fmt.Errorf("unsupported yahoo scalar payload: %s", string(trimmed))
}

func (s yahooScalarString) String() string {
	return strings.TrimSpace(string(s))
}

func (p *Provider) GetEconomicCalendar(ctx context.Context, start, end time.Time) (domain.EconomicCalendarSnapshot, error) {
	if end.Before(start) || end.Equal(start) {
		return domain.EconomicCalendarSnapshot{}, fmt.Errorf("economic calendar range is invalid")
	}

	countPerDay := 50
	resp, err := p.fetchEconomicCalendar(ctx, start, end, countPerDay)
	if err != nil {
		return domain.EconomicCalendarSnapshot{}, err
	}
	if maxRequired := economicCalendarMaxRequiredCount(resp); maxRequired > countPerDay {
		resp, err = p.fetchEconomicCalendar(ctx, start, end, maxRequired)
		if err != nil {
			return domain.EconomicCalendarSnapshot{}, err
		}
	}
	return normalizeEconomicCalendarSnapshot(start, end, resp)
}

func (p *Provider) fetchEconomicCalendar(ctx context.Context, start, end time.Time, countPerDay int) (economicCalendarResponse, error) {
	var resp economicCalendarResponse
	params := url.Values{}
	params.Set("countPerDay", fmt.Sprintf("%d", max(25, countPerDay)))
	params.Set("economicEventsHighImportanceOnly", "true")
	params.Set("economicEventsRegionFilter", "")
	params.Set("startDate", fmt.Sprintf("%d", start.UTC().UnixMilli()))
	params.Set("endDate", fmt.Sprintf("%d", end.UTC().UnixMilli()))
	params.Set("modules", "economicEvents")
	params.Set("lang", "en-US")
	params.Set("region", "US")
	err := p.fetchJSON(ctx, requestSpec{
		URL:      p.calendarBase,
		Params:   params,
		CacheKey: fmt.Sprintf("economic-calendar:%d:%d:%d", start.UTC().UnixMilli(), end.UTC().UnixMilli(), max(25, countPerDay)),
		TTL:      15 * time.Minute,
		Auth:     authOptional,
	}, &resp)
	if err != nil {
		return economicCalendarResponse{}, err
	}
	return resp, nil
}

func normalizeEconomicCalendarSnapshot(start, end time.Time, resp economicCalendarResponse) (domain.EconomicCalendarSnapshot, error) {
	if len(resp.Finance.Result.Items) == 0 {
		return domain.EconomicCalendarSnapshot{}, fmt.Errorf("yahoo economic calendar unavailable")
	}

	snapshot := domain.EconomicCalendarSnapshot{
		StartDate: start,
		EndDate:   end,
		Freshness: domain.FreshnessLive,
		Provider:  "yahoo",
		UpdatedAt: time.Now(),
	}

	for _, day := range resp.Finance.Result.Items[0].EconomicEvents {
		dayTime := yahooCalendarTimestamp(day.Timestamp)
		for _, record := range day.Records {
			eventAt := economicCalendarEventTimestamp(record.EventTime.String())
			event := domain.EconomicCalendarEvent{
				Date:              dayTime,
				EventAt:           eventAt,
				CountryCode:       strings.ToUpper(strings.TrimSpace(record.CountryCode)),
				Event:             strings.TrimSpace(record.Event),
				EventTime:         formatEconomicCalendarEventTime(eventAt),
				Period:            record.Period.String(),
				Actual:            record.Actual.String(),
				ConsensusEstimate: record.ConsensusEstimate.String(),
				Prior:             record.Prior.String(),
				RevisedFrom:       record.RevisedFrom.String(),
				Description:       record.Description.String(),
			}
			if event.Event == "" {
				continue
			}
			snapshot.Events = append(snapshot.Events, event)
		}
	}

	sort.SliceStable(snapshot.Events, func(i, j int) bool {
		left := snapshot.Events[i]
		right := snapshot.Events[j]
		leftTime := economicCalendarSortTime(left)
		rightTime := economicCalendarSortTime(right)
		if !leftTime.Equal(rightTime) {
			return leftTime.Before(rightTime)
		}
		if left.EventTime != right.EventTime {
			return left.EventTime < right.EventTime
		}
		if left.CountryCode != right.CountryCode {
			return left.CountryCode < right.CountryCode
		}
		return left.Event < right.Event
	})

	return snapshot, nil
}

func yahooCalendarTimestamp(raw int64) time.Time {
	switch {
	case raw > 1_000_000_000_000:
		return time.UnixMilli(raw).UTC()
	case raw > 0:
		return time.Unix(raw, 0).UTC()
	default:
		return time.Time{}
	}
}

func economicCalendarEventTimestamp(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	millis, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || millis <= 0 {
		return time.Time{}
	}
	return yahooCalendarTimestamp(millis)
}

func formatEconomicCalendarEventTime(timestamp time.Time) string {
	if timestamp.IsZero() {
		return ""
	}
	return timestamp.In(time.Local).Format("15:04")
}

func economicCalendarSortTime(item domain.EconomicCalendarEvent) time.Time {
	if !item.EventAt.IsZero() {
		return item.EventAt.In(time.Local)
	}
	return item.Date
}

func economicCalendarMaxRequiredCount(resp economicCalendarResponse) int {
	if len(resp.Finance.Result.Items) == 0 {
		return 0
	}
	maxRequired := 0
	for _, day := range resp.Finance.Result.Items[0].EconomicEvents {
		required := max(day.TotalCount, day.Count)
		if required < len(day.Records) {
			required = len(day.Records)
		}
		if required > maxRequired {
			maxRequired = required
		}
	}
	return maxRequired
}
