package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func identifySize(path string) (int, int, error) {
	out, err := exec.Command("magick", "identify", "-format", "%w %h", path).Output()
	if err != nil {
		return 0, 0, fmt.Errorf("magick identify %q: %w", path, err)
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected identify output: %q", string(out))
	}
	w, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse width: %w", err)
	}
	h, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse height: %w", err)
	}
	return w, h, nil
}

func ratioClose(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

func runMagick(args ...string) error {
	cmd := exec.Command("magick", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func cropResizeBlurBg(inputPath, outputPath string, width, height int, mode string, blurRadius int) error {
	if _, err := os.Stat(inputPath); err != nil {
		return fmt.Errorf("input not found: %s", inputPath)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case "blur", "cover", "contain":
	default:
		return fmt.Errorf("mode must be one of: blur, cover, contain")
	}

	srcW, srcH, err := identifySize(inputPath)
	if err != nil {
		return err
	}

	srcRatio := float64(srcW) / float64(srcH)
	tgtRatio := float64(width) / float64(height)
	target := fmt.Sprintf("%dx%d", width, height)

	if ratioClose(srcRatio, tgtRatio, 0.01) || mode == "cover" {
		return runMagick(
			inputPath,
			"-resize", target+"^",
			"-gravity", "center",
			"-extent", target,
			outputPath,
		)
	}

	if mode == "contain" {
		return runMagick(
			inputPath,
			"-resize", target,
			"-gravity", "center",
			"-background", "black",
			"-extent", target,
			outputPath,
		)
	}

	// blur mode: cover-blurred background + contained foreground composited on top
	return runMagick(
		inputPath,
		"(", "+clone",
		"-resize", target+"^",
		"-gravity", "center",
		"-extent", target,
		"-blur", fmt.Sprintf("0x%d", blurRadius),
		")",
		"(", "+clone",
		"-resize", target,
		")",
		"-delete", "0",
		"-gravity", "center",
		"-compose", "over",
		"-composite",
		outputPath,
	)
}
