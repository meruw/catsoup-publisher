"""Offline media check using the pinned local FFmpeg runtime."""
import hashlib
import pathlib
import subprocess
import tempfile
import struct

import imageio_ffmpeg
from prepare_instagram import prepare

def atoms(path):
    found = []
    def visit(stream, end):
        while stream.tell() < end:
            start = stream.tell()
            size, kind = struct.unpack(">I4s", stream.read(8))
            if size == 1:
                size = struct.unpack(">Q", stream.read(8))[0]
            if size == 0:
                size = end - start
            assert size >= 8 and start + size <= end
            found.append(kind)
            if kind in (b"moov", b"trak", b"edts"):
                visit(stream, start + size)
            stream.seek(start + size)
    with path.open("rb") as stream:
        visit(stream, path.stat().st_size)
    return found

ffmpeg = imageio_ffmpeg.get_ffmpeg_exe()
with tempfile.TemporaryDirectory(prefix="catsoup-media-check-") as folder:
    root = pathlib.Path(folder)
    for audio in (True, False):
        source = root / f"source-{audio}.mp4"
        target = root / f"prepared-{audio}.mp4"
        command = [ffmpeg, "-nostdin", "-v", "error", "-f", "lavfi", "-i", "color=c=red:s=180x320:r=30"]
        if audio:
            command += ["-f", "lavfi", "-i", "sine=frequency=440:sample_rate=48000"]
        command += ["-t", "2", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", "-b:a", "256k", "-ac", "2", str(source)]
        subprocess.run(command, check=True, capture_output=True)
        before = hashlib.sha256(source.read_bytes()).hexdigest()
        result = prepare(source, target)
        assert result["sha256"] == hashlib.sha256(target.read_bytes()).hexdigest()
        assert result["size"] == target.stat().st_size
        layout = atoms(target)
        assert b"edts" not in layout and b"elst" not in layout
        assert layout.index(b"moov") < layout.index(b"mdat")
        legacy = root / f"legacy-{audio}.mp4"
        prepare(source, legacy, legacy=True)
        assert b"edts" in atoms(legacy)
        assert before == hashlib.sha256(source.read_bytes()).hexdigest()
        subprocess.run([ffmpeg, "-v", "error", "-i", str(target), "-f", "null", "-"], check=True, capture_output=True)
        # Existing output must never be overwritten, including an original.
        try:
            prepare(source, source)
            raise AssertionError("source overwrite allowed")
        except RuntimeError:
            assert before == hashlib.sha256(source.read_bytes()).hexdigest()
print("PASS: audio and silent media, unchanged video stream/source, output hash, MP4 atom layout, legacy preparation, decoding and overwrite protection")
