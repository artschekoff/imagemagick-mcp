---
name: imagemagick-mcp
description: Crop/resize images with ImageMagick via MCP. Use when you need to resize to a target width/height and, if aspect ratio differs, pad with a blurred version of the same image. Provides a tool to generate final assets for social formats.
---

# ImageMagick MCP

## Overview
Provides a simple MCP tool to crop/resize images to exact dimensions. If aspect ratio differs, the background is filled with a blurred version of the source image.

## Quick Start
1. Use the project root virtualenv that has `mcp` installed (e.g., `.venv`).
2. Start server: `.venv/bin/python skills/imagemagick-mcp/scripts/server.py`.
3. Call `crop_resize_blur_bg` with `input_path`, `output_path`, `width`, `height`, and `mode`.

## Tool Usage
- `crop_resize_blur_bg`
  - Required: `input_path`, `output_path`, `width`, `height`
  - Optional: `mode` (`blur|cover|contain`, default `blur`)
  - Optional: `blur` (default `30`, only used in `blur` mode)

## Notes
- If the source aspect ratio matches the target, it will be resized directly.
- `blur`: blurred background from the same image, resized and composited.
- `cover`: fills target by cropping (no padding).
- `contain`: fits inside target with black padding.
