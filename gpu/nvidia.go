package gpu

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SoftwareInfo holds NVIDIA software version info (driver, CUDA runtime, nvcc).
type SoftwareInfo struct {
	Driver string
	CUDA   string
	NVCC   string
}

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
// GPU slice and NVIDIA software version info (driver, CUDA, nvcc).
func Query() ([]GPU, SoftwareInfo, error) {
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
		return nil, SoftwareInfo{}, fmt.Errorf("nvidia-smi: %w", err)
	}

	lines := splitTrimLines(string(out))
	if len(lines) == 0 {
		return nil, SoftwareInfo{}, fmt.Errorf("nvidia-smi returned no output")
	}

	gpus := make([]GPU, 0, len(lines))
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) < len(attrs) {
			return nil, SoftwareInfo{}, fmt.Errorf("nvidia-smi: expected %d fields, got %d in line: %q", len(attrs), len(fields), line)
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

	// Software versions — separate lightweight calls.
	sw := SoftwareInfo{}

	driver, err := queryDriver(ctx)
	if err != nil {
		driver = "unknown"
	}
	sw.Driver = driver

	cuda, err := queryCUDA(ctx)
	if err != nil {
		cuda = "unknown"
	}
	sw.CUDA = cuda

	nvcc, err := queryNVCC(ctx)
	if err != nil {
		nvcc = "not installed"
	}
	sw.NVCC = nvcc

	return gpus, sw, nil
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

// TemperatureBadge returns the temperature string with a bullet indicator.
// Colors are applied post-render since glamour strips ANSI codes from markdown.
func (g GPU) TemperatureBadge() string {
	return fmt.Sprintf("● %d°C", g.TemperatureC)
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

// queryCUDA extracts the CUDA version from the nvidia-smi header line.
// nvidia-smi --query-gpu=cuda_version is not supported, so we parse the banner.
func queryCUDA(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "nvidia-smi")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`CUDA Version:\s*([\d.]+)`)
	matches := re.FindSubmatch(out)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse CUDA version from nvidia-smi output")
	}
	return string(matches[1]), nil
}

// queryNVCC runs nvcc --version and extracts the compiler release version.
// Returns an error if nvcc is not found (CUDA Toolkit not installed).
func queryNVCC(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "nvcc", "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`release ([\d.]+)`)
	matches := re.FindSubmatch(out)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse nvcc version from output")
	}
	return string(matches[1]), nil
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
