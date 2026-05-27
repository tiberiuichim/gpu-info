package render

import (
	"fmt"
	"strings"

	"github.com/tiberiuichim/gpu-info/gpu"
)

// BuildMarkdown produces the markdown table from GPU data.
func BuildMarkdown(gpus []gpu.GPU, makerCache map[int]string, sw gpu.SoftwareInfo, connBadges map[int]string) string {
	var sb strings.Builder

	sb.WriteString("## 🖥️ GPU Overview\n\n")

	// Table header — thermal / runtime metrics only
	sb.WriteString("| GPU | Card | Mem | Temp | Fan/Load | Power |\n")
	sb.WriteString("|:-:|:--|:--|:-:|:-:|:--|\n")

	for _, g := range gpus {
		card := cardLabel(g.Name, makerCache[g.Index])
		fanUtil := fmt.Sprintf("%s / %s", g.FanDisplay(), g.UtilizationDisplay())
		sb.WriteString(fmt.Sprintf(
			"| **%d** | %s | %s | %s | %s | %s |\n",
			g.Index,
			card,
			g.MemoryDisplay(),
			g.TemperatureBadge(),
			fanUtil,
			g.PowerDisplay(),
		))
	}

	sb.WriteString("\n## 🔧 Hardware\n\n")
	sb.WriteString("| GPU | Card | Compute Cap | 🖥️ | PCIe |\n")
	sb.WriteString("|:-:|:--|:-:|:-:|:--|\n")

	for _, g := range gpus {
		card := cardLabel(g.Name, makerCache[g.Index])
		cc := g.ComputeCap
		if cc == "" {
			cc = "N/A"
		}
		// Prefer sysfs-based connected badge; fall back to nvidia-smi display_active.
		displayBadge := connBadges[g.Index]
		if displayBadge == "" {
			displayBadge = g.DisplayBadge()
		}
		sb.WriteString(fmt.Sprintf(
			"| **%d** | %s | %s | %s | %s |\n",
			g.Index,
			card,
			cc,
			displayBadge,
			g.PCIDisplay(),
		))
	}

	sb.WriteString("\n## 📦 Software\n\n")
	sb.WriteString(fmt.Sprintf("| Component | Version |\n"))
	sb.WriteString("|:-|:-:|\n")
	sb.WriteString(fmt.Sprintf("| **Driver** | `%s` |\n", sw.Driver))
	sb.WriteString(fmt.Sprintf("| **CUDA Runtime** | `%s` |\n", sw.CUDA))
	sb.WriteString(fmt.Sprintf("| **CUDA Toolkit** | `%s` |\n", sw.NVCC))

	return sb.String()
}

// cardLabel returns "Model — Maker" or just "Model" when maker is unknown.
func cardLabel(model string, maker string) string {
	if maker == "" || maker == "Unknown" {
		return model
	}
	return fmt.Sprintf("%s — %s", model, maker)
}
