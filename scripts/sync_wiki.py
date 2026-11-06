"""Copy the public guides to a local wiki checkout, or check for drift."""
import argparse
from pathlib import Path

GUIDES = Path(__file__).resolve().parents[1] / "docs" / "guides"
WIKI_NAVIGATION = {"home", "_sidebar", "_footer"}


def guide_body(source):
    lines = source.read_text(encoding="utf-8").splitlines(keepends=True)
    if not lines or lines[0].strip() != "---":
        raise ValueError(f"{source.name} is missing its Jekyll header")
    for index, line in enumerate(lines[1:], start=1):
        if line.strip() == "---":
            return "".join(lines[index + 1:]).strip() + "\n"
    raise ValueError(f"{source.name} has an incomplete Jekyll header")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("wiki", type=Path, help="Path to the cloned GitHub wiki")
    parser.add_argument("--check", action="store_true", help="Report differences without writing")
    args = parser.parse_args()
    if not (args.wiki / ".git").exists() or not (args.wiki / "Home.md").is_file():
        parser.error("Choose the wiki checkout containing Home.md")

    changed = []
    for source in sorted(GUIDES.glob("*.md")):
        if source.stem.casefold() in WIKI_NAVIGATION:
            parser.error(f"{source.name} is reserved for wiki navigation")
        body = guide_body(source)
        target = args.wiki / source.name
        if target.is_file() and target.read_text(encoding="utf-8") == body:
            continue
        changed.append(source.name)
        if not args.check:
            target.write_bytes(body.encode("utf-8"))

    if changed:
        print(("Out of sync: " if args.check else "Updated: ") + ", ".join(changed))
    else:
        print("Wiki guides are in sync.")
    return int(args.check and bool(changed))


if __name__ == "__main__":
    raise SystemExit(main())
