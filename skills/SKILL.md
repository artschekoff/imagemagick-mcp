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
