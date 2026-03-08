# Daily Study Goal

Set a target number of reviews per day and show progress on the Study page.

## Requirements

- New `settings` table in SQLite (key-value: `key TEXT PRIMARY KEY, value TEXT`)
- New endpoints: `GET /api/settings/daily-goal` and `PUT /api/settings/daily-goal` (body: `{ "goal": 50 }`)
- Default goal: 0 (disabled — no progress indicator shown)
- Today's review count is already available from the `/stats/summary` endpoint

## UI Details

- Progress ring or bar on the Study page showing reviews completed today vs. the goal
- When goal is reached, show a completion indicator (e.g. checkmark, color change)
- Settings accessible from the Study page (e.g. tap the progress ring to edit, or a small gear icon)
- When goal is set to 0 or unset, hide the progress indicator entirely
