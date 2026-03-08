# Cram Mode

Study cards outside the SRS schedule — drill all cards in a tag regardless of due dates.

## Requirements

- New query parameter `?mode=cram` on `GET /api/study/next` bypasses `next_review` date filtering
- Returns a random card from the filtered set (respects existing `?direction=` and `?tag=` filters)
- Cram reviews do NOT update `srs_state` — the SRS schedule stays untouched
- Cram reviews are still logged in `review_events` (with a flag or separate type to distinguish them)
- When all cards in the filtered set have been shown, return the "no cards due" response

## UI Details

- Toggle switch or button on the Study page to enter/exit cram mode
- Visual indicator that cram mode is active (e.g. badge or different header color)
- Rating buttons and swipe gestures work the same as normal study
- Direction and tag filters remain functional in cram mode
