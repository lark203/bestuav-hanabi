#!/usr/bin/env python3
"""发布前检查：版本、CHANGELOG、tag、模板快照、vet/build/test。

用法（仓库根目录）：
  python scripts/release-check.py                 # 检查当前版本，不校验 tag 是否存在
  python scripts/release-check.py --tag v0.1.4    # 发版前：tag 必须等于 v + mygo.json 版本
  python scripts/release-check.py --skip-tests    # 快速检查（跳过 go test）
  python scripts/release-check.py --require-clean # 工作区必须干净

退出码：0 全通过；1 有失败项。
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
VERSION_RE = re.compile(r"^(\d+)\.(\d+)\.(\d+)$")


class Check:
    def __init__(self) -> None:
        self.failed = 0

    def ok(self, msg: str) -> None:
        print(f"  [OK] {msg}")

    def fail(self, msg: str) -> None:
        self.failed += 1
        print(f"  [FAIL] {msg}")

    def warn(self, msg: str) -> None:
        print(f"  [WARN] {msg}")


def run(cmd: list[str], check: Check, label: str) -> bool:
    print(f"\n== {label} ==")
    print("$ " + " ".join(cmd))
    proc = subprocess.run(cmd, cwd=ROOT)
    if proc.returncode != 0:
        check.fail(f"{label} 失败（exit {proc.returncode}）")
        return False
    check.ok(label)
    return True


def main() -> int:
    ap = argparse.ArgumentParser(description="发布前检查")
    ap.add_argument("--tag", default="", help="计划发布的 tag，如 v0.1.4")
    ap.add_argument("--skip-tests", action="store_true", help="跳过 go test")
    ap.add_argument("--require-clean", action="store_true", help="工作区必须干净")
    args = ap.parse_args()

    check = Check()
    print("== 版本与 CHANGELOG ==")

    try:
        cfg = json.loads((ROOT / "mygo.json").read_text(encoding="utf-8"))
        version = str(cfg.get("version", ""))
    except Exception as exc:  # noqa: BLE001
        check.fail(f"读 mygo.json 失败: {exc}")
        return 1

    if not VERSION_RE.match(version):
        check.fail(f"mygo.json 版本格式不对: {version!r}")
    else:
        check.ok(f"mygo.json 版本 {version}")

    changelog = (ROOT / "CHANGELOG.md").read_text(encoding="utf-8")
    if re.search(rf"^##\s+{re.escape(version)}\b", changelog, re.M):
        check.ok(f"CHANGELOG 有 {version} 段")
    else:
        check.fail(f"CHANGELOG 缺少 `## {version}` 段（构建会把该段写进更新说明）")

    if args.tag:
        want = "v" + version
        if args.tag != want:
            check.fail(f"tag {args.tag} 与版本不匹配（应为 {want}）")
        else:
            check.ok(f"tag {args.tag} 与版本一致")
        exists = subprocess.run(
            ["git", "rev-parse", "-q", "--verify", f"refs/tags/{args.tag}"],
            cwd=ROOT, capture_output=True, text=True,
        )
        if exists.returncode == 0:
            check.fail(f"tag {args.tag} 已存在，不能重复发布")
        else:
            check.ok(f"tag {args.tag} 尚未创建")

    print("\n== 工作区 ==")
    dirty = subprocess.run(
        ["git", "status", "--porcelain"], cwd=ROOT, capture_output=True, text=True
    ).stdout.strip()
    if dirty:
        if args.require_clean:
            check.fail("工作区不干净（有未提交改动）")
        else:
            check.warn("工作区有未提交改动（确认是否属于本次发布）")
    else:
        check.ok("工作区干净")

    run([sys.executable, "scripts/snapshot-template.py", "--check"], check, "模板快照")
    run(["go", "vet", "./..."], check, "go vet")
    run(["go", "build", "./..."], check, "go build")
    if args.skip_tests:
        check.warn("已跳过 go test（--skip-tests）")
    else:
        run(["go", "test", "./..."], check, "go test")

    print("\n" + ("=" * 48))
    if check.failed:
        print(f"发布检查失败：{check.failed} 项")
        return 1
    print("发布检查全部通过。下一步：提交 → 打 annotated tag → 推 main 与 tag。")
    if args.tag:
        print(f"  git tag -a {args.tag} -m \"{args.tag}\"")
        print(f"  git push github {args.tag} && git push origin {args.tag}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
