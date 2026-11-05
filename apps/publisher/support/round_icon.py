"""Regenerate the committed rounded icon using the existing development Pillow."""

from pathlib import Path

from PIL import Image, ImageChops, ImageDraw


brand = Path(__file__).resolve().parents[1] / "frontend/public/brand"
source = Image.open(brand / "catsoup.png").convert("RGBA")
width, height = source.size
scale = 8
radius = round(min(width, height) * 0.16)
mask = Image.new("L", (width * scale, height * scale))
ImageDraw.Draw(mask).rounded_rectangle(
    (0, 0, width * scale - 1, height * scale - 1),
    radius=radius * scale,
    fill=255,
)
mask = mask.resize(source.size, Image.Resampling.LANCZOS)
result = source.copy()
result.putalpha(ImageChops.multiply(source.getchannel("A"), mask))
result.save(brand / "catsoup-rounded.png", optimize=True)
print(f"Saved {width}x{height} icon with {radius}px corners; RGB artwork unchanged.")
