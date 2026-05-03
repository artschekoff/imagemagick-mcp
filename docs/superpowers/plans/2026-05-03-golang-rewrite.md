# Go Rewrite + Makefile + Skill Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite the ImageMagick MCP server from Python to Go, add a Makefile for cross-platform builds and GitHub releases, and update SKILL.md for image modification workflows.

**Architecture:** A single Go package at the repo root exposes two MCP tools via `github.com/mark3labs/mcp-go` and shells out to `magick` CLI exactly as the Python version did. The Makefile cross-compiles for four platforms and delegates releases to `gh`. SKILL.md becomes an AI-facing guide for image modification tasks.

**Tech Stack:** Go 1.22+, github.com/mark3labs/mcp-go, ImageMagick CLI (`magick`), GNU Make, GitHub CLI (`gh`)

---

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `go.mod` | Go module declaration |
| Create | `main.go` | Entry point, injects `version` |
| Create | `imagemagick.go` | `identifySize`, `ratioClose`, `cropResizeBlurBg`, `runMagick` |
| Create | `imagemagick_test.go` | Tests for `imagemagick.go` |
| Create | `server.go` | MCP server init, tool registration, handlers |
| Create | `Makefile` | build / build-all / test / clean / release targets |
| Modify | `SKILL.md` | Replace with image-modification AI guide |
| Modify | `README.md` | Update for Go binary, remove Python quick-start |
| Modify | `CLAUDE.md` | Update commands for Go toolchain |

Python files in `scripts/` are kept as-is for reference but are no longer the primary server.

---

## Task 1: Initialize Go Module

**Files:**
- Create: `go.mod`

- [ ] **Step 1: Init module and fetch dependency**

```bash
cd /path/to/imagemagick-mcp
go mod init github.com/roxl/imagemagick-mcp
go get github.com/mark3labs/mcp-go@latest
```

Expected: `go.mod` and `go.sum` created, `mcp-go` version pinned (e.g. `v0.27.0`).

- [ ] **Step 2: Verify go.mod content**

`go.mod` must look like (version may differ):
```
module github.com/roxl/imagemagick-mcp

go 1.22

require github.com/mark3labs/mcp-go v0.27.0
```

Run: `go env GOPATH && cat go.mod`
Expected: module line and require block present.

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: init go module with mcp-go dependency"
```

---

## Task 2: Implement Core ImageMagick Logic with Tests

**Files:**
- Create: `imagemagick.go`
- Create: `imagemagick_test.go`

- [ ] **Step 1: Write failing tests**

Create `imagemagick_test.go`:

```go
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// requireMagick skips the test if magick is not installed.
func requireMagick(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("magick"); err != nil {
		t.Skip("magick not in PATH — skipping integration test")
	}
}

