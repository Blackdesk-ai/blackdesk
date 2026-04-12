package yahoo

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestGetEconomicCalendarUsesCalendarEventsEndpoint(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 19, 0, 0, 0, 0, time.UTC)

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/ws/screeners/v1/finance/calendar-events":
			q := req.URL.Query()
			if got := q.Get("modules"); got != "economicEvents" {
				t.Fatalf("unexpected modules %q", got)
			}
			if got := q.Get("economicEventsHighImportanceOnly"); got != "true" {
				t.Fatalf("unexpected importance flag %q", got)
			}
			if got := q.Get("startDate"); got != "1775952000000" {
				t.Fatalf("unexpected startDate %q", got)
			}
			if got := q.Get("endDate"); got != "1776556800000" {
				t.Fatalf("unexpected endDate %q", got)
			}
			body := []byte(`{
				"finance": {
					"result": {
						"economicEvents": [
							{
								"timestamp": 1775952000000,
								"timezone": "America/New_York",
								"records": [
									{
										"event": "Consumer Price Index YoY",
										"countryCode": "us",
										"eventTime": 1775997000000,
										"period": "Mar",
										"actual": "3.1%",
										"consensusEstimate": "3.0%",
										"prior": "3.2%",
										"revisedFrom": "",
										"description": "Inflation release."
									}
								]
							}
						]
					}
				}
			}`)
			return jsonResponse(req, http.StatusOK, body, "")
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetEconomicCalendar(ctx, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Events) != 1 {
		t.Fatalf("expected 1 event, got %+v", got.Events)
	}
	if got.Events[0].Event != "Consumer Price Index YoY" {
		t.Fatalf("unexpected event %+v", got.Events[0])
	}
	if got.Events[0].CountryCode != "US" {
		t.Fatalf("expected normalized country code, got %+v", got.Events[0])
	}
	expectedTime := time.UnixMilli(1775997000000).In(time.Local).Format("15:04")
	if got.Events[0].EventTime != expectedTime {
		t.Fatalf("expected formatted event time, got %+v", got.Events[0])
	}
	if got.Events[0].EventAt.IsZero() {
		t.Fatalf("expected parsed event timestamp, got %+v", got.Events[0])
	}
	if !got.StartDate.Equal(start) || !got.EndDate.Equal(end) {
		t.Fatalf("unexpected snapshot range %+v", got)
	}
}

func TestGetEconomicCalendarRefetchesWhenYahooSignalsMoreEvents(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 19, 0, 0, 0, 0, time.UTC)
	requests := 0

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/ws/screeners/v1/finance/calendar-events" {
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
		requests++
		countPerDay := req.URL.Query().Get("countPerDay")
		switch countPerDay {
		case "50":
			body := []byte(`{
				"finance": {
					"result": {
						"economicEvents": [
							{
								"timestamp": 1775952000000,
								"timezone": "America/New_York",
								"count": 50,
								"totalCount": 70,
								"records": []
							}
						]
					}
				}
			}`)
			return jsonResponse(req, http.StatusOK, body, "")
		case "70":
			body := []byte(`{
				"finance": {
					"result": {
						"economicEvents": [
							{
								"timestamp": 1775952000000,
								"timezone": "America/New_York",
								"count": 70,
								"totalCount": 70,
								"records": [
									{
										"event": "Consumer Price Index YoY",
										"countryCode": "US",
										"eventTime": 1775997000000
									}
								]
							}
						]
					}
				}
			}`)
			return jsonResponse(req, http.StatusOK, body, "")
		default:
			t.Fatalf("unexpected countPerDay %q", countPerDay)
			return nil, nil
		}
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetEconomicCalendar(ctx, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 {
		t.Fatalf("expected 2 requests, got %d", requests)
	}
	if len(got.Events) != 1 || got.Events[0].Event != "Consumer Price Index YoY" {
		t.Fatalf("expected refetched event set, got %+v", got.Events)
	}
}
