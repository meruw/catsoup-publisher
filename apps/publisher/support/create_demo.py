"""Generate an ignored, non-publishable sample without production assets."""
import importlib.util
import json
from pathlib import Path
import shutil
import subprocess

import imageio_ffmpeg


root = Path(__file__).resolve().parents[3]
folder = root / "deliveries/catsoup-demo"
manifest_path = folder / "manifest.json"
spec = importlib.util.spec_from_file_location("delivery_contract", root / "contracts/delivery/delivery.py")
contract = importlib.util.module_from_spec(spec)
spec.loader.exec_module(contract)

if folder.exists():
    if manifest_path.is_file():
        print("Demo already exists; validating it without overwriting any files.")
        raise SystemExit(contract.validate(str(manifest_path)))
    raise SystemExit("Demo folder already exists without a manifest; inspect it before retrying.")

(folder / "video").mkdir(parents=True)
sample = folder / "video/es_final.mp4"
subprocess.run([
    imageio_ffmpeg.get_ffmpeg_exe(), "-hide_banner", "-loglevel", "error", "-nostdin",
    "-f", "lavfi", "-i", "testsrc2=size=540x960:rate=30",
    "-f", "lavfi", "-i", "sine=frequency=440:sample_rate=48000",
    "-t", "12", "-c:v", "libx264", "-preset", "ultrafast", "-crf", "30",
    "-pix_fmt", "yuv420p", "-af", "volume=0.03", "-c:a", "aac",
    "-movflags", "+faststart", str(sample),
], check=True)

videos = []
for lang in ("es", "en"):
    for kind in ("final", "clean"):
        name = f"{lang}_{kind}"
        target = folder / f"video/{name}.mp4"
        if target != sample:
            shutil.copyfile(sample, target)
        videos.append(dict(id=name, lang=lang, kind=kind, narration=False,
                           burned_text=False, status="exported", path=f"video/{name}.mp4"))
manifest = dict(
    contract="catsoup-delivery", contract_version=2, episode_id="catsoup-demo",
    internal_title="Demo: colors and sound", version="v1", languages=["es", "en"],
    exported_on=contract.today(), status="exported", publishable=False,
    not_publishable_reason="Synthetic test sample. Not intended for publishing.",
    review=dict(by=None, date=None, notes="", files={}),
    approval=dict(by=None, date=None, files={}), videos=videos, texts=[],
    source=dict(notes="Generated locally from FFmpeg test sources; no real episode."), history=[],
)
manifest_path.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
contract.complete(str(manifest_path))
raise SystemExit(contract.validate(str(manifest_path)))
