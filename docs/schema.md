# Database Schema

SQLite with WAL journal mode, foreign keys enabled, 5-second busy timeout.

## Tables

### cards

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| id | INTEGER | AUTOINCREMENT | Primary key |
| czech | TEXT | NOT NULL | Czech word/phrase |
| english | TEXT | NOT NULL | English word/phrase |
| deleted_at | DATETIME | NULL | Soft delete timestamp |
| suspended | BOOLEAN | FALSE | Excluded from study when true |
| created_at | DATETIME | CURRENT_TIMESTAMP | Creation time |
| updated_at | DATETIME | CURRENT_TIMESTAMP | Last update time |

### card_tags

| Column | Type | Description |
|--------|------|-------------|
| card_id | INTEGER | FK → cards(id) |
| tag | TEXT | Tag name |

Primary key: `(card_id, tag)`

### srs_state

Each card has two rows — one per direction.

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| id | INTEGER | AUTOINCREMENT | Primary key |
| card_id | INTEGER | NOT NULL | FK → cards(id) |
| direction | TEXT | NOT NULL | `cz_en` or `en_cz` |
| ease_factor | REAL | 2.5 | SM-2 ease factor (min 1.3) |
| interval_days | REAL | 0 | Current interval in days |
| repetitions | INTEGER | 0 | Successful review count |
| next_review | DATETIME | CURRENT_TIMESTAMP | When this card is next due |
| status | TEXT | `new` | `new`, `learning`, or `review` |
| learning_step | INTEGER | 0 | Index into learning steps array |

Unique constraint: `(card_id, direction)`

### review_events

Audit log of every review.

| Column | Type | Description |
|--------|------|-------------|
| id | INTEGER | Primary key |
| srs_state_id | INTEGER | FK → srs_state(id) |
| card_id | INTEGER | FK → cards(id) |
| direction | TEXT | `cz_en` or `en_cz` |
| rating | INTEGER | 1 (Again), 2 (Hard), 3 (Good), 4 (Easy) |
| reviewed_at | DATETIME | Review timestamp |
| interval_before | REAL | Interval before this review |
| interval_after | REAL | Interval after this review |
| ease_before | REAL | Ease factor before |
| ease_after | REAL | Ease factor after |

## Indexes

```sql
idx_srs_next_review    ON srs_state(next_review)     -- study queue lookup
idx_srs_card_dir       ON srs_state(card_id, direction)
idx_review_events_date ON review_events(reviewed_at)  -- stats queries
idx_cards_deleted      ON cards(deleted_at)           -- soft delete filter
```

## Migrations

Schema is versioned via a `schema_version` table. Migrations run automatically on startup inside transactions. Current version: 1.
