"""Install the pinned Windows Go toolchain locally, verifying its official hash."""
import hashlib
import pathlib
import urllib.request
import zipfile

root = pathlib.Path(__file__).resolve().parents[3]
tools = root / ".tools"
tools.mkdir(exist_ok=True)
archive = tools / "go1.27.3.windows-amd64.zip"
expected = "f26081f438319757f8a1b46d7d0b2c5ec7583d4038e485e0aa725cfde7885fff"
if not archive.exists():
    urllib.request.urlretrieve("https://go.dev/dl/" + archive.name, archive)
if hashlib.file_digest(archive.open("rb"), "sha256").hexdigest() != expected:
    raise SystemExit("Go archive checksum mismatch; remove the invalid archive and retry.")
if not (tools / "go/bin/go.exe").exists():
    with zipfile.ZipFile(archive) as package:
        for member in package.infolist():
            if not (tools / member.filename).resolve().is_relative_to(tools.resolve()):
                raise SystemExit("Unsafe archive entry")
        package.extractall(tools)
print("Local Go 1.27.3 is available.")
