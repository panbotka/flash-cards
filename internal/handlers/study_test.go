package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/pi/flash-cards/internal/models"
)

func TestNextCard(t *testing.T) {
	r := setupTestRouter(t)
	card := createTestCard(t, r, "ahoj", "hello", []string{"greetings"})

	t.Run("returns new card as next due", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/study/next?direction=cz_en", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want 200: %s", w.Code, w.Body.String())
		}

		var resp map[string]json.RawMessage
		json.Unmarshal(w.Body.Bytes(), &resp)

		// Should have card, srsState, and intervalHints.
		if _, ok := resp["card"]; !ok {
			t.Error("response missing 'card'")
		}
		if _, ok := resp["srsState"]; !ok {
			t.Error("response missing 'srsState'")
		}
		if _, ok := resp["intervalHints"]; !ok {
			t.Error("response missing 'intervalHints'")
		}

		var respCard models.Card
		json.Unmarshal(resp["card"], &respCard)
		if respCard.ID != card.ID {
			t.Errorf("card id: got %d, want %d", respCard.ID, card.ID)
		}

		var state models.SRSState
		json.Unmarshal(resp["srsState"], &state)
		if state.Direction != "cz_en" {
			t.Errorf("direction: got %q, want %q", state.Direction, "cz_en")
		}
	})

	t.Run("en_cz direction returns en_cz state", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/study/next?direction=en_cz", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want 200", w.Code)
		}

		var resp map[string]json.RawMessage
		json.Unmarshal(w.Body.Bytes(), &resp)

		var state models.SRSState
		json.Unmarshal(resp["srsState"], &state)
		if state.Direction != "en_cz" {
			t.Errorf("direction: got %q, want %q", state.Direction, "en_cz")
		}
	})

	t.Run("invalid direction returns error", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/study/next?direction=bad", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status: got %d, want 400", w.Code)
		}
	})
}

func TestNextCardTagFilter(t *testing.T) {
	r := setupTestRouter(t)
	createTestCard(t, r, "pes", "dog", []string{"animals"})
	createTestCard(t, r, "stůl", "table", []string{"furniture"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/study/next?direction=cz_en&tag=animals", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}

	var resp map[string]json.RawMessage
	json.Unmarshal(w.Body.Bytes(), &resp)

	var card models.Card
	json.Unmarshal(resp["card"], &card)

	if card.Czech != "pes" {
		t.Errorf("expected animals card, got %q", card.Czech)
	}
}

func TestNextCardNoDue(t *testing.T) {
	r := setupTestRouter(t)

	// No cards at all — should return done with 0 new.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/study/next?direction=cz_en", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}

	var resp models.StudyDoneResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Done {
		t.Error("expected done=true when no cards exist")
	}
}

func TestNextCardSkipsSuspendedAndDeleted(t *testing.T) {
	r := setupTestRouter(t)
	suspended := createTestCard(t, r, "voda", "water", []string{"nature"})
	deleted := createTestCard(t, r, "oheň", "fire", []string{"nature"})

	// Suspend the first card.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/cards/"+itoa(suspended.ID)+"/suspend", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("suspend: got %d, want 200", w.Code)
	}

	// Delete the second card.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/cards/"+itoa(deleted.ID), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: got %d, want 204", w.Code)
	}

	// Next should return done since both cards are excluded.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/study/next?direction=cz_en", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}

	var resp models.StudyDoneResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Done {
		t.Error("expected done=true when all cards are suspended or deleted")
	}
}

