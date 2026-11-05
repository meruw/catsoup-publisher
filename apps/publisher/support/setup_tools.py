"""Install CATsoup's pinned Windows x64 build tools under .tools."""
import hashlib
import pathlib
import urllib.request
import zipfile

root = pathlib.Path(__file__).resolve().parents[3]
tools = root / ".tools"
tools.mkdir(exist_ok=True)


def install(url, checksum, executable):
    archive = tools / url.rsplit("/", 1)[1]
    if not archive.exists():
        temporary = archive.with_suffix(".download")
        try:
            urllib.request.urlretrieve(url, temporary)
            temporary.replace(archive)
        finally:
            temporary.unlink(missing_ok=True)
    digest = hashlib.sha256()
    with archive.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    if digest.hexdigest() != checksum:
        raise SystemExit(f"Checksum mismatch for {archive.name}; remove the archive and retry.")
    if not (tools / executable).is_file():
        with zipfile.ZipFile(archive) as package:
            for member in package.infolist():
                if not (tools / member.filename).resolve().is_relative_to(tools.resolve()):
                    raise SystemExit("Unsafe archive entry")
            package.extractall(tools)


install("https://go.dev/dl/go1.27.3.windows-amd64.zip",
        "f26081f438319757f8a1b46d7d0b2c5ec7583d4038e485e0aa725cfde7885fff", "go/bin/go.exe")
install("https://nodejs.org/dist/v24.22.2/node-v24.22.2-win-x64.zip",
        "ca1bff5d8ead13c9f2ea404a176db46bd430704d5b5113734049fa365e7799e2",
        "node-v24.22.2-win-x64/npm.cmd")
print("Local Go 1.27.3 and Node.js 24.22.2 are available.")
