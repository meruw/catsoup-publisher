"""Delivery contract v2 between production and Publisher: manifest.json tools.
Specification: contracts/delivery/CONTRACT.md

    python contracts\\delivery\\delivery.py complete  deliveries\\<id>\\manifest.json
    python contracts\\delivery\\delivery.py validate  deliveries\\<id>\\manifest.json [--require-publishable]
    python contracts\\delivery\\delivery.py review    deliveries\\<id>\\manifest.json --by NAME [--notes TEXT]
    python contracts\\delivery\\delivery.py approve   deliveries\\<id>\\manifest.json --by NAME

(python = the production .venv, which has imageio-ffmpeg.)

complete  measures every declared file (sha256; for videos also duration,
          resolution, fps and audio). If a file changed since it was reviewed
          or approved, that review/approval is void: it moves to `history` and
          the status drops back to "exported". Never raises a status.
validate  checks fields, paths, measurements and status rules. Prints two
          verdicts: VALID (may be ingested and previewed in any status) and
          PUBLISHABLE. Exit code 1 if not valid, or if --require-publishable
          and not publishable. Never modifies anything.
review    records a human review: status "reviewed", reviewer, date and the
          sha256 of every file reviewed. Only the creative owner runs it.
approve   records approval for publishing (requires a valid review of the
          same files and publishable: true). Only the creative owner runs it.
Paths in the manifest are relative to the manifest's folder.
"""
import argparse
import datetime
import hashlib
import json
import os
import re
import subprocess
import sys

