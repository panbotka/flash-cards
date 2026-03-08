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

func TestCreateCard(t *testing.T) {
	r := setupTestRouter(t)
	card := createTestCard(t, r, "ahoj", "hello", []string{"greetings"})

	if card.Czech != "ahoj" {
		t.Errorf("czech: got %q, want %q", card.Czech, "ahoj")
	}
	if card.English != "hello" {
		t.Errorf("english: got %q, want %q", card.English, "hello")
	}
	if len(card.Tags) != 1 || card.Tags[0] != "greetings" {
		t.Errorf("tags: got %v, want [greetings]", card.Tags)
	}
	if len(card.SRSStates) != 2 {
		t.Fatalf("srs states: got %d, want 2", len(card.SRSStates))
	}

	dirs := map[string]bool{}
	for _, s := range card.SRSStates {
		dirs[s.Direction] = true
		if s.Status != "new" {
			t.Errorf("srs state %s: status got %q, want %q", s.Direction, s.Status, "new")
		}
	}
	if !dirs["cz_en"] || !dirs["en_cz"] {
		t.Errorf("expected both cz_en and en_cz directions, got %v", dirs)
	}
}

func TestListCards(t *testing.T) {
	r := setupTestRouter(t)
	createTestCard(t, r, "pes", "dog", []string{"animals"})
	createTestCard(t, r, "kočka", "cat", []string{"animals"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/cards", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}

	var cards []models.Card
	if err := json.Unmarshal(w.Body.Bytes(), &cards); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if len(cards) != 2 {
		t.Errorf("card count: got %d, want 2", len(cards))
	}
	for _, c := range cards {
		if len(c.Tags) != 1 || c.Tags[0] != "animals" {
			t.Errorf("card %d tags: got %v, want [animals]", c.ID, c.Tags)
		}
	}
}

func TestGetCard(t *testing.T) {
	r := setupTestRouter(t)
	created := createTestCard(t, r, "dům", "house", []string{"places"})

	t.Run("existing card", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/cards/"+strconv.FormatInt(created.ID, 10), nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want 200", w.Code)
		}

		var card models.Card
		if err := json.Unmarshal(w.Body.Bytes(), &card); err != nil {
			t.Fatalf("parse response: %v", err)
		}
		if card.Czech != "dům" {
			t.Errorf("czech: got %q, want %q", card.Czech, "dům")
		}
		if len(card.SRSStates) != 2 {
			t.Errorf("srs states: got %d, want 2", len(card.SRSStates))
		}
	})

	t.Run("non-existent card", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/cards/99999", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status: got %d, want 404", w.Code)
		}
	})
}

func TestUpdateCard(t *testing.T) {
	r := setupTestRouter(t)
	created := createTestCard(t, r, "stůl", "desk", []string{"furniture"})
	cardURL := "/api/cards/" + strconv.FormatInt(created.ID, 10)

	t.Run("full update", func(t *testing.T) {
		english := "table"
		tags := []string{"furniture", "kitchen"}
		body, _ := json.Marshal(models.UpdateCardRequest{
			English: &english,
			Tags:    &tags,
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, cardURL, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want 200: %s", w.Code, w.Body.String())
		}

		var card models.Card
		json.Unmarshal(w.Body.Bytes(), &card)

		if card.Czech != "stůl" {
			t.Errorf("czech should be unchanged: got %q, want %q", card.Czech, "stůl")
		}
		if card.English != "table" {
			t.Errorf("english: got %q, want %q", card.English, "table")
		}
		if len(card.Tags) != 2 {
			t.Errorf("tags count: got %d, want 2", len(card.Tags))
		}
	})

	t.Run("partial update preserves other fields", func(t *testing.T) {
		tags := []string{"household"}
		body, _ := json.Marshal(models.UpdateCardRequest{Tags: &tags})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, cardURL, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want 200", w.Code)
		}

		var card models.Card
		json.Unmarshal(w.Body.Bytes(), &card)

		if card.Czech != "stůl" {
			t.Errorf("czech should be unchanged: got %q", card.Czech)
		}
		if card.English != "table" {
			t.Errorf("english should be unchanged: got %q", card.English)
		}
	})
}

