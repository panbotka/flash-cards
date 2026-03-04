# REST API Reference

Base path: `/api`

All endpoints except `/api/auth/login` and `/api/auth/check` require authentication when `APP_PASSWORD` is set.

---

## Auth

### POST /api/auth/login

Authenticate with password.

**Request:**
```json
{ "password": "string" }
```

**Response (200):**
```json
{ "ok": true }
```

**Response (401):**
```json
{ "error": "invalid password" }
```

Sets an HTTP-only session cookie (`flash_session`) valid for 30 days. If `APP_PASSWORD` is not set, always succeeds.

### POST /api/auth/logout

Clears the session cookie.

**Response (200):**
```json
{ "ok": true }
```

### GET /api/auth/check

Check authentication status. Always accessible (no auth required).

**Response (200):**
```json
{
  "authenticated": true,
  "authRequired": true
}
```

---

## Cards

### GET /api/cards

List all non-deleted cards.

**Query params:**
| Param | Description |
|-------|-------------|
| `tag` | Filter by tag (exact match) |
| `search` | Search czech/english text (case-insensitive LIKE) |

**Response (200):**
```json
[
  {
    "id": 1,
    "czech": "kolo",
    "english": "bike",
    "suspended": false,
    "createdAt": "2026-01-01T00:00:00Z",
    "updatedAt": "2026-01-01T00:00:00Z",
    "tags": ["transport"]
  }
]
```

Ordered by `created_at` DESC. Tags fetched in a batch query (not N+1).

### GET /api/cards/:id

Get single card with tags and SRS states. Includes soft-deleted cards (for restore).

**Response (200):**
```json
{
  "id": 1,
  "czech": "kolo",
  "english": "bike",
  "deletedAt": null,
  "suspended": false,
  "createdAt": "2026-01-01T00:00:00Z",
  "updatedAt": "2026-01-01T00:00:00Z",
  "tags": ["transport"],
  "srsStates": [
    {
      "id": 1,
      "cardId": 1,
      "direction": "cz_en",
      "easeFactor": 2.5,
      "intervalDays": 0,
      "repetitions": 0,
      "nextReview": "2026-01-01T00:00:00Z",
      "status": "new",
      "learningStep": 0
    },
    {
      "id": 2,
      "cardId": 1,
      "direction": "en_cz",
      "easeFactor": 2.5,
      "intervalDays": 0,
      "repetitions": 0,
      "nextReview": "2026-01-01T00:00:00Z",
      "status": "new",
      "learningStep": 0
    }
  ]
}
```

### POST /api/cards

Create a card. Automatically creates two SRS states (`cz_en` and `en_cz`).

**Request:**
```json
{
  "czech": "kolo",
  "english": "bike",
  "tags": ["transport"]
}
```

`czech` and `english` are required. `tags` is optional.

**Response (201):** Full card object with SRS states.

### PUT /api/cards/:id

Partial update. Only non-null fields are applied.

**Request:**
```json
{
  "czech": "kolo",
  "english": "bicycle",
  "tags": ["transport", "vehicles"]
}
```

If `tags` is provided, old tags are replaced entirely. Updates `updated_at`.

**Response (200):** Updated card object.

### DELETE /api/cards/:id

Soft delete (sets `deleted_at`).

**Response:** 204 No Content

### POST /api/cards/:id/suspend

Toggle the `suspended` flag.

**Response (200):** Updated card object with SRS states.

### POST /api/cards/:id/restore

Restore a soft-deleted card (clears `deleted_at`). Returns 404 if the card is not deleted.

**Response (200):** Restored card object with SRS states.

---

## Study

### GET /api/study/next

Get the next due card for review.

**Query params:**
| Param | Description |
|-------|-------------|
| `tag` | Filter by tag |
| `direction` | `cz_en` or `en_cz` |

**Response when card available (200):**
```json
{
  "card": {
    "id": 1,
    "czech": "kolo",
    "english": "bike",
    "tags": ["transport"]
  },
  "srsState": {
    "id": 1,
    "cardId": 1,
    "direction": "cz_en",
    "easeFactor": 2.5,
    "intervalDays": 0,
    "repetitions": 0,
    "nextReview": "2026-01-01T00:00:00Z",
    "status": "new",
    "learningStep": 0
  },
  "intervalHints": {
    "1": "1m",
    "2": "1m",
    "3": "10m",
    "4": "4d"
  }
}
```