func TestRatioClose(t *testing.T) {
	cases := []struct {
		a, b float64
		want bool
	}{
		{1.0, 1.0, true},
		{1.0, 1.005, true},   // within 0.01
		{1.0, 1.02, false},   // outside 0.01
		{16.0 / 9.0, 16.0 / 9.0, true},
		{1.333, 1.334, true}, // portrait/landscape near-match
	}
	for _, c := range cases {
		if got := ratioClose(c.a, c.b, 0.01); got != c.want {
			t.Errorf("ratioClose(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestModeValidation(t *testing.T) {
	requireMagick(t)

	// Create a 100x100 white test image using magick
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.png")
	if err := exec.Command("magick", "-size", "100x100", "xc:white", src).Run(); err != nil {
		t.Fatalf("create test image: %v", err)
	}

	dst := filepath.Join(tmp, "out.png")
	err := cropResizeBlurBg(src, dst, 200, 200, "invalid", 30)
	if err == nil {
		t.Fatal("expected error for invalid mode, got nil")
	}
}

func TestCropResizeBlurBg_SameRatio(t *testing.T) {
	requireMagick(t)

	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.png")
	if err := exec.Command("magick", "-size", "200x200", "xc:blue", src).Run(); err != nil {
		t.Fatalf("create test image: %v", err)
	}

	dst := filepath.Join(tmp, "out.png")
	if err := cropResizeBlurBg(src, dst, 100, 100, "blur", 30); err != nil {
		t.Fatalf("cropResizeBlurBg: %v", err)
	}

	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("output file not created: %v", err)
	}

	w, h, err := identifySize(dst)
	if err != nil {
		t.Fatalf("identifySize: %v", err)
	}
	if w != 100 || h != 100 {
		t.Errorf("expected 100x100, got %dx%d", w, h)
	}
}

func TestCropResizeBlurBg_BlurMode(t *testing.T) {
	requireMagick(t)

	tmp := t.TempDir()
	// 200x100 → target 100x100: aspect ratio mismatch, blur mode
	src := filepath.Join(tmp, "src.png")
	if err := exec.Command("magick", "-size", "200x100", "xc:red", src).Run(); err != nil {
		t.Fatalf("create test image: %v", err)
	}

	dst := filepath.Join(tmp, "out.png")
	if err := cropResizeBlurBg(src, dst, 100, 100, "blur", 30); err != nil {
		t.Fatalf("cropResizeBlurBg blur: %v", err)
	}

	w, h, err := identifySize(dst)
	if err != nil {
		t.Fatalf("identifySize: %v", err)
	}
	if w != 100 || h != 100 {
		t.Errorf("expected 100x100, got %dx%d", w, h)
	}
}

func TestCropResizeBlurBg_ContainMode(t *testing.T) {
	requireMagick(t)

	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.png")
	if err := exec.Command("magick", "-size", "200x100", "xc:green", src).Run(); err != nil {
		t.Fatalf("create test image: %v", err)
	}

	dst := filepath.Join(tmp, "out.png")
	if err := cropResizeBlurBg(src, dst, 100, 100, "contain", 30); err != nil {
		t.Fatalf("cropResizeBlurBg contain: %v", err)
	}

	w, h, err := identifySize(dst)
	if err != nil {
		t.Fatalf("identifySize: %v", err)
	}
	if w != 100 || h != 100 {
		t.Errorf("expected 100x100, got %dx%d", w, h)
	}
}

func TestCropResizeBlurBg_CoverMode(t *testing.T) {
	requireMagick(t)

	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.png")
	if err := exec.Command("magick", "-size", "200x100", "xc:yellow", src).Run(); err != nil {
		t.Fatalf("create test image: %v", err)
	}

	dst := filepath.Join(tmp, "out.png")
	if err := cropResizeBlurBg(src, dst, 100, 100, "cover", 30); err != nil {
		t.Fatalf("cropResizeBlurBg cover: %v", err)
	}

	w, h, err := identifySize(dst)
	if err != nil {
		t.Fatalf("identifySize: %v", err)
	}
	if w != 100 || h != 100 {
		t.Errorf("expected 100x100, got %dx%d", w, h)
	}
}

func TestCropResizeBlurBg_MissingInput(t *testing.T) {
	err := cropResizeBlurBg("/nonexistent/path.png", "/tmp/out.png", 100, 100, "blur", 30)
	if err == nil {
		t.Fatal("expected error for missing input, got nil")
	}
}

func TestCropResizeBlurBg_CreatesOutputDir(t *testing.T) {
	requireMagick(t)

	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.png")
	if err := exec.Command("magick", "-size", "50x50", "xc:white", src).Run(); err != nil {
		t.Fatalf("create test image: %v", err)
	}

	dst := filepath.Join(tmp, "subdir", "nested", "out.png")
	if err := cropResizeBlurBg(src, dst, 50, 50, "blur", 30); err != nil {
		t.Fatalf("cropResizeBlurBg: %v", err)
	}

	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("output not found at nested path: %v", err)
	}
}
```

- [ ] **Step 2: Run tests — confirm they fail**

```bash
go test ./... -v -run TestRatioClose
```

Expected: `imagemagick.go:1:1: no such file` or `undefined: ratioClose` — compilation error confirming nothing exists yet.

- [ ] **Step 3: Implement `imagemagick.go`**

Create `imagemagick.go`:

```go
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
```

- [ ] **Step 4: Run tests — confirm they pass**

```bash
go test ./... -v
```

Expected:
```
--- PASS: TestRatioClose (0.00s)
--- PASS: TestModeValidation (0.XXs)
--- PASS: TestCropResizeBlurBg_SameRatio (0.XXs)
--- PASS: TestCropResizeBlurBg_BlurMode (0.XXs)
--- PASS: TestCropResizeBlurBg_ContainMode (0.XXs)
--- PASS: TestCropResizeBlurBg_CoverMode (0.XXs)
--- PASS: TestCropResizeBlurBg_MissingInput (0.00s)
--- PASS: TestCropResizeBlurBg_CreatesOutputDir (0.XXs)
PASS
```

If `magick` is not installed, integration tests show `SKIP` — that is acceptable.

- [ ] **Step 5: Commit**

```bash
git add imagemagick.go imagemagick_test.go
git commit -m "feat: implement imagemagick core logic in go with tests"
```

---

## Task 3: Implement MCP Server

**Files:**
- Create: `server.go`

- [ ] **Step 1: Create `server.go`**

```go
package main

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const serverName = "imagemagick-mcp"

func serve() error {
	s := server.NewMCPServer(serverName, version,
		server.WithToolCapabilities(false),
	)

	s.AddTool(
		mcp.NewTool("describe_imagemagick_interface",
			mcp.WithDescription("Returns server metadata and available tools"),
		),
		describeHandler,
	)

	s.AddTool(
		mcp.NewTool("crop_resize_blur_bg",
			mcp.WithDescription("Resize/crop image to exact dimensions. If aspect ratio differs, use mode: blur|cover|contain."),
			mcp.WithString("input_path",
				mcp.Required(),
				mcp.Description("Absolute path to the source image"),
			),
			mcp.WithString("output_path",
				mcp.Required(),
				mcp.Description("Absolute path for the output image (parent dirs created automatically)"),
			),
			mcp.WithNumber("width",
				mcp.Required(),
				mcp.Description("Target width in pixels"),
			),
			mcp.WithNumber("height",
				mcp.Required(),
				mcp.Description("Target height in pixels"),
			),
			mcp.WithString("mode",
				mcp.Description("Resize mode: blur (default), cover, or contain"),
			),
			mcp.WithNumber("blur",
				mcp.Description("Blur radius for blur mode (default: 30)"),
			),
		),
		cropResizeBlurBgHandler,
	)

	return server.ServeStdio(s)
}

func describeHandler(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(
		`{"server":"imagemagick-mcp","tools":"crop_resize_blur_bg","note":"Uses ImageMagick. If aspect ratio differs, pads with blurred version of the image."}`,
	), nil
}

func cropResizeBlurBgHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments

	inputPath, _ := args["input_path"].(string)
	outputPath, _ := args["output_path"].(string)

	if inputPath == "" {
		return mcp.NewToolResultError("input_path is required"), nil
	}
	if outputPath == "" {
		return mcp.NewToolResultError("output_path is required"), nil
	}

	widthF, ok := args["width"].(float64)
	if !ok {
		return mcp.NewToolResultError("width is required and must be a number"), nil
	}
	heightF, ok := args["height"].(float64)
	if !ok {
		return mcp.NewToolResultError("height is required and must be a number"), nil
	}

	width := int(widthF)
	height := int(heightF)

	mode := "blur"
	if m, ok := args["mode"].(string); ok && m != "" {
		mode = m
	}

	blurRadius := 30
	if b, ok := args["blur"].(float64); ok {
		blurRadius = int(b)
	}

	if err := cropResizeBlurBg(inputPath, outputPath, width, height, mode, blurRadius); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf(`{"ok":"true","output_path":%q,"mode":%q}`, outputPath, mode),
	), nil
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./...
```

Expected: exits 0, no output. (`main.go` doesn't exist yet so this may error — create a stub in the next step to unblock.)

- [ ] **Step 3: Commit**

```bash
git add server.go
git commit -m "feat: add mcp server with tool registration and handlers"
```

---

## Task 4: Implement Entry Point

**Files:**
- Create: `main.go`

- [ ] **Step 1: Create `main.go`**

```go
package main

