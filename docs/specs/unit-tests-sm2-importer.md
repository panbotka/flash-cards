# Unit Tests for SM-2 Algorithm and Importer

Add Go table-driven tests for the two pure-function packages: `internal/srs` and `internal/importer`.

## Requirements

### SM-2 Tests (`internal/srs/sm2_test.go`)

- Test `ProcessReview` for each card status (`new`, `learning`, `review`) with each rating (Again=1, Hard=2, Good=3, Easy=4)
- Verify status transitions: new→learning, learning→learning, learning→review (graduation), review→learning (lapse)
- Verify ease factor adjustments: decreases on Again/Hard (clamped to 1.3 minimum), increases on Easy, unchanged on Good
- Verify interval calculations: learning steps (1min, 10min), graduating interval (1 day), easy graduating interval (4 days), review interval scaling
- Verify learning step progression: Again resets to step 0, Hard stays on current step, Good advances step, Easy graduates immediately
- Test `FormatInterval` returns human-readable strings (e.g., "1m", "10m", "1d", "4d")
- Test `clampInterval` bounds: minimum 1 day, maximum 365 days, rounds to 1 decimal
- Test `formatDays` output: "1d", "3d", "2w", "3mo", "1.0y"
- Test `formatMinutes` output: "1m", "10m", "2h", "1d"

### Importer Tests (`internal/importer/importer_test.go`)

- Test `DetectDelimiter` with tab-separated, semicolon-separated, " - " separated, " = " separated, and comma-separated content
- Test `DetectDelimiter` defaults to tab when no delimiter matches
- Test `Parse` produces correct `ImportCard` slices for each delimiter type
- Test `Parse` skips empty lines and lines that don't split into exactly two parts
- Test `Parse` trims whitespace from both sides of each field
- Test `Parse` skips lines where either field is empty after trimming

## Implementation Notes

- Use standard `testing` package with table-driven tests (`[]struct{ name string; ... }`)
- For SM-2 time-dependent tests, construct `SRSState` structs directly with known values rather than relying on `time.Now()`
- Run with `go test ./internal/srs/ ./internal/importer/`
