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
- ImageMagick (`magick` must be in PATH)

### Installing ImageMagick

**macOS**
```bash
brew install imagemagick
```

**Ubuntu / Debian**
```bash
sudo apt update && sudo apt install imagemagick
```

**Fedora / RHEL / CentOS**
```bash
sudo dnf install imagemagick
```

**Arch Linux**
```bash
sudo pacman -S imagemagick
```

**Windows**
Download the installer from [imagemagick.org/script/download.php](https://imagemagick.org/script/download.php) and check **"Add application directory to your system path"** during install.

Or with Chocolatey:
```powershell
choco install imagemagick
```

Or with Winget:
```powershell
winget install ImageMagick.ImageMagick
```

Verify installation:
```bash
magick --version
```

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