func TestSubmitReview(t *testing.T) {
	r := setupTestRouter(t)
	card := createTestCard(t, r, "město", "city", []string{"places"})

	// Find the cz_en SRS state ID.
	var srsStateID int64
	for _, s := range card.SRSStates {
		if s.Direction == "cz_en" {
			srsStateID = s.ID
			break
		}
	}

	t.Run("returns updated state and nextInterval", func(t *testing.T) {
		body, _ := json.Marshal(models.ReviewRequest{
			SRSStateID: srsStateID,
			Rating:     3,
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/study/review", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want 200: %s", w.Code, w.Body.String())
		}

		var resp models.ReviewResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp.SRSState.ID != srsStateID {
			t.Errorf("srs state id: got %d, want %d", resp.SRSState.ID, srsStateID)
		}
		if resp.SRSState.Status == "new" {
			t.Error("status should have changed from 'new' after review")
		}
		if resp.NextInterval == "" {
			t.Error("nextInterval should not be empty")
		}
	})

	t.Run("state is persisted", func(t *testing.T) {
		// Fetch the card again and verify the SRS state was updated.
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/cards/"+itoa(card.ID), nil)
		r.ServeHTTP(w, req)

		var fetched models.Card
		json.Unmarshal(w.Body.Bytes(), &fetched)

		for _, s := range fetched.SRSStates {
			if s.Direction == "cz_en" {
				if s.Status == "new" {
					t.Error("persisted cz_en state should no longer be 'new'")
				}
				// Rating=3 (Good) on step 0 moves to learning step 1,
				// so status should be "learning".
				if s.Status != "learning" {
					t.Errorf("persisted cz_en status: got %q, want %q", s.Status, "learning")
				}
				break
			}
		}
	})

	t.Run("review event is recorded", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/cards/"+itoa(card.ID)+"/history", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("history status: got %d, want 200", w.Code)
		}

		var hist models.CardHistoryResponse
		json.Unmarshal(w.Body.Bytes(), &hist)

		if len(hist.Reviews) == 0 {
			t.Error("expected at least one review event")
		}
		if hist.Reviews[0].Rating != 3 {
			t.Errorf("review rating: got %d, want 3", hist.Reviews[0].Rating)
		}
	})

	t.Run("invalid SRS state ID returns 404", func(t *testing.T) {
		body, _ := json.Marshal(models.ReviewRequest{
			SRSStateID: 99999,
			Rating:     3,
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/study/review", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status: got %d, want 404", w.Code)
		}
	})
}

func TestNewCard(t *testing.T) {
	r := setupTestRouter(t)
	createTestCard(t, r, "kniha", "book", []string{"objects"})

	t.Run("returns new card", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/study/new?direction=cz_en", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want 200", w.Code)
		}

		var resp map[string]json.RawMessage
		json.Unmarshal(w.Body.Bytes(), &resp)

		if _, ok := resp["card"]; !ok {
			t.Error("response missing 'card'")
		}

		var card models.Card
		json.Unmarshal(resp["card"], &card)
		if card.Czech != "kniha" {
			t.Errorf("card: got %q, want %q", card.Czech, "kniha")
		}
	})

	t.Run("returns done when no new cards", func(t *testing.T) {
		// Review the card so it's no longer new.
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/study/new?direction=cz_en", nil)
		r.ServeHTTP(w, req)

		var resp map[string]json.RawMessage
		json.Unmarshal(w.Body.Bytes(), &resp)
		var state models.SRSState
		json.Unmarshal(resp["srsState"], &state)

		body, _ := json.Marshal(models.ReviewRequest{
			SRSStateID: state.ID,
			Rating:     3,
		})
		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/api/study/review", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		// Now /study/new should return done.
		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/api/study/new?direction=cz_en", nil)
		r.ServeHTTP(w, req)

		var doneResp models.StudyDoneResponse
		json.Unmarshal(w.Body.Bytes(), &doneResp)
		if !doneResp.Done {
			t.Error("expected done=true after reviewing the only new card")
		}
	})
}

func TestFullStudyCycle(t *testing.T) {
	r := setupTestRouter(t)

	// 1. Create a card.
	card := createTestCard(t, r, "dům", "house", []string{"places"})

	// 2. Fetch next — should return this card.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/study/next?direction=cz_en", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("next status: got %d, want 200", w.Code)
	}

	var studyResp map[string]json.RawMessage
	json.Unmarshal(w.Body.Bytes(), &studyResp)

	var state models.SRSState
	json.Unmarshal(studyResp["srsState"], &state)

	if state.CardID != card.ID {
		t.Fatalf("next card id: got %d, want %d", state.CardID, card.ID)
	}

	// 3. Submit review with rating=4 (Easy) — moves to review with long interval.
	body, _ := json.Marshal(models.ReviewRequest{
		SRSStateID: state.ID,
		Rating:     4,
	})
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/study/review", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("review status: got %d, want 200: %s", w.Code, w.Body.String())
	}

	// 4. Fetch next again — card should no longer be immediately due.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/study/next?direction=cz_en", nil)
	r.ServeHTTP(w, req)

	var doneResp models.StudyDoneResponse
	json.Unmarshal(w.Body.Bytes(), &doneResp)

	if !doneResp.Done {
		t.Error("expected done=true after reviewing the only card")
	}
}

// itoa converts int64 to string (convenience for URL building).
func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}
