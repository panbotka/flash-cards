# Per-Card Review History

View the full review timeline for a specific card to see rating patterns and progress over time.

## Requirements

- New endpoint `GET /api/cards/:id/history` returns all `review_events` for that card, ordered by `reviewed_at` descending
- Each event includes: rating, direction, timestamp, interval before/after, ease before/after
- Response groups or labels events by direction (`cz_en` / `en_cz`)

## UI Details

- Tapping a card on the Cards page opens a detail/history view
- Show a chronological list of reviews with date, direction, and rating (color-coded: red/yellow/green)
- Show current SRS state summary for each direction (status, interval, ease, next review date)
- Empty state for cards with no reviews yet: "No reviews yet"
- Back button or gesture to return to the card list
