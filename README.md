# gpu-info

GPU status at a glance — rendered as styled markdown in your terminal.

![screenshot](screenshot.png)

## What it does

Replaces the classic `nvidia-smi | glow` pipeline with a single native Go binary:

```bash
# Instead of:
nvidia-smi | glow --style dark

# Just:
gpu-info
```

Outputs three sections:

### GPU Overview

| Column | Source |
|--------|--------|
| GPU | nvidia-smi index |
| Model | GPU name (stripped of "NVIDIA GeForce" prefix) |
| Mem | Used / total memory in GB |
| Temp | Temperature with bullet indicator (🟢 <70°C, 🟡 70-79°C, 🔴 ≥80°C) |
| Fan | Fan speed % |
| Util | GPU utilization % |
| PState | Performance state (P0 = max perf, P12 = min perf) |
| Power | Current draw / power limit in watts |

### Hardware

| Column | Source |
|--------|--------|
| Maker | lspci subsystem vendor (ASUS, MSI, ZOTAC, etc.) |
| Compute Cap | NVIDIA compute capability (e.g. 8.6 = Ampere) |
| Display | 🖥️ if display is active |
| PCIe | `pcie.link.width.current` / `pcie.link.width.max` (e.g. "x8 Gen4 / x16 Gen4") |

### Software

| Component | Source |
|-----------|--------|
| Driver | `nvidia-smi` driver version |
| CUDA | CUDA runtime version (max supported by driver) |
| nvcc | CUDA compiler version ("not installed" if CUDA Toolkit absent) |

## Installation

```bash
go install github.com/tiberiuichim/gpu-info@latest
```

Or build from source:

```bash
git clone https://github.com/tiberiuichim/gpu-info.git
cd gpu-info
go build -o gpu-info .
sudo mv gpu-info /usr/local/bin/
```

## Usage

```
gpu-info [flags]

Flags:
  -style string
        Glamour style: auto, dark, light, pink, notty, dracula, tokyo-night (default "dark")
  -width int
        Terminal width for word wrap (0 = auto-detect)
```

### Examples

```bash
# Default (dark style)
gpu-info

# Light style
gpu-info -style light

# Tokyo Night theme
gpu-info -style tokyo-night

# Force width for logging
gpu-info -width 120
```

## Requirements

- **Linux** (uses `nvidia-smi` and `lspci`)
- **NVIDIA GPU** with driver installed
- Go 1.22+ (to build from source)

## How it works

1. Single `nvidia-smi --query-gpu=...` call fetches all metrics at once (no N+1 queries)
2. `lspci -s <bus_id> -v` detects the board vendor per GPU
3. Markdown table is built in-memory
4. [glamour](https://github.com/charmbracelet/glamour) renders to ANSI output

## License

MIT
