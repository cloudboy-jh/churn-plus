# Churn-Plus

> The next-generation evolution of Churn: A TUI-based code analysis tool powered by multi-pass LLM pipelines.

**Status**: Alpha (v0.1.0) - Future Churn 3.0

Churn-Plus is a complete architectural rewrite of [Churn](https://github.com/cloudboy-jh/churn) built in Go with a persistent, multi-pane terminal interface. It will eventually replace the current Bun/Ink CLI as the canonical Churn experience.

## Features

- **4-Pane TUI**: File tree, code preview, findings, and pipeline status in a single view
- **Multi-Pass Analysis**: Lint → Refactor → Local Refinement → Summary
- **Multi-Model Support**: OpenAI (GPT), Anthropic (Claude), Google (Gemini), Ollama (local)
- **Structured Findings**: Severity levels, categorization, and actionable recommendations
- **Real-Time Updates**: Watch pipeline progress as it analyzes your code
- **Local-First**: Uses `.churn/` directory pattern, fully compatible with original Churn
- **Privacy Option**: Ollama integration for 100% local analysis

## Installation

### From Source

```bash
git clone https://github.com/cloudboy-jh/churn-plus.git
cd churn-plus
go build -o churn-plus ./cmd/churn-plus
```

### Using Go Install

```bash
go install github.com/cloudboy-jh/churn-plus/cmd/churn-plus@latest
```

## Quick Start

1. **Set up API keys** (or use Ollama for local models):

```bash
export ANTHROPIC_API_KEY="your-key-here"
# OR
export OPENAI_API_KEY="your-key-here"
# OR install Ollama from https://ollama.ai
```

2. **Run Churn-Plus** in your project directory:

```bash
churn-plus
```

You'll see an interactive menu:
```
[ASCII Logo]

Project: /your/project
Files: 234 | Languages: TypeScript, Python | Frameworks: React

Current Model Pipeline:
  1. Lint/Sanity: claude-3.5-haiku (Anthropic)
  2. Refactor: claude-3.5-sonnet (Anthropic)
  3. Summary: claude-3.5-sonnet (Anthropic)

┌─ Menu ──────────────────────────┐
│ > Start Analysis          ENTER │
│   Configure Model Pipeline      │
│   Settings                      │
│   Exit                     ESC  │
└─────────────────────────────────┘
```

3. **Start Analysis**:
   - Press `ENTER` to start
   - Or navigate with `↑/↓` arrows

4. **Navigate the TUI** (after starting):
   - `Tab` - Cycle focus between panes
   - `h/j/k/l` or arrow keys - Navigate
   - `q` - Quit

**Quick run (skip menu)**:
```bash
churn-plus --run
```

## Configuration

### Global Config: `~/.churn/config.json`

```json
{
  "api_keys": {
    "anthropic": "sk-ant-...",
    "openai": "sk-...",
    "google": "..."
  },
  "default_model": {
    "provider": "anthropic",
    "model": "claude-3.5-sonnet"
  },
  "concurrency": {
    "ollama": 20,
    "openai": 8,
    "anthropic": 10,
    "google": 8
  },
  "cache": {
    "enabled": true,
    "ttl": 24,
    "max_size": 100
  },
  "ui": {
    "show_line_numbers": true,
    "syntax_highlight": true,
    "theme": "default"
  }
}
```

### Project Config: `.churn/config.json`

```json
{
  "last_run": "2025-01-15T14:30:00Z",
  "model": {
    "provider": "anthropic",
    "model": "claude-3.5-sonnet"
  },
  "ignore_patterns": [
    "node_modules",
    ".git",
    "dist",
    "build"
  ],
  "pipeline": {
    "passes": [
      {
        "name": "lint",
        "description": "Quick structural checks for unused code",
        "enabled": true,
        "model": "claude-3-5-haiku-20241022",
        "provider": "anthropic"
      },
      {
        "name": "refactor",
        "description": "Deep analysis for architectural improvements",
        "enabled": true,
        "model": "claude-3.5-sonnet",
        "provider": "anthropic"
      },
      {
        "name": "summary",
        "description": "Coherence check and overall assessment",
        "enabled": true,
        "model": "claude-3.5-sonnet",
        "provider": "anthropic"
      }
    ]
  }
}
```

You can now configure your pipeline using the interactive menu or by editing the config file directly!

## Architecture

Churn-Plus is built on three core layers:

1. **TUI Layer** (BubbleTea): 4-pane layout with real-time updates
2. **Analysis Engine**: Project scanning, context building, pipeline orchestration
3. **Model Providers**: Unified interface for OpenAI, Anthropic, Google, Ollama

See [ARCHITECTURE.md](../CHURN-PLUS_FULL_ARCHITECTURE.md) for complete details.

## Multi-Pass Pipeline

1. **Pass 1: Lint/Sanity** - Fast structural checks (unused code, basic issues)
2. **Pass 2: Structural Refactor** - Deep analysis for architectural improvements
3. **Pass 3: Local Refinement** - Optional Ollama pass for validation
4. **Pass 4: Consistency & Summary** - Ensures coherence across findings

## Reports

Analysis reports are saved to `.churn/reports/` as timestamped JSON files:

```
.churn/
├── config.json
└── reports/
    ├── churn-report-2025-01-15T14-30-00.json
    └── churn-report-2025-01-15T16-45-22.json
```

## Migrating from Churn 1.x/2.x

Churn-Plus uses the same `.churn/` directory structure as the original Churn, so migration is seamless:

1. Your existing `.churn/config.json` will work as-is
2. Old reports are preserved in `.churn/reports/`
3. API keys from environment variables work identically

See [MIGRATION.md](./MIGRATION.md) for detailed migration guide.

## Development

```bash
# Install dependencies
go mod tidy

# Run locally
go run ./cmd/churn-plus

# Build
go build -o churn-plus ./cmd/churn-plus

# Run tests
go test ./...
```

## Roadmap to Churn 3.0

- [x] Go/BubbleTea foundation
- [x] 4-pane TUI
- [x] Multi-pass pipeline
- [x] Multi-provider LLM support
- [x] Interactive menu with pipeline configuration
- [x] Settings submenu for viewing configuration
- [x] Pipeline configuration persistence
- [ ] Enhanced syntax highlighting (Chroma integration)
- [ ] Search within findings (`/` key)
- [ ] Apply suggestions (with confirmation)
- [ ] Plugin system for custom passes
- [ ] Performance optimizations
- [ ] Comprehensive test suite
- [ ] Release as Churn 3.0

## Contributing

Contributions welcome! Please open an issue or PR.

## License

MIT License - see LICENSE file

## Credits

- Original Churn by [@cloudboy-jh](https://github.com/cloudboy-jh)
- Built with [BubbleTea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss)

---

**Note**: This is the future of Churn. Once stable, it will be renamed to `churn` and replace the current CLI.
