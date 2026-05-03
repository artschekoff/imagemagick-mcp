# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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

## Architecture

Single Go package at the repo root exposes two MCP tools via `github.com/mark3labs/mcp-go` and shells out to the `magick` CLI. Runs with stdio transport.

**Files:**
- `main.go` — entry point, injects `version` at build time
- `server.go` — MCP server init, tool registration (`describe_imagemagick_interface`, `crop_resize_blur_bg`), and request handlers
- `imagemagick.go` — shells out to `magick`: `identifySize`, `ratioClose`, `cropResizeBlurBg`, `runMagick`
- `imagemagick_test.go` — unit tests for `ratioClose`; integration tests for all three resize modes (skip if `magick` absent)

**Resize logic in `cropResizeBlurBg`:**
1. Validates input exists, creates output dir, validates mode and dimensions
2. Reads source size via `magick identify`
3. If aspect ratios match (within 1%) or `mode=cover`: resize+crop with `^` and extent
4. `mode=contain`: resize to fit + black padding
5. `mode=blur` (default, aspect mismatch): composites source over a blurred cover-cropped background
