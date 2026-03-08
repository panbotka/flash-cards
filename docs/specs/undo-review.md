# Undo Last Review

Let users undo their most recent rating to fix accidental swipes or mis-taps during study.

## Requirements

- New endpoint `POST /api/study/undo` reverts the most recent review for the current direction and tag filter
- Delete the `review_events` row and restore `srs_state` using the `_before` columns (`interval_before`, `ease_before`, etc.)
- Return the restored card so the UI can re-display it for re-rating
- Only the single most recent review is undoable — no multi-step undo
- If there is no review to undo (session just started), return 404

## UI Details

- Undo button appears on the Study page after a rating is submitted
- Button disappears when the next card is rated (only the latest review is undoable)
- Tapping undo restores the previously rated card in its revealed (flipped) state
- Works for both tap ratings and swipe ratings