import (
	"log"
	"os"
)

// version is set at build time via -ldflags "-X main.version=<tag>"
var version = "dev"

func main() {
	if err := serve(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Build and smoke-test**

```bash
go build -o bin/imagemagick-mcp .
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./bin/imagemagick-mcp
```

Expected: JSON response listing `describe_imagemagick_interface` and `crop_resize_blur_bg`.

- [ ] **Step 3: Run full test suite**

```bash
go test ./... -v
```

Expected: all tests pass (integration tests skip if `magick` absent).

- [ ] **Step 4: Commit**

```bash
git add main.go bin/
git commit -m "feat: add entry point, wire serve() from main"
```

Note: add `bin/` to `.gitignore` if not already present:
```bash
echo 'bin/' >> .gitignore
git add .gitignore
git commit -m "chore: ignore build artifacts"
```

---

## Task 5: Create Makefile

**Files:**
- Create: `Makefile`

- [ ] **Step 1: Create `Makefile`**

```makefile
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BINARY   := imagemagick-mcp
BIN_DIR  := bin
LDFLAGS  := -ldflags "-X main.version=$(VERSION)"

.PHONY: build build-all test clean release

build:
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY) .

build-all:
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY)-darwin-arm64 .
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY)-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY)-windows-amd64.exe .

test:
	go test ./... -v

clean:
	rm -rf $(BIN_DIR)

release: build-all
	gh release create $(VERSION) $(BIN_DIR)/* --generate-notes --title "Release $(VERSION)"
```

- [ ] **Step 2: Verify `make build` works**

```bash
make build
```

Expected:
```
go build -ldflags "-X main.version=..." -o bin/imagemagick-mcp .
```
and `bin/imagemagick-mcp` exists.

- [ ] **Step 3: Verify `make test` works**

```bash
make test
```

Expected: all tests pass.

- [ ] **Step 4: Tag and dry-run release (do NOT actually publish)**

```bash
git tag v1.0.0
make build-all
ls bin/
```

Expected:
```
imagemagick-mcp-darwin-amd64
imagemagick-mcp-darwin-arm64
imagemagick-mcp-linux-amd64
imagemagick-mcp-windows-amd64.exe
```

Do NOT run `make release` unless you intend to publish. Remove the tag if testing only:
```bash
git tag -d v1.0.0
```

- [ ] **Step 5: Commit**

```bash
git add Makefile
git commit -m "chore: add makefile for build, test, cross-compile, and release"
```

---

## Task 6: Update SKILL.md for Image Modification Workflows

**Files:**
- Modify: `SKILL.md`

- [ ] **Step 1: Replace `SKILL.md` content**

```markdown
---
name: imagemagick-mcp
description: >
  Resize, crop, and reformat images to exact pixel dimensions via ImageMagick.
  Use when you need to produce a final asset at a specific size (social banners,
  profile pictures, thumbnails, OG images) and want smart background handling
  when the source aspect ratio doesn't match the target.
---

# ImageMagick MCP — Image Modification Skill

## When to Use This Skill

Use `crop_resize_blur_bg` whenever you need to:
- Resize an image to an exact pixel size (e.g. 1200×630 for Open Graph)
- Crop a wide image to a square thumbnail
- Produce social-media-ready assets without distortion

## Quick Reference

### Common Target Sizes

| Format | Width | Height | Recommended mode |
|--------|-------|--------|-----------------|
| Open Graph / Twitter card | 1200 | 630 | `blur` |
| Instagram square | 1080 | 1080 | `blur` |
| Instagram story / Reel | 1080 | 1920 | `blur` |
| YouTube thumbnail | 1280 | 720 | `cover` |
| LinkedIn banner | 1584 | 396 | `cover` |
| Avatar / profile picture | 400 | 400 | `cover` |
| App icon | 512 | 512 | `contain` |

### Mode Selection Guide

- **`blur`** (default): Fills letterbox/pillarbox areas with a blurred copy of the source. Best for editorial or brand images where black bars would look ugly.
- **`cover`**: Crops to fill — no padding, no bars. Best when the subject is centered and you can afford minor edge cropping (thumbnails, avatars).
- **`contain`**: Fits the whole image inside the target, pads remainder with black. Best for technical diagrams, logos with transparency, or when no cropping is acceptable.

## Tool Invocation

```json
{
  "tool": "crop_resize_blur_bg",
  "input_path": "/absolute/path/to/source.jpg",
  "output_path": "/absolute/path/to/output.jpg",
  "width": 1200,
  "height": 630,
  "mode": "blur",
  "blur": 30
}
```

- `blur` parameter is only used when `mode=blur`. Higher values = softer background.
- Output directory is created automatically.
- Input and output paths must be absolute.

## Workflow Patterns

### Batch resize to multiple social formats

Call `crop_resize_blur_bg` once per format. Example sequence:
1. OG image → 1200×630 `blur`
2. Twitter card → 1200×628 `blur`
3. Instagram square → 1080×1080 `blur`
4. Story → 1080×1920 `blur`

### Avatar from a portrait photo

Use `cover` so the face stays centered and fills the square without padding:
```json
{"mode": "cover", "width": 400, "height": 400}
```

### Logo on white/black background

Use `contain` so the logo is never clipped:
```json
{"mode": "contain", "width": 512, "height": 512}
```

## Setup (Go binary)

1. Install ImageMagick: `brew install imagemagick` (macOS) or `apt install imagemagick` (Linux).
2. Build the server: `make build` (requires Go 1.22+).
3. Register in your MCP client config:

```json
{
  "mcpServers": {
    "imagemagick": {
      "command": "/path/to/imagemagick-mcp/bin/imagemagick-mcp"
    }
  }
}
```
```

- [ ] **Step 2: Commit**

```bash
git add SKILL.md
git commit -m "docs: rewrite skill.md as ai-facing image modification guide"
```

---

## Task 7: Update README and CLAUDE.md

**Files:**
- Modify: `README.md`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Replace `README.md`**

```markdown
# ImageMagick MCP Server

Minimal MCP server that wraps ImageMagick for crop/resize workflows. Written in Go.

**Author:** [roxl.net](https://roxl.net)

## What It Does

- Resizes images to exact dimensions
- Three modes when aspect ratios differ:
  - `blur`: blurred background from the source image (default)
  - `cover`: crop to fill (no padding)
  - `contain`: fit inside with black padding
- Exposes a single MCP tool: `crop_resize_blur_bg`

## Requirements

- Go 1.22+ (to build)
- ImageMagick with `magick` in PATH (`brew install imagemagick`)

## Quick Start

```bash
# Build
make build

# Register the binary in your MCP client
# Binary: bin/imagemagick-mcp
```

## MCP Tool: `crop_resize_blur_bg`

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `input_path` | string | yes | — | Absolute path to source image |
| `output_path` | string | yes | — | Absolute path for output |
| `width` | number | yes | — | Target width in pixels |
| `height` | number | yes | — | Target height in pixels |
| `mode` | string | no | `blur` | `blur`, `cover`, or `contain` |
| `blur` | number | no | `30` | Blur radius (blur mode only) |

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build for current platform → `bin/imagemagick-mcp` |
| `make build-all` | Cross-compile for darwin/linux/windows (amd64 + arm64) |
| `make test` | Run all tests |
| `make clean` | Remove `bin/` |
| `make release` | Build all + create GitHub release (requires `gh`) |

## Release

```bash
git tag v1.2.0
make release
```
```

- [ ] **Step 2: Update `CLAUDE.md` commands section**

In `CLAUDE.md`, replace the Commands section:

```markdown
## Commands

```bash
# Build for current platform
make build

# Cross-compile for all platforms
make build-all

# Run tests (integration tests require magick in PATH)
make test

# Remove build artifacts
make clean

# Create GitHub release (requires: git tag vX.Y.Z first, then gh auth)
git tag vX.Y.Z
make release

# Run server directly (for debugging)
./bin/imagemagick-mcp
```

ImageMagick must be installed separately with `magick` available in `PATH`.
```

- [ ] **Step 3: Commit**

```bash
git add README.md CLAUDE.md
git commit -m "docs: update readme and claude.md for go binary and makefile"
```

---

## Self-Review

**Spec coverage check:**
- [x] Rewrite to Go — Tasks 1–4 cover module init, core logic, server, entry point
- [x] Makefile with build and release — Task 5 covers all targets including `make release` via `gh`
- [x] Skill for image modifications — Task 6 rewrites SKILL.md as AI-facing guide with mode selection, common sizes, and workflow patterns
- [x] Tests — Task 2 covers unit (`ratioClose`) and integration tests for all three modes
- [x] Version injection — `main.go` `var version = "dev"`, Makefile sets via ldflags
- [x] Cross-platform — `build-all` target covers darwin/linux/windows

**Placeholder scan:** No TBDs, all code blocks are complete, no "similar to Task N" references.

**Type consistency:** `cropResizeBlurBg` signature is consistent across `imagemagick.go` (definition), `imagemagick_test.go` (test calls), and `server.go` (handler call). `identifySize` likewise.
