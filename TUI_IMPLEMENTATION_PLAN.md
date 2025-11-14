# Two-Pane Horizontal TUI with Menu System & LLM Hand-Off

## Goal
Build a **persistent terminal application** that:
1. Starts with an **interactive menu** (START, MODEL SELECT, SETTINGS, EXIT)
2. Launches into a **TWO-PANE HORIZONTAL LAYOUT** (left = findings list, right = finding details)
3. Primary feature: **LLM hand-off** to automatically fix findings
4. This is a **STATEFUL TUI** that runs continuously, not a CLI enter-exit cycle

## Visual Design
- **SOLID PANES** with background colors (not just box outlines)
- **MODAL OVERLAYS** with solid backgrounds
- **Churn color scheme** from https://github.com/cloudboy-jh/churn:
  - Background: `#1b1b1b` (dark gray)
  - Primary Red: `#ff5656` (coral red)
  - Secondary Red: `#ff8585` (lighter coral)
  - Text Primary: `#f2e9e4` (cream)
  - Muted: `#a6adc8` (light purple-gray)

## Current State Analysis
- ✅ BubbleTea/Lipgloss already installed
- ✅ Theme system exists with coral red (#ff5656) branding
- ✅ Finding data structure defined (`internal/engine/types.go`)
- ✅ LLM providers exist (Anthropic, OpenAI, Google, Ollama)
- ✅ Menu system EXISTS (`internal/ui/menu.go`) - can adapt this
- ❌ NO cmd/ directory (need to create entrypoint)
- ❌ Current UI is 4-pane grid (need NEW 2-pane horizontal TUI)

## Implementation Plan

### 1. Create Command Entrypoint
**File**: `cmd/churn-plus/main.go`
- Entry point that launches BubbleTea application
- Start with menu system, transition to TUI on "START"
- Handle args like `--help`, `--version`

### 2. Build Menu System (Entry Screen)
**File**: `internal/ui/menu/menu.go` (adapt existing or create new)

**Menu UI** (with solid background):
```
 ██████   ██  ██   ██  ██   ██  ██████   ██████
██   ██   ██  ██   ██  ██   ██  ██   ██  ██   ██
██        ██████   ██  ██   ██  ██████   ██   ██
██   ██   ██  ██   ██  ██   ██  ██  ██   ██   ██
 ██████   ██  ██    ████     ██  ██       ██████

Project: /path/to/project
Latest Report: 2025-11-14T12:30:00 (15 findings)

┌─ Main Menu ─────────────────────────┐
│ ▶ START ANALYSIS                     │  ← Highlighted with coral red bg
│   MODEL SELECT                       │
│   SETTINGS                           │
│   EXIT                               │
└──────────────────────────────────────┘

↑/↓: navigate | Enter: select | q: quit
```

**Navigation**:
- `↑`/`↓` (arrow keys): Navigate menu options
- `Enter`: Select option
- `q`: Exit application

**Menu Options**:
1. **START ANALYSIS**: Load findings → launch 2-pane TUI
2. **MODEL SELECT**: Sub-menu to choose LLM provider/model
3. **SETTINGS**: View/edit config (read-only for now)
4. **EXIT**: Quit

### 3. Model Select Sub-Menu
**File**: `internal/ui/menu/model_select.go`

```
┌─ Select LLM Provider ───────────────┐
│ ▶ Anthropic (Claude)                 │  ← Solid coral red bg
│   OpenAI (GPT)                       │
│   Google (Gemini)                    │
│   Ollama (Local)                     │
│   < Back                             │
└──────────────────────────────────────┘

┌─ Select Model ──────────────────────┐
│ ▶ claude-3-5-sonnet-20241022         │  ← Solid coral red bg
│   claude-3-5-haiku-20241022          │
│   claude-3-opus-20240229             │
│   < Back                             │
└──────────────────────────────────────┘
```

- Navigate providers → select → choose specific model
- Save to `.churn/config.json`
- Return to main menu

### 4. Settings View
**File**: `internal/ui/menu/settings.go`

```
┌─ Settings ──────────────────────────┐
│ Provider: anthropic                  │
│ Model: claude-3-5-sonnet-20241022    │
│ API Key: sk-ant-*********************│
│ Cache: enabled                       │
│ Concurrency: 10                      │
│                                      │
│ Config: ~/.churn/config.json         │
│                                      │
│ Press 'q' to go back                 │
└──────────────────────────────────────┘
```

- Read-only view of current config
- `q`: Return to main menu

### 5. Build Two-Pane TUI (Analysis View)
**Files**: 
- `internal/ui/tui/model.go` - Main BubbleTea model for TUI
- `internal/ui/tui/list_pane.go` - Left pane (findings list)
- `internal/ui/tui/detail_pane.go` - Right pane (finding details + actions)

**Layout** (activated after "START ANALYSIS") - SOLID PANES:
```
┌──────────────────┬───────────────────────────────────────┐
│ FINDINGS (15)    │ FINDING DETAILS                       │
│ [dark bg]        │ [dark bg]                             │
│                  │                                       │
│ ▶ file.ts:42 🔴  │ File: src/file.ts:42-48               │
│   [coral red bg] │ Severity: HIGH 🔴                     │
│   app.go:15  🟡  │ Type: unused-import                   │
│   main.py:8  🔵  │                                       │
│   util.js:23 🟠  │ Reasoning:                            │
│                  │ This import statement is never used   │
│                  │ in the module. Removing it will...    │
│                  │                                       │
│                  │ Suggested Patch:                      │
│                  │ - import { unused } from './deps'     │
│                  │ + // removed unused import            │
│                  │                                       │
│                  │ ┌───────────────────────────────────┐ │
│                  │ │ (l) Send to LLM  ← PRIMARY ACTION │ │
│                  │ │ (p) Preview Patch                 │ │
│                  │ │ (a) Apply Patch                   │ │
│                  │ │ (m) Back to Menu                  │ │
│                  │ │ (q) Quit                          │ │
│                  │ └───────────────────────────────────┘ │
└──────────────────┴───────────────────────────────────────┘
 ↑/↓: navigate | Enter: select | l: LLM hand-off | m: menu
```

**Key Points**:
- **HORIZONTAL SPLIT**: Left 1/3 width, Right 2/3 width
- **SOLID BACKGROUNDS**: Each pane has dark background fill
- **SELECTED ITEM**: Solid coral red background
- **NO VERTICAL STACKING** (unless terminal too narrow)
- **STATEFUL**: Maintains selected finding, focus, LLM history

### 6. Implement Navigation & Keybindings
**File**: `internal/ui/tui/keys.go`

**Left Pane** (findings list):
- `↑`/`↓` (arrow keys): Navigate findings
- `Enter`: Select finding → focus right pane
- `m`: Return to menu
- `q`: Quit application

**Right Pane** (finding details):
- `l`: **Send to LLM** (PRIMARY FEATURE)
- `p`: Preview patch in modal
- `a`: Apply patch (with confirmation)
- `m`: Return to menu
- `q`: Return to left pane

**Global**:
- `Ctrl+C`: Force quit anytime

### 7. Implement LLM Hand-Off (PRIMARY FEATURE)
**File**: `internal/ui/tui/llm_handler.go`

**Flow when user presses `l`**:
1. Show modal: "🔄 Sending to LLM..." (solid background overlay)
2. Build prompt with finding context:
   ```
   You are a code fixing assistant. Fix this issue:
   
   File: {file}
   Issue: {message}
   Code: {code_snippet}
   
   Provide a patch in unified diff format.
   ```
3. Stream LLM response in real-time modal (solid background)
4. Parse response for patch
5. Update finding with LLM suggestion
6. Return to detail pane with "✅ LLM Response Received"

**LLM Response Modal** (SOLID OVERLAY):
```
      ╔═══════════════════════════════════════╗
      ║ LLM Response (streaming...)           ║
      ║ [solid dark bg with coral border]     ║
      ║                                       ║
      ║ I'll fix this unused import issue:    ║
      ║                                       ║
      ║ ```diff                               ║
      ║ - import { unused } from './deps'     ║
      ║ + // Removed unused import            ║
      ║ ```                                   ║
      ║                                       ║
      ║ This change is safe because the       ║
      ║ import is never referenced...         ║
      ║                                       ║
      ║ Press 'a' to apply | 'q' to close     ║
      ╚═══════════════════════════════════════╝
```

### 8. Implement Patch System
**File**: `internal/ui/tui/patch_engine.go`

**Preview Patch Modal** (`p` key) - SOLID OVERLAY:
```
      ╔═══════════════════════════════════════╗
      ║ Patch Preview: src/file.ts            ║
      ║ [solid dark bg]                       ║
      ║                                       ║
      ║ @@ -1,5 +1,4 @@                       ║
      ║  import { useState } from 'react';    ║
      ║ -import { unused } from './deps';     ║
      ║  export function App() {              ║
      ║    const [state, setState] = ...;     ║
      ║                                       ║
      ║ Press 'a' to apply | 'q' to close     ║
      ╚═══════════════════════════════════════╝
```

**Apply Patch** (`a` key) - SOLID OVERLAY:
```
      ╔═══════════════════════════════════════╗
      ║ Confirm Apply Patch                   ║
      ║ [solid dark bg]                       ║
      ║                                       ║
      ║ This will modify: src/file.ts         ║
      ║ A backup will be saved to:            ║
      ║   src/file.ts.bak                     ║
      ║                                       ║
      ║ Continue? (y/n)                       ║
      ╚═══════════════════════════════════════╝
```

- Create `.bak` file before applying
- Write changes to disk
- Show success/error message

### 9. State Management & Transitions
**File**: `internal/ui/app_state.go`

**States**:
1. `StateMenu` - Main menu
2. `StateModelSelect` - Model selection sub-menu
3. `StateSettings` - Settings view
4. `StateTUI` - Two-pane analysis view
5. `StateLLMModal` - LLM response streaming
6. `StatePatchPreview` - Patch preview modal

**Transitions**:
- Menu → TUI (START selected)
- Menu → ModelSelect (MODEL SELECT selected)
- Menu → Settings (SETTINGS selected)
- TUI → Menu (`m` key)
- TUI → LLMModal (`l` key)
- Any state → Exit (`q` key in menu, Ctrl+C anywhere)

### 10. Responsive Layout
- **Left pane**: `width / 3` (minimum 30 chars)
- **Right pane**: `2 * width / 3`
- If `width < 100`: Show warning or stack vertically (last resort)
- Update on `tea.WindowSizeMsg`

### 11. Styling with Lipgloss (SOLID BACKGROUNDS)
Use Churn color scheme:
```go
// Background fills
paneBg := lipgloss.NewStyle().
    Background(lipgloss.Color("#1b1b1b")).
    Foreground(lipgloss.Color("#f2e9e4"))

// Selected item (SOLID coral red background)
selectedBg := lipgloss.NewStyle().
    Background(lipgloss.Color("#ff5656")).
    Foreground(lipgloss.Color("#f2e9e4")).
    Bold(true)

// Modal overlay (SOLID)
modalBg := lipgloss.NewStyle().
    Background(lipgloss.Color("#1b1b1b")).
    Foreground(lipgloss.Color("#f2e9e4")).
    BorderStyle(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#ff5656")).
    Padding(1, 2)
```

**Colors**:
- Background: `#1b1b1b` (dark gray) - SOLID fill
- Primary Red: `#ff5656` (coral red) - Selected items, borders
- Secondary Red: `#ff8585` (lighter coral) - Accents
- Text Primary: `#f2e9e4` (cream) - Main text
- Muted: `#a6adc8` (light purple-gray) - Help text
- **Severity icons**: 🔴 Critical, 🟠 High, 🟡 Medium, 🔵 Low

### 12. Documentation
**Update**: `README.md`

Add sections:
```markdown
## Interactive TUI

Launch the terminal UI:
```bash
churn-plus
```

### Main Menu
- **START ANALYSIS**: Load findings and enter interactive mode
- **MODEL SELECT**: Choose LLM provider and model
- **SETTINGS**: View current configuration
- **EXIT**: Quit application

### Two-Pane Analysis View
- **Left Pane**: Navigate findings with arrow keys
- **Right Pane**: View details and take actions

### Keybindings
| Key | Action |
|-----|--------|
| ↑/↓ | Navigate up/down |
| Enter | Select item |
| l | Send to LLM (primary feature) |
| p | Preview patch |
| a | Apply patch |
| m | Return to menu |
| q | Quit/back |
| Ctrl+C | Force quit |

### LLM Hand-Off Workflow
1. Navigate to a finding with arrow keys
2. Press `l` to send to LLM
3. Watch streaming response in modal
4. Press `a` to apply suggested fix
5. Patch is applied with backup (.bak file)
```

### 13. Testing & Git
- Create feature branch: `feature/tui-menu-and-two-pane`
- Unit tests for:
  - Menu navigation logic
  - State transitions
  - LLM prompt building
  - Patch parsing
- Integration test: Menu → TUI → LLM → Patch
- Commit message: "feat: interactive menu + two-pane horizontal TUI with LLM hand-off"

## File Structure
```
cmd/churn-plus/
  main.go                    (NEW - CLI entrypoint)

internal/ui/
  app_state.go               (NEW - state machine)
  
internal/ui/menu/
  menu.go                    (NEW - main menu)
  model_select.go            (NEW - model selection)
  settings.go                (NEW - settings view)

internal/ui/tui/
  model.go                   (NEW - two-pane TUI model)
  list_pane.go               (NEW - left findings list)
  detail_pane.go             (NEW - right detail view)
  llm_handler.go             (NEW - LLM integration)
  patch_engine.go            (NEW - patch preview/apply)
  keys.go                    (NEW - keybinding definitions)
  modals.go                  (NEW - modal overlays with solid bg)

README.md                    (UPDATE - document TUI)
```

## Key Design Decisions
1. ✅ **MENU FIRST** - Interactive menu is entry point
2. ✅ **ARROW KEY NAVIGATION** - Use ↑/↓ instead of j/k
3. ✅ **SOLID PANES** - Background fills, not just outlines
4. ✅ **SOLID MODALS** - Overlay modals with solid backgrounds
5. ✅ **HORIZONTAL LAYOUT** - Left-right split, NOT vertical
6. ✅ **STATEFUL APPLICATION** - Runs continuously with state machine
7. ✅ **LLM HAND-OFF PRIMARY** - `l` key is the main workflow
8. ✅ **MODEL SELECTION** - In-app UI to choose provider/model
9. ✅ **CHURN COLORS** - Use color scheme from original Churn
10. ✅ **NO FILE EXPLORER** - Simple two-pane design only
11. ✅ **BACKUP BEFORE APPLY** - Always create .bak files

## Implementation Order
1. Create `cmd/churn-plus/main.go` entrypoint
2. Build menu system with solid backgrounds and arrow key navigation
3. Implement model select sub-menu
4. Build two-pane TUI layout with solid backgrounds
5. Add LLM hand-off functionality with modal overlay
6. Implement patch preview/apply with solid modals
7. Wire up state transitions
8. Add documentation
9. Test & commit

## Color Reference (from Churn)
```
Background:   #1b1b1b (dark gray)
Primary Red:  #ff5656 (coral red)
Secondary Red:#ff8585 (lighter coral)
Text Primary: #f2e9e4 (cream)
Muted:        #a6adc8 (light purple-gray)
Info:         #8ab4f8 (blue)
Success:      #a6e3a1 (green)
Warning:      #f9e2af (yellow)
Error:        #f38ba8 (pink-red)
```

Ready to implement! 🚀