CONTRACT, VERSION = "catsoup-delivery", 2
STATUSES = ("exported", "reviewed", "approved_for_publishing")
KINDS = ("final", "clean")
RANK = {s: i for i, s in enumerate(STATUSES)}


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as fh:
        for chunk in iter(lambda: fh.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def probe(path):
    import imageio_ffmpeg
    err = subprocess.run([imageio_ffmpeg.get_ffmpeg_exe(), "-hide_banner", "-i", path],
                         capture_output=True, text=True).stderr
    h, m, s = re.search(r"Duration: (\d+):(\d+):([\d.]+)", err).groups()
    v = re.search(r"Video: .*?, (\d+)x(\d+)[ ,].*?([\d.]+) fps", err)
    a = re.search(r"Audio: (\w+).*?(\d+) Hz, (\w+)", err)
    return dict(duration_s=round(int(h) * 3600 + int(m) * 60 + float(s), 2),
                resolution=f"{v.group(1)}x{v.group(2)}", fps=float(v.group(3)),
                audio=f"{a.group(1)} {a.group(2)} Hz {a.group(3)}" if a else None)


def load(path):
    return json.load(open(path, encoding="utf-8"))


def save(path, man):
    with open(path, "w", encoding="utf-8", newline="\n") as fh:
        json.dump(man, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


def items(man):
    return [("video", v) for v in man.get("videos", [])] + [("text", t) for t in man.get("texts", [])]


def current_hashes(man, base):
    return {it["id"]: sha256(os.path.join(base, it["path"])) for _, it in items(man)}


def today():
    return datetime.date.today().isoformat()


# ----------------------------------------------------------------- commands ---

def complete(path):
    base = os.path.dirname(os.path.abspath(path))
    man = load(path)
    for kind, it in items(man):
        f = os.path.join(base, it["path"])
        it["sha256"] = sha256(f)
        if kind == "video":
            it["measured"] = probe(f)
    now = current_hashes(man, base)
    for key in ("approval", "review"):
        rec = man.get(key) or {}
        files = rec.get("files") or {}
        changed = sorted(i for i in files if now.get(i) != files[i])
        if files and changed:
            man.setdefault("history", []).append(dict(date=today(), event=f"{key}_invalidated",
                                                      changed=changed, previous=rec))
            man[key] = {"by": None, "date": None, "files": {}} | ({"notes": ""} if key == "review" else {})
            print(f"{key} by {rec.get('by')} voided: changed {changed}")
    if not (man.get("review") or {}).get("files"):
        if man.get("status") != "exported":
            man["status"] = "exported"
        for v in man.get("videos", []):
            v["status"] = "exported"
    elif not (man.get("approval") or {}).get("files") and man.get("status") == "approved_for_publishing":
        man["status"] = "reviewed"
        for v in man.get("videos", []):
            if v.get("status") == "approved_for_publishing":
                v["status"] = "reviewed"
    save(path, man)
    print(f"measurements written to {path}")


def check(path):
    """Returns (errors, warnings, publishable)."""
    base = os.path.dirname(os.path.abspath(path))
    man = load(path)
    err, warn = [], []
    if man.get("contract") != CONTRACT or man.get("contract_version") != VERSION:
        err.append(f"contract/contract_version is not {CONTRACT} v{VERSION}")
    eid = man.get("episode_id", "")
    if not re.fullmatch(r"catsoup(-[a-z0-9]+)+", eid):
        err.append(f"invalid episode_id: {eid!r}")
    if os.path.basename(base) != eid:
        err.append(f"folder ({os.path.basename(base)}) does not match episode_id ({eid})")
    st = man.get("status")
    if st not in STATUSES:
        err.append(f"invalid episode status: {st!r}")
    if not man.get("publishable") and not man.get("not_publishable_reason"):
        err.append("publishable=false without not_publishable_reason")

    langs = set(man.get("languages", []))
    finals = {v.get("lang") for v in man.get("videos", []) if v.get("kind") == "final"}
    for lang in sorted(langs - finals):
        err.append(f"no final video for language {lang}")
    ids = [it.get("id") for _, it in items(man)]
    if len(ids) != len(set(ids)):
        err.append("duplicated ids")

    now = {}
    for kind, it in items(man):
        name, f = it.get("id"), os.path.join(base, it.get("path", ""))
        if not os.path.isfile(f):
            err.append(f"{name}: missing file {it.get('path')}")
            continue
        now[name] = sha256(f)
        if it.get("sha256") != now[name]:
            err.append(f"{name}: sha256 differs (file changed? run complete)")
        if kind != "video":
            continue
        if it.get("kind") not in KINDS:
            err.append(f"{name}: invalid kind {it.get('kind')!r}")
        if it.get("status") not in STATUSES:
            err.append(f"{name}: invalid status {it.get('status')!r}")
        m = probe(f)
        for k in ("duration_s", "resolution", "fps"):
            if (it.get("measured") or {}).get(k) != m[k]:
                err.append(f"{name}: {k} declared {(it.get('measured') or {}).get(k)} != measured {m[k]}")
        if not m["audio"]:
            warn.append(f"{name}: no audio track")

    rev, apr = man.get("review") or {}, man.get("approval") or {}

    def signed(rec, label):
        if not (rec.get("by") and rec.get("date")):
            err.append(f"status needs {label}.by and {label}.date")
        files = rec.get("files") or {}
        stale = sorted(i for i in now if files.get(i) != now[i])
        if stale:
            err.append(f"{label} does not cover the current files: {stale}")

    if st in ("reviewed", "approved_for_publishing"):
        signed(rev, "review")
    if st == "approved_for_publishing":
        signed(apr, "approval")
        if not man.get("publishable"):
            err.append("approved_for_publishing with publishable=false")
    for v in man.get("videos", []):
        vs = v.get("status")
        if vs in RANK and st in RANK:
            if RANK[vs] > RANK[st]:
                err.append(f"{v.get('id')}: status {vs} above the episode ({st})")
            if v.get("kind") == "final" and RANK[vs] < RANK[st]:
                err.append(f"{v.get('id')}: final video {vs} below the episode ({st})")
    publishable = (not err and st == "approved_for_publishing" and bool(man.get("publishable")))
    return err, warn, publishable, man


def validate(path, require_publishable=False):
    err, warn, publishable, man = check(path)
    for w in warn:
        print("WARNING:", w)
    for e in err:
        print("ERROR:", e)
    print(f"{path}: {'VALID' if not err else f'INVALID ({len(err)} errors)'}; "
          f"{'PUBLISHABLE' if publishable else 'NOT PUBLISHABLE'} "
          f"(status={man.get('status')}, publishable={man.get('publishable')})")
    if err or (require_publishable and not publishable):
        return 1
    return 0


def sign(path, key, by, notes=None):
    base = os.path.dirname(os.path.abspath(path))
    err, _, _, man = check(path)
    structural = [e for e in err if "needs review" not in e and "does not cover" not in e
                  and "below the episode" not in e and "above the episode" not in e]
    if structural:
        raise SystemExit("Fix the manifest first:\n  " + "\n  ".join(structural))
    files = current_hashes(man, base)
    if key == "approval":
        rev = man.get("review") or {}
        if (rev.get("files") or {}) != files or not rev.get("by"):
            raise SystemExit("Approval needs a review of exactly these files first")
        if not man.get("publishable"):
            raise SystemExit("publishable is false: set it to true (and clear the reason) before approving")
        new = "approved_for_publishing"
    else:
        new = "reviewed"
    rec = {"by": by, "date": today(), "files": files}
    if key == "review":
        rec["notes"] = notes or ""
    man[key] = rec
    man["status"] = new
    for v in man.get("videos", []):
        v["status"] = new
    save(path, man)
    print(f"{key} recorded by {by}; status = {new}")


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("command", choices=["complete", "validate", "review", "approve"])
    ap.add_argument("manifest")
    ap.add_argument("--require-publishable", action="store_true")
    ap.add_argument("--by")
    ap.add_argument("--notes")
    a = ap.parse_args()
    if a.command == "complete":
        complete(a.manifest)
    elif a.command == "validate":
        sys.exit(validate(a.manifest, a.require_publishable))
    else:
        if not a.by:
            raise SystemExit("--by NAME is required")
        sign(a.manifest, "review" if a.command == "review" else "approval", a.by, a.notes)
