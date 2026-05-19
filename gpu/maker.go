package gpu

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// DetectMakers runs lspci for each GPU and returns a map from GPU index to
// normalized maker name. On lspci failure the map entries remain empty strings
// (caller should treat "" as "Unknown").
func DetectMakers(gpus []GPU) (map[int]string, error) {
	cache := make(map[int]string, len(gpus))

	for _, g := range gpus {
		busID := g.BusID()
		maker, err := lookupMaker(busID)
		if err != nil {
			// Return partial results; caller decides whether to abort.
			return cache, err
		}
		cache[g.Index] = maker
	}

	return cache, nil
}

func lookupMaker(busID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "lspci", "-s", busID, "-v")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	subsystem := extractSubsystem(string(out))
	return normalizeMaker(subsystem), nil
}

// extractSubsystem finds the "Subsystem:" line in lspci -v output.
func extractSubsystem(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Subsystem:") {
			// "Subsystem: ASUSTek Computer, inc. Device 314b (rev ...)"
			parts := strings.SplitN(line, "Subsystem:", 2)
			if len(parts) < 2 {
				continue
			}
			// Strip trailing stuff after the device code if desired, but keep
			// the raw string for normalizeMaker to match against.
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}

// normalizeMaker maps a raw lspci subsystem string to a clean vendor name.
func normalizeMaker(raw string) string {
	lower := strings.ToLower(raw)

	switch {
	case strings.Contains(lower, "asus") || strings.Contains(lower, "asustek"):
		return "ASUS"
	case strings.Contains(lower, "zotac"):
		return "ZOTAC"
	case strings.Contains(lower, "msi"):
		return "MSI"
	case strings.Contains(lower, "gigabyte"):
		return "GIGABYTE"
	case strings.Contains(lower, "evga"):
		return "EVGA"
	case strings.Contains(lower, "pny"):
		return "PNY"
	case strings.Contains(lower, "inno3d"):
		return "INNO3D"
	case strings.Contains(lower, "palit"):
		return "PALIT"
	case strings.Contains(lower, "colorful"):
		return "COLORFUL"
	case strings.Contains(lower, "msi"):
		return "MSI"
	case strings.Contains(lower, "sapphire"):
		return "SAPPHIRE"
	case strings.Contains(lower, "powercolor"):
		return "POWERCOLOR"
	case strings.Contains(lower, "xfx"):
		return "XFX"
	case strings.Contains(lower, "asrock"):
		return "ASROCK"
	default:
		if raw != "" {
			// Truncate to first 10 chars as a fallback.
			words := strings.Fields(raw)
			if len(words) > 0 {
				return truncate(words[0], 10)
			}
		}
		return "Unknown"
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
