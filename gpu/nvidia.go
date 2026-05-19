package gpu

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// GPU holds info gathered from a single GPU.
type GPU struct {
	Index         int
	Name          string
	MemoryTotalMB int
	TemperatureC  int
	Utilization   int
	PowerDrawW    int
	PowerLimitW   int
	DisplayActive bool
	PCIBusID      string
}

// Query runs nvidia-smi once to gather all GPU metrics, returning the parsed
// GPU slice and the driver version string.
func Query() ([]GPU, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Single nvidia-smi invocation: one row per GPU, all metrics at once.
	attrs := []string{
		"index",
		"name",
		"memory.total",
		"temperature.gpu",
		"utilization.gpu",
		"power.draw",
		"power.limit",
		"display_active",
		"pci.bus_id",
	}
	queryArgs := []string{
		"--query-gpu=" + strings.Join(attrs, ","),
		"--format=csv,noheader,nounits",
	}

	cmd := exec.CommandContext(ctx, "nvidia-smi", queryArgs...)
	out, err := cmd.Output()
	if err != nil {
		return nil, "", fmt.Errorf("nvidia-smi: %w", err)
	}

	lines := splitTrimLines(string(out))
	if len(lines) == 0 {
		return nil, "", fmt.Errorf("nvidia-smi returned no output")
	}

	gpus := make([]GPU, 0, len(lines))
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) < len(attrs) {
			return nil, "", fmt.Errorf("nvidia-smi: expected %d fields, got %d in line: %q", len(attrs), len(fields), line)
		}

		idx, _ := strconv.Atoi(strings.TrimSpace(fields[0]))
		memMB, _ := strconv.Atoi(strings.TrimSpace(fields[2]))
		temp, _ := strconv.Atoi(strings.TrimSpace(fields[3]))
		util, _ := strconv.Atoi(strings.TrimSpace(fields[4]))
		powerDraw, _ := strconv.Atoi(strings.TrimSpace(fields[5]))
		powerLimit, _ := strconv.Atoi(strings.TrimSpace(fields[6]))

		busID := strings.TrimSpace(fields[8])

		gpus = append(gpus, GPU{
			Index:         idx,
			Name:          cleanName(fields[1]),
			MemoryTotalMB: memMB,
			TemperatureC:  temp,
			Utilization:   util,
			PowerDrawW:    powerDraw,
			PowerLimitW:   powerLimit,
			DisplayActive: strings.TrimSpace(strings.ToLower(fields[7])) == "enabled",
			PCIBusID:      busID,
		})
	}

	// Driver version — separate lightweight call.
	driver, err := queryDriver(ctx)
	if err != nil {
		// Non-fatal: we still have GPU data.
		driver = "unknown"
	}

	return gpus, driver, nil
}

func queryDriver(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// driver_version is identical across GPUs; take the first line.
	lines := strings.Split(string(out), "\n")
	return strings.TrimSpace(lines[0]), nil
}

// cleanName strips the "NVIDIA GeForce " / "NVIDIA " prefix from the GPU name.
func cleanName(raw string) string {
	name := strings.TrimSpace(raw)
	name = strings.TrimPrefix(name, "NVIDIA GeForce ")
	name = strings.TrimPrefix(name, "NVIDIA ")
	return name
}

// MemoryGB converts MiB to a clean GB string (e.g. 24576 → "24 GB").
func (g GPU) MemoryGB() string {
	gb := g.MemoryTotalMB / 1024
	return fmt.Sprintf("%d GB", gb)
}

// TemperatureBadge returns the temperature string with an emoji indicator.
func (g GPU) TemperatureBadge() string {
	switch {
	case g.TemperatureC >= 80:
		return fmt.Sprintf("%d°C 🔴", g.TemperatureC)
	case g.TemperatureC >= 70:
		return fmt.Sprintf("%d°C 🟡", g.TemperatureC)
	default:
		return fmt.Sprintf("%d°C 🟢", g.TemperatureC)
	}
}

// PowerDisplay returns "drawW / limitW".
func (g GPU) PowerDisplay() string {
	return fmt.Sprintf("%dW / %dW", g.PowerDrawW, g.PowerLimitW)
}

// UtilizationDisplay returns "X%".
func (g GPU) UtilizationDisplay() string {
	return fmt.Sprintf("%d%%", g.Utilization)
}

// DisplayBadge returns a monitor emoji if display is active.
func (g GPU) DisplayBadge() string {
	if g.DisplayActive {
		return "🖥️"
	}
	return ""
}

// normalizeBusID strips leading zeros from the PCI bus ID so it matches lspci output.
func (g GPU) BusID() string {
	bus := g.PCIBusID
	// nvidia-smi may output like "00000000:01:00.0" — strip leading zeros per segment.
	parts := strings.Split(bus, ":")
	for i, p := range parts {
		parts[i] = strings.TrimLeft(p, "0")
		if parts[i] == "" {
			parts[i] = "0"
		}
	}
	return strings.Join(parts, ":")
}

func splitTrimLines(s string) []string {
	lines := strings.Split(s, "\n")
	result := make([]string, 0, len(lines))
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}
