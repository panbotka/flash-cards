# Per-Tag Stats

Break down study statistics by tag to identify which vocabulary areas need the most work.

## Requirements

- New endpoint `GET /api/stats/tags` returns per-tag metrics
- Metrics per tag: total cards, total reviews, accuracy (fraction 0-1), maturity distribution (new/learning/young/mature counts)
- Joins through `card_tags` table to associate reviews with tags
- Tags sorted by total cards descending

## UI Details

- New section on the Stats page showing a tag breakdown table or card list
- Each tag row shows: tag name, card count, accuracy percentage, maturity bar (colored segments)
- Tapping a tag could filter the existing stats charts to that tag (stretch goal — not required for MVP)