func TestDeleteCard(t *testing.T) {
	r := setupTestRouter(t)
	created := createTestCard(t, r, "kniha", "book", []string{"objects"})

	// Delete the card.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/cards/"+strconv.FormatInt(created.ID, 10), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("delete status: got %d, want 204", w.Code)
	}

	// Verify it no longer appears in list.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/cards", nil)
	r.ServeHTTP(w, req)

	var cards []models.Card
	json.Unmarshal(w.Body.Bytes(), &cards)
	if len(cards) != 0 {
		t.Errorf("card count after delete: got %d, want 0", len(cards))
	}
}

func TestSuspendAndRestore(t *testing.T) {
	r := setupTestRouter(t)
	created := createTestCard(t, r, "voda", "water", []string{"nature"})
	cardURL := "/api/cards/" + strconv.FormatInt(created.ID, 10)

	// Suspend the card.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, cardURL+"/suspend", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("suspend status: got %d, want 200", w.Code)
	}

	var card models.Card
	json.Unmarshal(w.Body.Bytes(), &card)
	if !card.Suspended {
		t.Error("card should be suspended after suspend")
	}

	// Verify suspended card doesn't appear in study.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/study/next?direction=cz_en", nil)
	r.ServeHTTP(w, req)

	var studyResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &studyResp)
	if done, ok := studyResp["done"]; ok && done == false {
		if cardData, ok := studyResp["card"].(map[string]interface{}); ok {
			if int64(cardData["id"].(float64)) == created.ID {
				t.Error("suspended card should not appear in study")
			}
		}
	}

	// Unsuspend by calling suspend again (toggle).
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, cardURL+"/suspend", nil)
	r.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &card)
	if card.Suspended {
		t.Error("card should be unsuspended after second toggle")
	}
}

func TestListTags(t *testing.T) {
	r := setupTestRouter(t)
	createTestCard(t, r, "pes", "dog", []string{"animals", "pets"})
	createTestCard(t, r, "kočka", "cat", []string{"animals"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}

	var tags []string
	json.Unmarshal(w.Body.Bytes(), &tags)

	if len(tags) != 2 {
		t.Fatalf("tag count: got %d, want 2", len(tags))
	}
	// Tags are ordered alphabetically.
	if tags[0] != "animals" || tags[1] != "pets" {
		t.Errorf("tags: got %v, want [animals pets]", tags)
	}
}

func TestRenameTag(t *testing.T) {
	r := setupTestRouter(t)
	createTestCard(t, r, "pes", "dog", []string{"animals"})
	createTestCard(t, r, "kočka", "cat", []string{"animals"})

	body, _ := json.Marshal(models.RenameTagRequest{
		OldName: "animals",
		NewName: "creatures",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/tags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("rename status: got %d, want 200: %s", w.Code, w.Body.String())
	}

	// Verify old tag is gone and new tag exists.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	r.ServeHTTP(w, req)

	var tags []string
	json.Unmarshal(w.Body.Bytes(), &tags)

	if len(tags) != 1 || tags[0] != "creatures" {
		t.Errorf("tags after rename: got %v, want [creatures]", tags)
	}
}

func TestDeleteTag(t *testing.T) {
	r := setupTestRouter(t)
	createTestCard(t, r, "pes", "dog", []string{"animals"})
	createTestCard(t, r, "stůl", "table", []string{"furniture"})

	// Delete "animals" tag — should soft-delete the dog card.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/tags/animals", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("delete tag status: got %d, want 200: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if deleted, ok := resp["deleted"].(float64); !ok || deleted != 1 {
		t.Errorf("deleted count: got %v, want 1", resp["deleted"])
	}

	// Verify only the furniture card remains.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/cards", nil)
	r.ServeHTTP(w, req)

	var cards []models.Card
	json.Unmarshal(w.Body.Bytes(), &cards)

	if len(cards) != 1 {
		t.Fatalf("card count: got %d, want 1", len(cards))
	}
	if cards[0].Czech != "stůl" {
		t.Errorf("remaining card: got %q, want %q", cards[0].Czech, "stůl")
	}
}