`intervalHints` show the human-readable interval for each rating (1=Again, 2=Hard, 3=Good, 4=Easy).

**Response when done (200):**
```json
{
  "done": true,
  "newAvailable": 5
}
```

Returns the most overdue card first (`next_review ASC`). Excludes deleted and suspended cards.

### GET /api/study/new

Same as `/study/next` but only returns new (never-reviewed) cards. Used when the user chooses to continue after all due cards are done.

### POST /api/study/review

Submit a review rating.

**Request:**
```json
{
  "srsStateId": 1,
  "rating": 3
}
```

`rating` must be 1-4.

**Response (200):**
```json
{
  "srsState": {
    "id": 1,
    "cardId": 1,
    "direction": "cz_en",
    "easeFactor": 2.5,
    "intervalDays": 1.0,
    "repetitions": 1,
    "nextReview": "2026-01-02T10:30:00Z",
    "status": "review",
    "learningStep": 0
  },
  "nextInterval": "1d"
}
```

Updates the SRS state and records a review event with before/after interval and ease values.

---

## Import

### POST /api/cards/import/preview

Parse bulk content and check for duplicates. Does not persist anything.

**Request:**
```json
{
  "content": "kolo\tbike\nauto\tcar"
}
```

Auto-detects delimiter (tab, semicolon, ` - `, ` = `, comma).

**Response (200):**
```json
{
  "cards": [
    { "czech": "kolo", "english": "bike", "isDuplicate": true },
    { "czech": "auto", "english": "car", "isDuplicate": false }
  ],
  "duplicates": 1,
  "total": 2
}
```

Duplicate check is case-insensitive against non-deleted cards.

### POST /api/cards/import/commit

Import cards. Skips duplicates and empty entries.

**Request:**
```json
{
  "cards": [
    { "czech": "auto", "english": "car" },
    { "czech": "kolo", "english": "bike" }
  ],
  "tags": ["lesson-1"]
}
```

**Response (200):**
```json
{
  "imported": 1,
  "skipped": 1
}
```

Runs in a single transaction. Creates SRS states for both directions on each imported card.

---

## Stats

All stats endpoints return data for non-deleted, non-suspended cards only.

### GET /api/stats/summary

**Response (200):**
```json
{
  "reviewsToday": 15,
  "totalCards": 120,
  "streak": 5,
  "accuracyToday": 0.87
}
```

- `streak`: Consecutive days with at least one review, going back from today.
- `accuracyToday`: Fraction (0-1) of today's reviews with rating >= 3.

### GET /api/stats/heatmap

Daily review counts for the last 365 days.

**Response (200):**
```json
[
  { "date": "2025-03-04", "count": 12 },
  { "date": "2025-03-05", "count": 0 }
]
```

Always returns 365 entries. Days with no reviews have `count: 0`.

### GET /api/stats/accuracy

Daily accuracy for the last 30 days. Only includes days with reviews.

**Response (200):**
```json
[
  { "date": "2026-02-01", "accuracy": 0.85, "total": 20 }
]
```

`accuracy` is a 0-1 fraction.

### GET /api/stats/maturity

Card maturity distribution.

**Response (200):**
```json
{
  "new": 30,
  "learning": 10,
  "young": 45,
  "mature": 35
}
```

- **new**: `status = 'new'`
- **learning**: `status = 'learning'`
- **young**: `status = 'review'` and `interval_days < 21`
- **mature**: `status = 'review'` and `interval_days >= 21`

Each card is classified by its worst direction (e.g., if one direction is "new" and the other "review", the card counts as "new").

### GET /api/stats/forecast

Upcoming due reviews per day for the next 30 days.

**Response (200):**
```json
[
  { "date": "2026-03-05", "count": 8 }
]
```

Always returns 30 entries.

### GET /api/stats/hardest

Cards with the lowest accuracy. Requires at least 3 reviews per card.

**Response (200):**
```json
[
  {
    "cardId": 5,
    "czech": "kolo",
    "english": "bike",
    "totalReviews": 15,
    "againCount": 8,
    "accuracy": 0.47
  }
]
```

Returns up to 20 cards, ordered by accuracy ascending. `accuracy` is a 0-1 fraction.
