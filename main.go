package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tiberiuichim/gpu-info/gpu"
	"github.com/tiberiuichim/gpu-info/render"
)

func main() {
	style := flag.String("style", "dark", "Glamour style: auto, dark, light, pink, notty, dracula, tokyo-night")
	width := flag.Int("width", 0, "Terminal width for word wrap (0 = auto-detect)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "gpu-info — GPU status rendered as styled markdown\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	gpus, sw, err := gpu.Query()
	if err != nil {
		fmt.Fprintf(os.Stderr, "gpu-info: %v\n", err)
		os.Exit(1)
	}

	makerCache, err := gpu.DetectMakers(gpus)
	if err != nil {
		// lspci failure is non-fatal; makers will show as "Unknown"
		fmt.Fprintf(os.Stderr, "gpu-info: lspci: %v (proceeding)\n", err)
	}

	md := render.BuildMarkdown(gpus, makerCache, sw)

	output, err := render.Render(md, *style, *width)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gpu-info: render: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}
