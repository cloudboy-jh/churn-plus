# Migration Guide: Churn 1.x/2.x → Churn-Plus

This guide helps you migrate from the original Churn (Bun/Ink CLI) to Churn-Plus (Go/BubbleTea TUI).

## Key Differences

| Feature | Churn 1.x/2.x | Churn-Plus |
|---------|---------------|------------|
| **Runtime** | Bun/Node.js | Go (compiled binary) |
| **UI** | Ink (React for terminal) | BubbleTea (Elm architecture) |
| **Interface** | Single-run commands | Persistent 4-pane TUI |
| **Analysis** | Single prompt | Multi-pass pipeline (4 passes) |
| **Config Format** | JSON | JSON (same format) |
| **Config Location** | `~/.churn/config.json` | `~/.churn/config.json` (same) |
| **Reports** | `.churn/reports/` | `.churn/reports/` (same) |

## What Stays the Same

✅ Configuration file format (JSON)  
✅ Directory structure (`.churn/`)  
✅ API key environment variables  
✅ Ignore patterns  
✅ Report format (extended with pipeline info)  
✅ Multi-provider support (OpenAI, Anthropic, Google, Ollama)

## Migration Steps

### 1. Install Churn-Plus

```bash
# Option A: From source
git clone https://github.com/cloudboy-jh/churn-plus.git
cd churn-plus
go build -o churn-plus ./cmd/churn-plus

# Option B: Using Go
go install github.com/cloudboy-jh/churn-plus/cmd/churn-plus@latest
```

### 2. Check Your Config

Your existing `~/.churn/config.json` should work as-is, but verify the structure:

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

**New fields** (optional, will use defaults if omitted):

```json
{
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

### 3. Update Environment Variables (if needed)

Churn-Plus uses the same environment variables:

```bash
export ANTHROPIC_API_KEY="your-key-here"
export OPENAI_API_KEY="your-key-here"
export GOOGLE_API_KEY="your-key-here"
```

### 4. Run Churn-Plus

```bash
cd your-project
churn-plus
```

The TUI will launch and start the analysis pipeline automatically.

### 5. Review Reports

Reports are saved in the same location:

```
.churn/reports/churn-report-2025-01-15T14-30-00.json
```

The format is extended with pipeline information but remains JSON-compatible.

## Command Mapping

| Churn 1.x/2.x | Churn-Plus | Notes |
|---------------|------------|-------|
| `churn start` | `churn-plus` | Launches TUI instead of interactive menu |
| `churn run` | `churn-plus` | Pipeline runs automatically |
| `churn model` | Edit config | Model selection via config file |
| `churn review` | TUI Findings pane | Navigate findings in real-time |
| `churn export` | Auto-saved | Reports saved automatically to `.churn/reports/` |
| `churn ask [question]` | Not yet implemented | Coming in future release |

## Keybindings

### Churn 1.x/2.x (Ink UI)
- Arrow keys to navigate menus
- Enter to select
- Ctrl+C to quit

### Churn-Plus (BubbleTea TUI)
- `Tab` - Cycle focus between panes
- `h/j/k/l` or arrow keys - Navigate within panes
- `Space` - Trigger actions (file selection, etc.)
- `q` - Quit
- `g` - Jump to top (in code view)
- `G` - Jump to bottom (in code view)

## Breaking Changes

⚠️ **Removed Commands**:
- `churn ask` - Not yet implemented (planned for future release)
- `churn pass --to <llm>` - Not applicable (multi-pass is built-in)

⚠️ **Behavioral Changes**:
- Analysis runs automatically on startup (no manual trigger needed)
- Pipeline executes all 4 passes sequentially (vs single pass in v1.x/2.x)
- TUI is persistent (doesn't exit after one run)

## Compatibility Matrix

| Churn Version | Config Compatible | Reports Compatible | Side-by-side Install |
|---------------|-------------------|---------------------|----------------------|
| 1.x | ✅ Yes | ✅ Yes (reads old reports) | ✅ Yes (different binaries) |
| 2.x | ✅ Yes | ✅ Yes (reads old reports) | ✅ Yes (different binaries) |

## Rollback Plan

If you need to rollback to Churn 1.x/2.x:

1. Your config in `~/.churn/config.json` remains intact
2. Remove Churn-Plus-specific fields (`concurrency`, `cache`, `ui`) if needed
3. Old reports in `.churn/reports/` are still valid
4. Reinstall Churn 1.x/2.x:

```bash
npm install -g churn  # or bun/pnpm/yarn
```

## FAQ

### Q: Can I run both Churn 1.x and Churn-Plus side-by-side?

**A:** Yes! They use different binary names (`churn` vs `churn-plus`) and share the same config/reports directory peacefully.

### Q: Will my old reports work with Churn-Plus?

**A:** Yes, Churn-Plus can read old report files. New reports include additional pipeline metadata but maintain backward compatibility.

### Q: Do I need to reconfigure API keys?

**A:** No, if you're using environment variables (`ANTHROPIC_API_KEY`, etc.), they work identically. Config file API keys also work the same way.

### Q: What if I don't want the multi-pass pipeline?

**A:** Currently, all 4 passes run automatically. Single-pass mode is planned for a future release. For now, you can configure which model each pass uses to control cost/speed.

### Q: Can I still use Ollama?

**A:** Yes! Ollama support is built-in. Churn-Plus will automatically detect models via `ollama list`.

### Q: When will Churn-Plus become Churn 3.0?

**A:** Once Churn-Plus reaches feature parity with Churn 2.x and passes stability testing, it will be renamed to `churn` and released as 3.0. The old CLI will be archived.

## Support

- **Issues**: https://github.com/cloudboy-jh/churn-plus/issues
- **Discussions**: https://github.com/cloudboy-jh/churn-plus/discussions
- **Original Churn**: https://github.com/cloudboy-jh/churn

---

**Migration Timeline**:
- **Now**: Alpha testing with Churn-Plus
- **Q2 2025**: Beta release with feature parity
- **Q3 2025**: Churn-Plus → Churn 3.0 (planned)
