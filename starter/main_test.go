
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func resetEvents() {
	eventsMu.Lock()
	defer eventsMu.Unlock()
	events = make(map[string]Event)
}

func TestPostEvents(t *testing.T) {
	resetEvents()

	payload := `[
		{
			"event_id": "evt_test_1",
			"campaign_id": "cmp_test",
			"contact_id": "ct_1",
			"type": "sent",
			"timestamp": "2026-08-10T14:20:00Z"
		},
		{
			"event_id": "evt_test_2",
			"campaign_id": "cmp_test",
			"contact_id": "ct_2",
			"type": "opened",
			"timestamp": "2026-08-10T14:25:00Z"
		}
	]`

	req := httptest.NewRequest(
		http.MethodPost,
		"/events",
		bytes.NewBufferString(payload),
	)

	rec := httptest.NewRecorder()

	handlePostEvents(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var result map[string]int

	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}

	if result["accepted"] != 2 {
		t.Errorf("expected 2 accepted, got %d", result["accepted"])
	}

	if result["duplicates"] != 0 {
		t.Errorf("expected 0 duplicates, got %d", result["duplicates"])
	}

	if result["rejected"] != 0 {
		t.Errorf("expected 0 rejected, got %d", result["rejected"])
	}
}

func TestCampaignStats(t *testing.T) {
	resetEvents()

	events["evt_1"] = Event{
		EventID:    "evt_1",
		CampaignID: "cmp_test",
		ContactID:  "ct_1",
		Type:       "sent",
		Timestamp:  "2026-08-10T14:20:00Z",
	}

	events["evt_2"] = Event{
		EventID:    "evt_2",
		CampaignID: "cmp_test",
		ContactID:  "ct_1",
		Type:       "opened",
		Timestamp:  "2026-08-10T14:25:00Z",
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/campaigns/cmp_test/stats",
		nil,
	)

	req.SetPathValue("campaignID", "cmp_test")

	rec := httptest.NewRecorder()

	handleGetStats(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var result map[string]int

	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}

	if result["sent"] != 1 {
		t.Errorf("expected sent=1, got %d", result["sent"])
	}

	if result["opened"] != 1 {
		t.Errorf("expected opened=1, got %d", result["opened"])
	}
}


func TestDuplicatesAndInvalidEvents(t *testing.T) {
	resetEvents()

	payload := `[
		{
			"event_id": "evt_1",
			"campaign_id": "cmp_test",
			"contact_id": "ct_1",
			"type": "sent",
			"timestamp": "2026-08-10T14:20:00Z"
		},
		{
			"event_id": "evt_1",
			"campaign_id": "cmp_test",
			"contact_id": "ct_1",
			"type": "sent",
			"timestamp": "2026-08-10T14:20:00Z"
		},
		{
			"event_id": "evt_2",
			"campaign_id": "cmp_test",
			"contact_id": "ct_2",
			"type": "invalid",
			"timestamp": "2026-08-10T14:20:00Z"
		}
	]`

	req := httptest.NewRequest(
		http.MethodPost,
		"/events",
		bytes.NewBufferString(payload),
	)

	rec := httptest.NewRecorder()
	handlePostEvents(rec, req)

	var result map[string]int
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}

	if result["accepted"] != 1 {
		t.Errorf("expected accepted=1, got %d", result["accepted"])
	}

	if result["duplicates"] != 1 {
		t.Errorf("expected duplicates=1, got %d", result["duplicates"])
	}

	if result["rejected"] != 1 {
		t.Errorf("expected rejected=1, got %d", result["rejected"])
	}
}
