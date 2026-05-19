package render

import (
	"fmt"
	"strings"

	"github.com/tiberiuichim/gpu-info/gpu"
)

// BuildMarkdown produces the markdown table from GPU data.
func BuildMarkdown(gpus []gpu.GPU, makerCache map[int]string, sw gpu.SoftwareInfo) string {
	var sb strings.Builder

	sb.WriteString("## 🖥️ GPU Overview\n\n")

	// Table header
	sb.WriteString("| GPU | Model | Maker | VRAM | Temp | Util | Power | Display |\n")
	sb.WriteString("|:-:|:--|:-:|:-:|:-:|:-:|:--|:-:|\n")

	for _, g := range gpus {
		maker := makerCache[g.Index]
		if maker == "" {
			maker = "Unknown"
		}

		sb.WriteString(fmt.Sprintf(
			"| **%d** | %s | %s | %s | %s | %s | %s | %s |\n",
			g.Index,
			g.Name,
			maker,
			g.MemoryGB(),
			g.TemperatureBadge(),
			g.UtilizationDisplay(),
			g.PowerDisplay(),
			g.DisplayBadge(),
		))
	}

	sb.WriteString("\n## 📦 Software\n\n")
	sb.WriteString(fmt.Sprintf("| Component | Version |\n"))
	sb.WriteString("|:-|:-:|\n")
	sb.WriteString(fmt.Sprintf("| **Driver** | `%s` |\n", sw.Driver))
	sb.WriteString(fmt.Sprintf("| **CUDA** | `%s` |\n", sw.CUDA))
	sb.WriteString(fmt.Sprintf("| **nvcc** | `%s` |\n", sw.NVCC))

	return sb.String()
}
