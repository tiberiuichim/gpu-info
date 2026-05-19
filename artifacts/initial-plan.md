# gpu-info — Initial Plan

## Goal

Replace the Bash `gpu-info.sh` script with a Go CLI tool that gathers NVIDIA GPU info and renders it as styled markdown in the terminal — no `glow` piping needed.

## Architecture

```
gpu-info/
├── main.go              # CLI entry point, flag parsing
├── gpu/
│   ├── nvidia.go        # nvidia-smi JSON parsing, GPU struct
│   └── maker.go         # lspci subsystem → maker normalization
├── render/
│   ├── markdown.go      # Build markdown table + footnotes
│   └── glamour.go       # glamour TermRenderer setup
├── cmd/
│   └── root.go          # cobra/root command (if needed) or flag set
├── go.mod
├── go.sum
├── artifacts/
│   └── initial-plan.md  # this file
└── README.md
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/charmbracelet/glamour` | Markdown → ANSI renderer |
| `github.com/charmbracelet/glow/v2/utils` | `GlamourStyle()` for named style resolution |
| `github.com/charmbracelet/lipgloss` | `ColorProfile()` detection |

All from the Charm ecosystem — lightweight, no heavy frameworks needed. Plain `flag` or minimal `cobra` for CLI flags.

## GPU Data Collection

### nvidia-smi JSON output (primary source)

Instead of parsing CSV (fragile), use `nvidia-smi --query-gpu=... --format=csv,noheader,nounits` or better yet `nvidia-smi -q -x` (XML) or `nvidia-smi --query-gpu=... --format=json` (if available). The most reliable approach:

```bash
nvidia-smi --query-gpu=index,name,memory.total,temperature.gpu,utilization.gpu,power.draw,power.limit,display_active,pci.bus_id,driver_version --format=csv,noheader,nounits
```

Parse the CSV output into a struct per GPU.

### lspci (maker detection)

Run `lspci -s <bus_id> -v` and grep the `Subsystem:` line, then normalize against the same vendor map as the Bash script (ASUS, ZOTAC, MSI, GIGABYTE, EVGA, PNY, INNO3D, PALIT, COLORFUL).

## Markdown Output

Replicate the script's markdown structure:

```markdown
## 🖥️ GPU Overview

| GPU | Model | Maker | VRAM | Temp | Util | Power | Display |
|:-:|:--|:-:|:-:|:-:|:-:|:--|:-:|
| **0** | RTX 4090 | ASUS | 24 GB | 42°C 🟢 | 3% | 18W / 450W | |

> **Driver:** `535.154.05`
```

Temperature indicator logic (preserved from script):
- ≥80°C → 🔴
- ≥70°C → 🟡
- <70°C → 🟢

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--style` | `dark` | Glamour style: `auto`, `dark`, `light`, `pink`, `notty`, `dracula`, `tokyo-night` |
| `--width` | `0` (auto) | Terminal width for word wrap (0 = auto-detect) |
| `--help` | — | Show usage |

## Implementation Phases

### Phase 1: Project scaffolding
- Initialize `go.mod` with module path `github.com/<user>/gpu-info`
- Create directory structure
- Wire up `main.go` with flag parsing

### Phase 2: GPU data collection
- Implement `gpu/nvidia.go`: run `nvidia-smi` CSV query, parse into `[]GPUInfo`
- Implement `gpu/maker.go`: run `lspci`, normalize vendor string
- Handle edge cases: no GPUs, nvidia-smi not found, lspci failure → graceful errors

### Phase 3: Markdown rendering
- Implement `render/markdown.go`: build markdown table from `[]GPUInfo`
- Implement `render/glamour.go`: create `TermRenderer` with `glow/v2/utils.GlamourStyle()`
- Pipe rendered ANSI to stdout

### Phase 4: Polish
- Error handling and user-friendly messages
- `--help` flag with usage info
- README.md
- Test on actual hardware

## Key Design Decisions

1. **glamour directly, not glow's TUI** — The user wants a one-shot render (like the script piped to glow), not an interactive pager. Use `glamour.NewTermRenderer` directly.

2. **glow/v2/utils for style resolution** — Leverage `utils.GlamourStyle(style, false)` to get the same named styles that `glow --style dark` supports.

3. **CSV parsing over JSON/XML** — `nvidia-smi --format=csv,noheader,nounits` works across all driver versions and avoids XML parsing overhead. One command, one struct per line.

4. **Single-command nvidia-smi call** — Instead of N+1 calls (one per GPU per metric), batch all queries into a single `nvidia-smi` invocation for performance.

5. **Plain `flag` package** — No need for cobra/urfave/cli for 2 flags. Keep it minimal.
