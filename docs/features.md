# Features Overview

## Spaced Repetition (SM-2)

The app uses a modified SM-2 algorithm to schedule card reviews at optimal intervals.

### Card States

```
New → Learning → Review
         ↑          |
         └──────────┘  (on "Again" rating)
```

- **New**: Never reviewed. Enters learning on first review.
- **Learning**: Working through learning steps with sub-day intervals.
- **Review**: Graduated to SRS. Intervals measured in days.

### Learning Steps

New cards go through two learning steps before entering the review queue:

| Step | Interval |
|------|----------|
| 1 | 1 minute |
| 2 | 10 minutes |
| Graduate | 1 day |
| Easy graduate | 4 days (skips learning) |

### Ratings

| Rating | Name | Learning Effect | Review Effect |
|--------|------|-----------------|---------------|
| 1 | Again | Reset to step 1 (1m). Ease -0.20 | Return to learning at step 1. Ease -0.20 |
| 2 | Hard | Stay on current step. Ease -0.15 | Interval × 1.2. Ease -0.15 |
| 3 | Good | Advance to next step (or graduate). Ease unchanged | Interval × ease factor. Ease unchanged |
| 4 | Easy | Graduate immediately (4d). Ease +0.15 | Interval × ease × 1.3. Ease +0.15 |

### Constraints

- Minimum ease factor: 1.3
- Minimum review interval: 1 day
- Maximum review interval: 365 days

---

## Bidirectional Cards

Each card creates two independent SRS entries:

- **Czech → English** (`cz_en`): Shows Czech word, expects English answer
- **English → Czech** (`en_cz`): Shows English word, expects Czech answer

Each direction has its own ease factor, interval, and review schedule. Study sessions can be filtered by direction.

---

## Card Management

- **Tags**: Cards can have multiple tags. Tags are used to filter study sessions and the card list.
- **Soft delete**: Deleted cards are hidden but not destroyed. They can be restored.
- **Suspend**: Temporarily exclude a card from study without deleting it.
- **Search**: Filter the card list by Czech or English text (case-insensitive).

---

## Bulk Import

Import cards from text with automatic delimiter detection.

**Supported delimiters** (detected automatically):
- Tab (`\t`)
- Semicolon (`;`)
- Space-dash-space (` - `)
- Space-equals-space (` = `)
- Comma (`,`)

**Workflow:**
1. Paste text or upload a file (.csv, .tsv, .txt)
2. Preview parsed cards with duplicate detection
3. Optionally assign tags to the entire batch
4. Confirm import — duplicates are skipped silently

Duplicates are detected by case-insensitive match on both Czech and English text against existing non-deleted cards.

---

## Statistics

### Summary
- Reviews completed today
- Streak (consecutive days with reviews)
- Total cards in collection
- Today's accuracy (% of Good + Easy ratings)

### Activity Heatmap
GitHub-style 365-day grid showing daily review counts. Color intensity scales with activity.

### Accuracy Chart
Line chart of daily accuracy percentage over the last 30 days.

### Maturity Distribution
Breakdown of cards by learning stage:

| Category | Criteria |
|----------|----------|
| New | Never reviewed |
| Learning | In learning steps |
| Young | In review with interval < 21 days |
| Mature | In review with interval ≥ 21 days |

Each card is classified by its worst direction.

### Forecast
Bar chart predicting how many reviews are due each day for the next 30 days.

### Hardest Cards
Table of cards with the lowest accuracy (most "Again" ratings), for cards with at least 3 reviews.

---

## Authentication

- Password set via `APP_PASSWORD` environment variable
- Session stored in an HTTP-only cookie (HMAC-SHA256 signed), valid for 30 days
- If `APP_PASSWORD` is not set, the app runs without authentication (development mode)

---

## Keyboard Shortcuts

Available on the study page:

| Key | Action |
|-----|--------|
| Space | Flip card (reveal answer) |
| 1 | Rate: Hard |
| 2 | Rate: Good |
| 3 | Rate: Easy |

Shortcuts are disabled when focus is on an input field.

## Swipe Gestures

On mobile, swipe gestures provide an alternative to buttons:

| Direction | Rating |
|-----------|--------|
| Left | Hard |
| Up | Good |
| Right | Easy |

Swipe works immediately without needing to flip the card first. Visual overlay shows the rating color during swipe.
