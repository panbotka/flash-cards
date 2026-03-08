# Test Infrastructure and Cards API Integration Tests

Create a reusable test helper for handler integration tests, then use it to test the cards API endpoints.

## Requirements

### Test Helper (`internal/handlers/testhelper_test.go`)

- Create a `setupTestRouter()` function that returns a `*gin.Engine` backed by an in-memory SQLite database
- The helper must run the same schema migrations as production (`internal/db` package)
- Enable WAL mode and foreign keys on the test database, matching production settings
- Register all handler groups on the test router so cross-handler interactions work
- No authentication middleware on the test router (matches dev mode)
- Provide a `createTestCard(t, router, czech, english, tags)` helper that POSTs to `/api/cards` and returns the created card — used by multiple test files

### Cards Handler Tests (`internal/handlers/cards_test.go`)

- Test `POST /api/cards` — creates a card with czech, english, and tags; verify 201 response and that two SRS states (cz_en, en_cz) are created
- Test `GET /api/cards` — returns list of cards with their tags and SRS states
- Test `GET /api/cards/:id` — returns a single card by ID; verify 404 for non-existent ID
- Test `PUT /api/cards/:id` — updates czech, english, and tags fields; verify partial updates work (only sending `tags` doesn't clear `czech`)
- Test `DELETE /api/cards/:id` — soft-deletes a card; verify it no longer appears in `GET /api/cards`
- Test `POST /api/cards/:id/suspend` and `POST /api/cards/:id/restore` — verify suspended cards don't appear in study
- Test `GET /api/tags` — returns distinct tag list
- Test `PUT /api/tags` — rename a tag; verify cards are updated
- Test `DELETE /api/tags/:tag` — deletes a tag and soft-deletes all cards with that tag

## Implementation Notes

- Use `net/http/httptest` and `gin.SetMode(gin.TestMode)` for test requests
- Parse JSON responses into structs or `map[string]interface{}` for assertions
- Each test function should call `setupTestRouter()` for a fresh database — no shared state between tests
- Run with `go test ./internal/handlers/`
