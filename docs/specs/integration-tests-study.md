# Study API Integration Tests

Test the study session endpoints using the test infrastructure from the cards integration test task.

## Requirements

### Study Handler Tests (`internal/handlers/study_test.go`)

- Test `GET /api/study/next` — with a new card in the database, returns it as the next due card with interval hints
- Test `GET /api/study/next` with `?direction=en_cz` — returns the en_cz SRS state, not cz_en
- Test `GET /api/study/next` with `?tag=food` — only returns cards tagged "food"
- Test `GET /api/study/next` when no cards are due — returns `{"done": true, "newAvailable": N}`
- Test `GET /api/study/next` skips suspended and soft-deleted cards
- Test `GET /api/study/next` returns invalid direction error for bad `?direction=` value
- Test `POST /api/study/review` — submit a rating (2, 3, or 4) for an SRS state; verify the response contains updated SRS state and a human-readable `nextInterval` string
- Test `POST /api/study/review` — verify the SRS state is actually persisted (fetch the card again and check updated fields)
- Test `POST /api/study/review` — verify a review event is recorded (query the database or check via stats endpoint)
- Test `POST /api/study/review` with invalid SRS state ID — returns 404
- Test `GET /api/study/new` — returns the next new card; returns done when no new cards exist
- Test full study cycle: create card → fetch next → submit review → verify card is no longer immediately due

## Implementation Notes

- Reuse `setupTestRouter()` and `createTestCard()` from `testhelper_test.go` (same package, shared automatically)
- For the "no longer due" assertion, check that `GET /api/study/next` returns `done: true` after reviewing all available cards
- Run with `go test ./internal/handlers/`
