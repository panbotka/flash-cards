# Export Cards to CSV

Export all cards as a CSV file for backup or migration to other tools (Anki, Quizlet).

## Requirements

- New endpoint `GET /api/cards/export` returns a CSV file download
- CSV columns: `czech`, `english`, `tags` (space-delimited, matching import format)
- Include all non-deleted cards, sorted by ID
- Response uses `Content-Disposition: attachment; filename="flash-cards.csv"` and `Content-Type: text/csv`
- Exported CSV should be re-importable via the existing import feature

## UI Details

- Export button on the Cards page (e.g. in the header area near the search bar)
- Tapping the button triggers a browser file download — no preview or confirmation needed
