# GitHub 发版与自动更新步骤

仓库：<https://github.com/lark203/bestuav-hanabi>

本文记录从代码推送到 GitHub、触发 Actions、发布 release、验证客户端自动更新的完整步骤。

## 0. 当前状态

- 代码已推送到 GitHub 的 `main` 和 `dev`，默认分支是 `main`。
- `.github/workflows/release.yml` 已启用，推 `v*` tag 时触发。
- Actions 的 Workflow permissions 已设为 **Read and write**。
- `mygo.json` 已配置：

  ```json
  "updates": {
    "publicKey": "n4yh...（已固化）",
    "github": "lark203/bestuav-hanabi",
    "tagPrefix": "v"
  }
  ```

- **v0.1.0 已发布**，Actions 的三平台构建全部成功；资产包含 Windows 安装包、
  macOS dmg、Linux deb/tar.gz、各平台 `update-*.json`。
- 代码已 bump 到 **0.1.1**（全屏修复版本），tag `v0.1.1` 尚未推送。
- `MYGO_UPDATER_PRIVATE_KEY` 已配置。换仓库/重建时要重新配置。
- **注意**：v0.1.0 的构建时间早于「全屏开场不居中」修复；要拿到修复必须发 0.1.1。

## 1. 配置更新签名私钥（已配置）

如果换仓库或重新配置，打开：

`Settings → Secrets and variables → Actions → New repository secret`

- Name：`MYGO_UPDATER_PRIVATE_KEY`
- Value：`D:\workbuddy\bestuav-hanabi-update-keys\mygo-update.key` 的完整内容
  （一行，去掉末尾换行）

这是 `go tool mygo keygen` 生成的私钥，用于给更新包签名；**不要提交到仓库**。
没有它时 workflow 的 `Check update signing key` 会直接失败。

`GITHUB_TOKEN` 不需要手动配置，workflow 使用 `github.token`。

## 2. 推送 tag 触发构建

`mygo.json` 的版本必须和 tag 一致。当前已经是 `0.1.1`：

```powershell
cd D:\workbuddy\bestuav-hanabi
git checkout main
git tag v0.1.1
git push github v0.1.1
```

只有推 GitHub 的 tag 会触发 CI，推 AtomGit 不会。

如果 tag 已存在：

```powershell
git tag -d v0.1.1
git push github :refs/tags/v0.1.1
git tag v0.1.1
git push github v0.1.1
```

## 3. 等 Actions 完成

<https://github.com/lark203/bestuav-hanabi/actions>

`release` workflow 分两段：

1. **prepare**：校验 `v<version>` 与 `mygo.json` 一致，创建 draft release。
2. **build**：Windows / macOS / Linux 三个 runner 串行构建（约 8 分钟），
   各自上传安装包、更新压缩包、`update-<target>.json` 和 `SHA256SUMS`。

Windows 安装包在 Windows runner 上生成；macOS 是 `darwin/universal`；Linux 是
`linux/amd64`。

## 4. 发布 draft release

<https://github.com/lark203/bestuav-hanabi/releases>

打开 `v0.1.1` draft，检查资产，点 **Publish release**。建议勾选
**Set as latest release**。draft 对客户端不可见，必须发布后自动更新才能查到。

## 5. 验证自动更新

1. 用 CI 产出的 `hanabi-setup-0.1.0-amd64.exe` 安装 0.1.0。
   **不要用 owner 改成 `lark203` 之前的本地 `build/` 包**，那个包链的是旧地址。
2. 确认 `mygo.json` 的 `version` 是 `0.1.1`（已 bump），提交。
3. 打 tag 并推送：
   ```powershell
   git tag v0.1.1
   git push github v0.1.1
   ```
4. 等 `v0.1.1` 的 draft 构建完，发布。
5. 打开已安装的 0.1.0 → 设置 → 版本与更新 → 检查更新，应能看到 0.1.1，
   下载安装后重启。

## 6. 本机推送与代理

本机 WorkBuddy 会注入 `HTTP_PROXY=127.0.0.1:5437`，它会拦 `github.com`。
按 bestuav-mygo 的记录，给 github.com 单独走 Clash 的 7890：

```powershell
git config --global http.https://github.com.proxy http://127.0.0.1:7890
```

回滚：

```powershell
git config --global --unset http.https://github.com.proxy
```

凭证沿用本机 Git Credential Manager 里 bestuav-mygo 同一套 GitHub 账号即可。

## 7. 可选签名

当前**没有配置**代码签名，所以：

- Windows：SmartScreen 会提示未知发布者。
- macOS：ad hoc 签名，别的 Mac 上 Gatekeeper 可能拦截。

需要时再加：

- Windows：`WINDOWS_CERTIFICATE`（base64 pfx）+ `WINDOWS_CERTIFICATE_PASSWORD`，
  并在 workflow 里加 osslsigncode 步骤。
- macOS：`MACOS_CERTIFICATE`、密码、`APPLE_ID`、`APPLE_TEAM_ID`、
  `APPLE_APP_PASSWORD`。

## 8. 常见问题

| 现象 | 原因 / 处理 |
|---|---|
| `Check update signing key` 失败 | 没加 `MYGO_UPDATER_PRIVATE_KEY` secret |
| `gh release create` 403 | Actions Workflow permissions 不是 Read and write |
| tag 推了但没触发 | tag 推到了 AtomGit；必须推 GitHub |
| 版本号不一致 | `mygo.json` 的 `version` 必须是 tag 去掉 `v` |
| 客户端检查不到更新 | release 还是 draft，或没有发布为 latest |
| `git push` 502 / connection reset | 配 `http.https://github.com.proxy` 走 7890 |
| 安装后更新包校验失败 | 私钥被替换，和 `updates.publicKey` 不匹配 |
