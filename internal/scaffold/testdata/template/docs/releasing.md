# 发布流程（bestuav-hanabi）

发布由 **tag** 触发 `.github/workflows/release.yml`：校验 tag 与 `mygo.json` 的
`version` 完全一致（`v0.1.3` ↔ `0.1.3`），创建 draft release，三平台跑
`go tool mygo build -platform ... -upload` 上传安装包与 `update-*.json`。
**最后一步发布 draft 是手动的**（工作流只把它留成 draft）。

## 1. 准备

1. 版本号：改 `mygo.json` 的 `version`（如 `0.1.2` → `0.1.3`）。
2. `CHANGELOG.md` 顶部加 `## 0.1.3 - YYYY-MM-DD` 段（构建会把它写进更新说明；
   找不到对应段落时 `mygo build` 会失败）。
3. 如果改了模板源码：`python scripts/snapshot-template.py` 并提交。
4. 本地检查：

```powershell
go vet ./...
go test ./...
python scripts/snapshot-template.py --check
```

## 2. 提交与打 tag

```powershell
git add -A
git commit -m "release: v0.1.3"
git tag -a v0.1.3 -m "v0.1.3"
git push github main
git push origin main
git push github v0.1.3
git push origin v0.1.3
```

tag 必须是 annotated tag；tag 与 `mygo.json` 不一致时 `prepare` job 会直接失败。

## 3. 等三平台构建

- Windows / macOS / Linux 串行构建（`max-parallel: 1`），各自上传产物到 draft release。
- 需要仓库 secret `MYGO_UPDATER_PRIVATE_KEY`（缺失时构建直接报错——没有它客户端
  无法校验更新包）。
- 构建日志的 Summary 会列出各平台产物大小。

## 4. 发布

确认 draft 里三平台产物与 `update-*.json` 齐全后：

```bash
gh release edit v0.1.3 --draft=false --latest
```

发布后旧客户端在「设置 → 版本与更新」里检查到新版本，按 mygo 的签名校验流程更新。

## 回滚

- 安装包回滚：把 `--latest` 指回上一个 release，或撤销 draft。
- 代码回滚：`git revert` 后重新发一个 patch 版本；**不要移动已发布的 tag**
  （更新源的版本比较会乱）。
