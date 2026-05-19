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
	Index           int
	Name            string
	MemoryTotalMB   int
	MemoryUsedMB    int
	MemoryFreeMB    int
	TemperatureC    int
	FanSpeed        int
	Utilization     int
	PowerDrawW      int
	PowerLimitW     int
	PState          string
	DisplayActive   bool
	ComputeCap      string
	PCILinkWidthCur int
	PCILinkWidthMax int
	PCILinkGenCur   int
	PCILinkGenMax   int
	PCIBusID        string
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
		"memory.used",
		"memory.free",
		"temperature.gpu",
		"fan.speed",
		"utilization.gpu",
		"power.draw",
		"power.limit",
		"pstate",
		"display_active",
		"compute_cap",
		"pcie.link.width.current",
		"pcie.link.width.max",
		"pcie.link.gen.current",
		"pcie.link.gen.max",
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
		memTotal, _ := strconv.Atoi(strings.TrimSpace(fields[2]))
		memUsed, _ := strconv.Atoi(strings.TrimSpace(fields[3]))
		memFree, _ := strconv.Atoi(strings.TrimSpace(fields[4]))
		temp, _ := strconv.Atoi(strings.TrimSpace(fields[5]))
		fan, _ := strconv.Atoi(strings.TrimSpace(fields[6]))
		util, _ := strconv.Atoi(strings.TrimSpace(fields[7]))
		powerDrawF, _ := strconv.ParseFloat(strings.TrimSpace(fields[8]), 32)
		powerLimitF, _ := strconv.ParseFloat(strings.TrimSpace(fields[9]), 32)
		pstate := strings.TrimSpace(fields[10])
		computeCap := strings.TrimSpace(fields[12])
		linkWidthCur, _ := strconv.Atoi(strings.TrimSpace(fields[13]))
		linkWidthMax, _ := strconv.Atoi(strings.TrimSpace(fields[14]))
		linkGenCur, _ := strconv.Atoi(strings.TrimSpace(fields[15]))
		linkGenMax, _ := strconv.Atoi(strings.TrimSpace(fields[16]))
		busID := strings.TrimSpace(fields[17])

		gpus = append(gpus, GPU{
			Index:           idx,
			Name:            cleanName(fields[1]),
			MemoryTotalMB:   memTotal,
			MemoryUsedMB:    memUsed,
			MemoryFreeMB:    memFree,
			TemperatureC:    temp,
			FanSpeed:        fan,
			Utilization:     util,
			PowerDrawW:      int(powerDrawF),
			PowerLimitW:     int(powerLimitF),
			PState:          pstate,
			DisplayActive:   strings.TrimSpace(strings.ToLower(fields[11])) == "enabled",
			ComputeCap:      computeCap,
			PCILinkWidthCur: linkWidthCur,
			PCILinkWidthMax: linkWidthMax,
			PCILinkGenCur:   linkGenCur,
			PCILinkGenMax:   linkGenMax,
			PCIBusID:        busID,
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

// MemoryDisplay returns "used/total GB" (e.g. "22/24 GB").
func (g GPU) MemoryDisplay() string {
	usedGB := g.MemoryUsedMB / 1024
	totalGB := g.MemoryTotalMB / 1024
	return fmt.Sprintf("%d/%d GB", usedGB, totalGB)
}

// FanDisplay returns fan speed with a percentage sign.
func (g GPU) FanDisplay() string {
	return fmt.Sprintf("%d%%", g.FanSpeed)
}

// PStateDisplay returns the performance state string.
func (g GPU) PStateDisplay() string {
	if g.PState == "" {
		return "N/A"
	}
	return g.PState
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

// PCIDisplay returns the current/max PCIe link info (e.g. "x16 Gen4 / x16 Gen4").
func (g GPU) PCIDisplay() string {
	cur := fmt.Sprintf("x%d Gen%d", g.PCILinkWidthCur, g.PCILinkGenCur)
	max := fmt.Sprintf("x%d Gen%d", g.PCILinkWidthMax, g.PCILinkGenMax)
	if cur == "x0 Gen0" {
		cur = "N/A"
	}
	if max == "x0 Gen0" {
		max = "N/A"
	}
	return fmt.Sprintf("%s / %s", cur, max)
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
