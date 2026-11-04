#!/usr/bin/env python3
"""Snapshot the repository into internal/scaffold/testdata/template.

The snapshot is embedded into the binary (go:embed) so `newproject` works from a
built app without the source tree. Assets and the generator itself are excluded:
assets are optional/copyrighted, and including the generator would recurse.

go.mod / go.sum are stored with a .snapshot suffix: go:embed refuses a directory
that contains a go.mod (it treats it as another module). The generator renames
them back when materializing the template.

Usage:
  python scripts/snapshot-template.py          # write
  python scripts/snapshot-template.py --check  # verify, exit 1 when stale
"""

from __future__ import annotations

import hashlib
import shutil
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
DST = ROOT / "internal" / "scaffold" / "testdata" / "template"

# Directories (relative to ROOT) not part of the template.
SKIP_DIRS = {
    ".git",
    ".idea",
    ".workbuddy",
    "build",
    "assets",
    "internal/scaffold",
    "scripts/newproject",
}
# Repo-only files that do not belong in the template snapshot.
SKIP_FILES = {"scripts/snapshot-template.py", "internal/view/newproject_page.go", "internal/view/newproject_page_test.go"}

SNAP_SUFFIX = {"go.mod": ".snapshot", "go.sum": ".snapshot"}


def snap_name(rel: str) -> str:
    return rel + SNAP_SUFFIX.get(rel, "")


def rel_files(root: Path):
    out = []
    for p in root.rglob("*"):
        if not p.is_file():
            continue
        rel = p.relative_to(root).as_posix()
        if rel in SKIP_FILES:
            continue
        if any(rel == d or rel.startswith(d + "/") for d in SKIP_DIRS):
            continue
        out.append(rel)
    return sorted(out)


def digest(p: Path) -> str:
    h = hashlib.sha256()
    with p.open("rb") as f:
        for chunk in iter(lambda: f.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def write_snapshot() -> None:
    if DST.exists():
        shutil.rmtree(DST)
    for rel in rel_files(ROOT):
        dst = DST / snap_name(rel)
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(ROOT / rel, dst)
    print(f"snapshot refreshed: {len(rel_files(DST))} files -> {DST.relative_to(ROOT)}")


def check_snapshot() -> int:
    want = {snap_name(r): r for r in rel_files(ROOT)}
    have = rel_files(DST) if DST.exists() else []
    missing = [k for k in want if k not in have]
    extra = [r for r in have if r not in want]
    changed = [k for k, orig in want.items() if k in have and digest(ROOT / orig) != digest(DST / k)]
    if missing or extra or changed:
        print("snapshot is stale:", file=sys.stderr)
        for r in missing[:10]:
            print("  missing:", r, file=sys.stderr)
        for r in extra[:10]:
            print("  extra:", r, file=sys.stderr)
        for r in changed[:10]:
            print("  changed:", r, file=sys.stderr)
        print("run: python scripts/snapshot-template.py", file=sys.stderr)
        return 1
    print(f"snapshot up to date ({len(have)} files)")
    return 0


if __name__ == "__main__":
    if "--check" in sys.argv:
        raise SystemExit(check_snapshot())
    write_snapshot()
