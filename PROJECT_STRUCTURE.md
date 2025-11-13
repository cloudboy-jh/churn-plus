# Churn-Plus Project Structure

## Overview
Churn-Plus is a complete rewrite of Churn in Go, featuring a BubbleTea TUI with multi-pass LLM analysis.

## Directory Structure

```
churn-plus/
├── cmd/
│   └── churn-plus/
│       └── main.go                 # CLI entry point
│
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration management (JSON)
│   │
│   ├── engine/
│   │   ├── context.go              # Project context builder
│   │   ├── diff.go                 # Diff engine
│   │   ├── factory.go              # Component factory
│   │   ├── findings.go             # Findings aggregation
│   │   ├── pipeline.go             # Pipeline orchestrator
│   │   ├── prompts.go              # Prompt generation & parsing
│   │   ├── provider.go             # Provider interface re-export
│   │   ├── scanner.go              # File scanner
│   │   ├── types.go                # Core data types
│   │   │
│   │   ├── languages/
│   │   │   └── typescript.go       # Language-specific rules
│   │   │
│   │   ├── passes/
│   │   │   ├── lint.go             # Pass 1: Lint/Sanity
│   │   │   ├── refactor.go         # Pass 2: Structural Refactor
│   │   │   ├── local.go            # Pass 3: Local Refinement
│   │   │   └── summary.go          # Pass 4: Summary
│   │   │
│   │   └── providers/
│   │       ├── provider.go         # Provider interface
│   │       ├── anthropic.go        # Claude provider
│   │       ├── openai.go           # GPT provider
│   │       ├── google.go           # Gemini provider
│   │       └── ollama.go           # Ollama provider (uses `ollama list`)
│   │
│   ├── theme/
│   │   └── theme.go                # Lipgloss theme (ported from original Churn)
│   │
│   └── ui/
│       ├── app.go                  # Main BubbleTea app
│       └── panes/
│           ├── filetree.go         # File tree pane
│           ├── codeview.go         # Code preview pane
│           ├── findings.go         # Findings pane
│           └── pipeline.go         # Pipeline status pane
│
├── go.mod                          # Go module definition
├── go.sum                          # Go dependencies lock
├── Makefile                        # Build automation
├── README.md                       # Project documentation
├── MIGRATION.md                    # Migration guide from Churn 1.x/2.x
├── LICENSE                         # MIT License
├── .gitignore                      # Git ignore rules
└── churn-plus.exe                  # Built binary (11MB)
```

## Component Breakdown

### 1. **cmd/churn-plus/** - CLI Entry Point
- **main.go**: Argument parsing, configuration loading, engine initialization, TUI launch

### 2. **internal/config/** - Configuration System
- **config.go**: Global (`~/.churn/config.json`) and project (`.churn/config.json`) config management
- Supports JSON format matching original Churn
- Environment variable overrides for API keys

### 3. **internal/engine/** - Analysis Engine

#### Core Modules
- **types.go**: Data structures (Finding, Pipeline, Pass, ProjectContext, etc.)
- **factory.go**: Creates providers, pipelines, scans projects
- **scanner.go**: File traversal, language detection, file tree building
- **context.go**: Project context detection (frameworks, languages, tools)
- **pipeline.go**: Orchestrates multi-pass execution, emits events
- **prompts.go**: Builds prompts for each file/pass, parses LLM responses
- **findings.go**: Aggregates, deduplicates, sorts findings
- **diff.go**: Generates unified diffs between original and suggested code
- **provider.go**: Re-exports provider types to avoid import cycles

#### Subpackages
- **languages/**: Language-specific analysis rules (currently TypeScript)
- **passes/**: 4 default pass definitions (lint, refactor, local, summary)
- **providers/**: LLM provider implementations
  - Unified interface for OpenAI, Anthropic, Google, Ollama
  - Ollama uses `ollama list` to detect local models

### 4. **internal/theme/** - Branding & Styling
- **theme.go**: Lipgloss color palette, ASCII logo, styles
- Ported from original Churn (red gradient, box drawing characters)
- Moved to separate package to avoid import cycles

### 5. **internal/ui/** - BubbleTea TUI

#### Main App
- **app.go**: BubbleTea model with 4-pane layout, focus management, event handling

#### Panes
- **filetree.go**: Displays project file tree with navigation
- **codeview.go**: Shows file content with syntax highlighting markers
- **findings.go**: Lists findings with severity icons and filtering
- **pipeline.go**: Real-time pipeline status with progress indicators

## Key Features Implemented

✅ **Multi-Pass Pipeline**: 4 configurable passes (Lint → Refactor → Local → Summary)  
✅ **Multi-Provider Support**: OpenAI, Anthropic, Google, Ollama  
✅ **4-Pane TUI**: File tree, code preview, findings, pipeline status  
✅ **Real-Time Updates**: Pipeline events stream to UI via channels  
✅ **Configuration System**: JSON format in `.churn/` directory  
✅ **Project Scanning**: Language detection, framework detection, ignore patterns  
✅ **Findings Management**: Aggregation, deduplication, severity filtering  
✅ **Diff Engine**: Unified diff generation  
✅ **Report Generation**: JSON reports saved to `.churn/reports/`  
✅ **Branding**: Original Churn theme ported to Lipgloss  

## Architecture Patterns

- **Factory Pattern**: Component creation (providers, pipelines)
- **Strategy Pattern**: Different analysis passes
- **Observer Pattern**: Pipeline events → UI updates
- **Adapter Pattern**: Unified LLM interface across providers
- **Elm Architecture**: BubbleTea Model-View-Update pattern

## Build & Run

```bash
# Build
make build

# Run locally
make run

# Build for all platforms
make build-all

# Install to $GOPATH/bin
make install
```

## Configuration

### Global: `~/.churn/config.json`
```json
{
  "api_keys": {
    "anthropic": "sk-ant-...",
    "openai": "sk-..."
  },
  "default_model": {
    "provider": "anthropic",
    "model": "claude-3.5-sonnet"
  }
}
```

### Project: `.churn/config.json`
```json
{
  "ignore_patterns": ["node_modules", ".git", "dist"],
  "model": {
    "provider": "anthropic",
    "model": "claude-3.5-sonnet"
  }
}
```

## Future Enhancements (Roadmap to Churn 3.0)

- [ ] Enhanced syntax highlighting (Chroma integration)
- [ ] Search within findings (`/` key)
- [ ] Apply suggestions with confirmation
- [ ] Plugin system for custom passes
- [ ] More language-specific modules (Python, Rust, Go)
- [ ] Comprehensive test suite
- [ ] CI/CD pipeline
- [ ] Binary releases (GitHub Releases)
- [ ] Homebrew tap / AUR package

## Technical Notes

### Import Cycle Resolution
- Moved `ModelProvider` interface to `providers` package
- Created `theme` package separate from `ui` to break cycles
- Pass definitions inline in factory (removed separate `passes` imports)

### Dependencies
- **bubbletea**: TUI framework (Elm architecture)
- **lipgloss**: Styling and layout
- Go standard library for everything else (no heavy dependencies)

### Compatibility
- Config format: 100% compatible with Churn 1.x/2.x
- Directory structure: Uses same `.churn/` pattern
- Reports: Extended JSON format (backward compatible)

---

**Status**: Alpha (v0.1.0) - Builds successfully, ready for testing  
**Target**: Future Churn 3.0 release
