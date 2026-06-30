# Shadowbox UI Redesign
### Music Acquisition Console

## Design Goals

- Preserve the retro terminal aesthetic.
- Replace the single-column menu with a pane-based interface.
- Make navigation faster.
- Display useful information at all times.
- Feel similar to modern TUIs like lazygit, btop, and cmus.

---

# Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ SHADOWBOX                                          Music Acquisition Console│
├──────────────┬──────────────────────────────────────────────┬───────────────┤
│ Navigation   │ Search / Library / Downloads                 │ Details       │
│              │                                              │               │
│ > Search     │ Query: nujabes                              │ Queue: 2      │
│   URL        │───────────────────────────────────────────── │ Speed:18 MB/s │
│   Playlist   │ ► Feather — Nujabes                         │ FLAC          │
│   Library    │   Reflection Eternal                        │ YouTube Music │
│   Downloads  │   Luv(sic) Pt.3                             │               │
│   Enhance    │                                              │ Metadata      │
│   Settings   │                                              │               │
├──────────────┴──────────────────────────────────────────────┴───────────────┤
│ TAB Switch Pane • / Search • Enter Download • q Quit • Queue:2 • Cache:94GB│
└─────────────────────────────────────────────────────────────────────────────┘
```

---

# Navigation

```
Search
URL Download
Playlist Download

Library
Downloads

Enhance

Settings
```

The active selection should be cyan while inactive items remain gray.

---

# Search Screen

```
Search

Query:
> nujabes

Results

► Feather
  Reflection Eternal
  Luv(sic) Pt.3
  Aruarian Dance

Enter  Download
Space  Preview
```

---

# Library Screen

```
Artists        Albums             Tracks

Nujabes        Modal Soul         Feather
J Dilla        Donuts             Workinonit
Madlib         Piñata             Thuggin
MF DOOM        MM..FOOD           One Beer
```

---

# Downloads Screen

```
Downloads

███████████████████░░░░░ 72%

Current
-----------------------------------
Feather.mp3

Speed
18 MB/s

ETA
00:13

Queue

1. Reflection Eternal
2. Luv(sic) Pt.3
```

---

# Details Pane

Display contextual information depending on the current view.

Search

- Source
- Duration
- Bitrate
- Album
- Artist

Downloads

- Progress
- Speed
- ETA
- Threads

Library

- Album Art (optional sixel/chafa)
- Year
- Genre
- Track Count

---

# Status Bar

```
LIBRARY | Queue:2 | Downloads:1 | Cache:94GB | TAB Pane | / Search | q Quit
```

---

# Color Palette

Magenta
- Branding
- Logo

Cyan
- Selected items
- Borders
- Cursor

White
- Primary text

Gray
- Secondary text

Green
- Success
- Completed downloads

Yellow
- Warnings

Red
- Errors

---

# Improvements over Current UI

✓ Smaller logo after startup

✓ Three-pane layout

✓ Dedicated workspace for every mode

✓ Status bar

✓ Better keyboard discoverability

✓ Live queue information

✓ Better visual hierarchy

✓ Easier navigation

✓ Scales well to larger terminals

---

# Inspiration

- lazygit
- btop
- cmus
- Spotify TUI
- ranger