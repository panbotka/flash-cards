# UI/UX Specification

## Design System

### Theme

Dark mode only.

| Token | Value | Usage |
|-------|-------|-------|
| Surface | `#0a0a0a` | Page background |
| Surface Raised | `#141414` | Elevated elements |
| Surface Card | `#1a1a1a` | Cards, inputs, modals |
| Surface Hover | `#222222` | Hover states |
| Border Subtle | `#2a2a2a` | Card borders, dividers |
| Border Default | `#333333` | Input focus borders |
| Text Primary | `#f5f5f7` | Headings, body text |
| Text Secondary | `#a1a1a6` | Labels, subtitles |
| Text Tertiary | `#6e6e73` | Hints, disabled text |
| Accent | `#5e9eff` | Active tabs, buttons, links |

### Rating Colors

| Rating | Color | Usage |
|--------|-------|-------|
| Again | `#ff453a` | Red — forgot |
| Hard | `#ff9f0a` | Orange — difficult |
| Good | `#30d158` | Green — normal |
| Easy | `#5e9eff` | Blue — easy |

### Typography

System font stack: `-apple-system, BlinkMacSystemFont, SF Pro Display, SF Pro Text, system-ui, sans-serif`

- Card words: `text-3xl font-semibold`
- Section titles: `text-xs uppercase tracking-wider` in tertiary color
- Tag chips: `text-xs`
- Nav labels: `text-[10px] font-medium`

### Components

- **Cards/surfaces**: `bg-[#1a1a1a] rounded-xl border border-[#2a2a2a]`
- **Inputs**: Same as cards, `text-center` for login, left-aligned elsewhere
- **Buttons**: Rounded-xl, full-width where appropriate, `active:scale-95` press feedback
- **Pills/chips**: `rounded-full px-3 py-1 text-xs`
- **Modals**: Dark backdrop with `bg-black/60 backdrop-blur-sm`, content with card styling

---

## Pages

### Login Page

Full-screen centered layout, no navigation bar.

- "Flash Cards" title (large, bold, white)
- "Czech — English" subtitle (gray)
- Password input (centered, rounded, dark surface)
- "Sign In" button (accent blue, full-width)
- Error message in red below input on failure
- Auto-focuses password field on mount

### Study Page (default route: `/`)

The primary learning interface.

**Layout:**
```
┌─────────────────────┐
│     [tag chips]      │
│                      │
│  ┌────────────────┐  │
│  │                │  │
│  │   word text    │  │
│  │                │  │
│  │  Tap to reveal │  │
│  └────────────────┘  │
│                      │
│  [Again][Hard][Good][Easy]  │  ← visible after flip
│                      │
│  Space=flip  1-4=rate│  ← desktop only
│                      │
├─────────────────────┤
│ Study  Cards  Stats  │
└─────────────────────┘
```

**Flash Card:**
- 3D flip animation (CSS `perspective` + `rotateY(180deg)`, 600ms transition)
- Front: prompt word (large centered text) + "Tap to reveal" hint
- Back: answer word + direction label at top (e.g., "Czech → English")
- Click/tap anywhere on card to flip
- Min height 280px, max width md, rounded-2xl

**Rating Buttons:**
- 4-column grid below the card
- Each button: rating name + next interval preview (e.g., "Good" / "1d")
- Semi-transparent background with colored border matching rating color
- Slide-up animation on reveal (opacity + translateY transition)
- Disabled until card is flipped

**Done State:**
- "All caught up!" message centered
- Count of new cards available
- "Continue with new cards" button (accent) + "Done for today" button (subtle)

**Loading:**
- Pulsing skeleton card (animate-pulse)

### Cards Page (`/cards`)

Card collection management.

**Layout:**
```
┌─────────────────────┐
│ Cards               │
│ [🔍 Search cards...]│
│ [All][food][transport]│
│ 5 cards             │
│ ┌─────────────────┐ │
│ │ kolo        [⋮] │ │
│ │ bike             │ │
│ │ [transport] [New]│ │
│ └─────────────────┘ │
│ ┌─────────────────┐ │
│ │ ...              │ │
│ └─────────────────┘ │
│                 [+]  │  ← FAB
├─────────────────────┤
│ Study  Cards  Stats  │
└─────────────────────┘
```

**Card Rows:**
- Czech text (white, medium weight)
- English text (gray, smaller)
- Tag pills + status badge (New=blue, Learning=orange, Review=green)
- Suspended cards: dimmed (opacity-50) + red "Suspended" badge
- Three-dot menu with: Edit, Suspend/Unsuspend, Delete (red)

**Add/Edit Modal:**
- Backdrop overlay with blur
- Fields: Czech, English, Tags (comma-separated)
- Auto-focus on Czech field
- Save and Cancel buttons

**Floating Action Button:**
- Bottom-right, accent blue circle with "+" icon
- Positioned above the bottom nav (`bottom-24`)

### Import Page (`/import`)

Bulk card import with preview.

**Layout:**
```
┌─────────────────────┐
│ Import              │
│ Paste tab-separated…│
│ ┌─────────────────┐ │
│ │ pes  dog        │ │
│ │ kočka  cat      │ │
│ │                 │ │
│ └─────────────────┘ │
│ [Upload file]        │
│ Tags: [lesson-1]     │
│ [    Preview    ]    │
├─────────────────────┤
│ Study  Cards  Stats  │
└─────────────────────┘
```

**After Preview:**
```
┌─────────────────────┐
│ 3 cards, 1 duplicate │
│ ✓ pes / dog         │
│ ✓ kočka / cat       │
│ ⚠ kolo / bike [Dup] │
│ [Back] [Import 2]   │
└─────────────────────┘
```

- Textarea: min-height 200px, monospace font, dark surface
- File upload: accepts .csv, .tsv, .txt
- Preview button: accent blue, full-width
- Duplicate rows: orange warning icon + "Duplicate" badge
- Import button: green, shows count of non-duplicate cards
- Success banner after import

### Stats Page (`/stats`)

Analytics dashboard. All sections stack vertically with `gap-4`.

**Sections:**

1. **Overview** — 2×2 grid of summary cards (Reviews Today, Streak, Total Cards, Accuracy Today)
2. **Activity** — GitHub-style heatmap (53 weeks × 7 days), blue color scale
3. **Accuracy (30 Days)** — Line chart, accent blue line, Y-axis 0-100%
4. **Maturity** — Horizontal bar chart (New/Learning/Young/Mature)
5. **Forecast** — Vertical bar chart of due reviews for next 30 days
6. **Hardest Cards** — Table with Czech, English, Reviews, Accuracy %

Each section: titled in uppercase tracking-wider tertiary text, wrapped in card surface.

---

## Navigation

Bottom tab bar, fixed at viewport bottom.

| Tab | Route | Icon |
|-----|-------|------|
| Study | `/` | Book/card |
| Cards | `/cards` | List/stack |
| Stats | `/stats` | Bar chart |

- Active tab: accent blue icon + label
- Inactive: tertiary gray
- Height: 64px + `env(safe-area-inset-bottom)` for iOS
- Background: `#0a0a0a` with top border `#2a2a2a`
- Hidden on login page

---

## Responsive Design

- Mobile-first: designed for 375px (iPhone SE) and up
- Touch targets: minimum 44px
- No hover-dependent interactions
- Cards and forms use `max-w-md` for readable width on larger screens
- Bottom nav is thumb-accessible
