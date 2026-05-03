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

This is a single-file MCP server (`scripts/server.py`) built with `FastMCP`. It exposes two tools:

- `describe_imagemagick_interface` — returns server metadata
- `crop_resize_blur_bg` — the main tool: resizes an image to exact `width`×`height`

The server shells out to `magick` (ImageMagick CLI) for all image operations. No image data passes through Python — only paths and subprocess calls.

### Resize logic in `crop_resize_blur_bg`

1. Reads source dimensions via `magick identify`.
2. If aspect ratios match (within 1% tolerance) **or** `mode=cover`: resize+crop with `-resize WxH^ -gravity center -extent WxH`.
3. `mode=contain`: resize to fit with `-resize WxH`, then extend canvas with black padding.
4. `mode=blur` (default, aspect mismatch): composites the source (fit inside) over a blurred, cover-cropped version of itself as background — requires a three-layer `magick` command using parenthesized clone operations.

The server runs with `mcp.run(transport="stdio")` — it is consumed by MCP clients that spawn it as a subprocess.
