#!/usr/bin/env python3
"""ImageMagick MCP server for crop/resize with blurred background padding."""

from __future__ import annotations

import subprocess
from pathlib import Path
from typing import Dict, Optional

try:
    from mcp.server.fastmcp import FastMCP
except ImportError as exc:
    raise SystemExit("Missing dependency 'mcp'. Install with: pip install mcp") from exc

SERVER_NAME = "imagemagick-mcp"

mcp = FastMCP(SERVER_NAME)


def _run(cmd: list[str]) -> None:
    subprocess.run(cmd, check=True)


def _magick_cmd() -> list[str]:
    return ["magick"]


def _identify_size(path: Path) -> tuple[int, int]:
    cmd = _magick_cmd() + ["identify", "-format", "%w %h", path.as_posix()]
    out = subprocess.check_output(cmd, text=True).strip()
    w_str, h_str = out.split()
    return int(w_str), int(h_str)


def _ratio_close(a: float, b: float, tol: float = 0.01) -> bool:
    return abs(a - b) <= tol


@mcp.tool()
def describe_imagemagick_interface() -> Dict[str, str]:
    return {
        "server": SERVER_NAME,
        "tools": "crop_resize_blur_bg",
        "note": "Uses ImageMagick. If aspect ratio differs, pads with blurred version of the image.",
    }


@mcp.tool()
def crop_resize_blur_bg(
    input_path: str,
    output_path: str,
    width: int,
    height: int,
    mode: str = "blur",
    blur: int = 30,
) -> Dict[str, Optional[str]]:
    """Resize/crop to target size. If aspect ratio differs, use mode: blur|cover|contain."""
    src = Path(input_path)
    if not src.exists():
        raise FileNotFoundError(f"Input not found: {src}")

    dst = Path(output_path)
    dst.parent.mkdir(parents=True, exist_ok=True)

    src_w, src_h = _identify_size(src)
    src_ratio = src_w / src_h
    tgt_ratio = width / height

    mode = mode.strip().lower()
    if mode not in {"blur", "cover", "contain"}:
        raise ValueError("mode must be one of: blur, cover, contain")

    if _ratio_close(src_ratio, tgt_ratio) or mode == "cover":
        cmd = _magick_cmd() + [
            src.as_posix(),
            "-resize",
            f"{width}x{height}^",
            "-gravity",
            "center",
            "-extent",
            f"{width}x{height}",
            dst.as_posix(),
        ]
        _run(cmd)
        return {"ok": "true", "output_path": dst.as_posix(), "mode": mode}

    if mode == "contain":
        cmd = _magick_cmd() + [
            src.as_posix(),
            "-resize",
            f"{width}x{height}",
            "-gravity",
            "center",
            "-background",
            "black",
            "-extent",
            f"{width}x{height}",
            dst.as_posix(),
        ]
        _run(cmd)
        return {"ok": "true", "output_path": dst.as_posix(), "mode": mode}

    # Build blurred background and composite centered foreground
    cmd = _magick_cmd() + [
        src.as_posix(),
        "(",
        "+clone",
        "-resize",
        f"{width}x{height}^",
        "-gravity",
        "center",
        "-extent",
        f"{width}x{height}",
        "-blur",
        f"0x{blur}",
        ")",
        "(",
        "+clone",
        "-resize",
        f"{width}x{height}",
        ")",
        "-delete",
        "0",
        "-gravity",
        "center",
        "-compose",
        "over",
        "-composite",
        dst.as_posix(),
    ]
    _run(cmd)
    return {"ok": "true", "output_path": dst.as_posix(), "mode": mode}


def main() -> None:
    mcp.run(transport="stdio")


if __name__ == "__main__":
    main()
