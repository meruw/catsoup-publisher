"""Prepare a temporary Instagram transport copy; never modify the delivery."""
import hashlib
import json
import pathlib
import subprocess
import sys

import imageio_ffmpeg


def run(ffmpeg, *args):
    result = subprocess.run([ffmpeg, "-nostdin", "-hide_banner", "-v", "error", *args],
                            capture_output=True, timeout=100,
                            creationflags=getattr(subprocess, "CREATE_NO_WINDOW", 0))
    if result.returncode:
        raise RuntimeError("Instagram media preparation failed")
    return result.stdout


def prepare(source, destination, legacy=False):
    if destination.exists() or not source.is_file():
        raise RuntimeError("Preparation requires a new temporary destination")
    ffmpeg = imageio_ffmpeg.get_ffmpeg_exe()
    # Stream-copy the animation. Only the audio is encoded to the documented
    # AAC / 48 kHz / 128 kb/s target. Optional audio supports silent clips.
    # Meta requires front-loaded moov and no edit lists. Shift timestamps
    # together to retain A/V alignment without negative decode timestamps.
    # Legacy mode exists only for byte-identical recovery of saved uploads.
    run(ffmpeg, "-n", "-i", str(source), "-map", "0:v:0", "-map", "0:a:0?",
        "-c:v", "copy", "-c:a", "aac", "-b:a", "128k", "-ar", "48000",
        "-ac", "2", "-movflags", "+faststart",
        *([] if legacy else ["-avoid_negative_ts", "make_zero", "-use_editlist", "0"]),
        str(destination))
    def video_digest(path):
        return run(ffmpeg, "-i", str(path), "-map", "0:v:0", "-c:v", "copy",
                   "-f", "hash", "-hash", "sha256", "-").strip()
    if video_digest(source) != video_digest(destination):
        raise RuntimeError("Video changed during Instagram preparation")
    with destination.open("rb") as stream:
        digest = hashlib.file_digest(stream, "sha256").hexdigest()
    return {"sha256": digest, "size": destination.stat().st_size}


if __name__ == "__main__":
    try:
        print(json.dumps(prepare(pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2]), "--legacy" in sys.argv[3:])))
    except Exception:
        # Do not expose subprocess output, local paths or file metadata in errors.
        print("Instagram media preparation failed", file=sys.stderr)
        sys.exit(1)
