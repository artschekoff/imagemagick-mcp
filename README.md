# ImageMagick MCP Server

Minimal MCP server that wraps ImageMagick for crop/resize workflows. It can resize to an exact target size and, when the aspect ratio differs, it can pad with a blurred background, crop-to-cover, or contain with black padding.

## What It Does
- Resizes images to exact dimensions
- Supports three modes when aspect ratios differ:
  - `blur`: blurred background from the source image
  - `cover`: crop to fill (no padding)
  - `contain`: fit inside with black padding
- Exposes a single MCP tool: `crop_resize_blur_bg`

## Quick Start
1. Install dependencies:

```bash
python3 -m pip install -r skills/imagemagick-mcp/scripts/requirements.txt
```

2. Ensure ImageMagick is installed and `magick` is in your `PATH`.

3. Start the server:

```bash
python3 skills/imagemagick-mcp/scripts/server.py
```

## MCP Tool
- `crop_resize_blur_bg`
  - Required: `input_path`, `output_path`, `width`, `height`
  - Optional: `mode` (`blur|cover|contain`, default `blur`)
  - Optional: `blur` (default `30`, only used in `blur` mode)

## Notes
- If the source aspect ratio matches the target, the image is resized directly.
- Output directories are created automatically.
- The server runs with MCP stdio transport.

## Files
- Skill definition: `skills/imagemagick-mcp/SKILL.md`
- Server implementation: `skills/imagemagick-mcp/scripts/server.py`
