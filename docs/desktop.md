# 桌面客户端 / Desktop Client

## 安装 / Installation

版本 / Version: `1.0.20260920`。标识符 / Identifier: `org.mutantcat.chickreomte`。

| 平台 / Platform | 文件 / File | 要求 / Requirements |
| --- | --- | --- |
| Windows x64 | NSIS `.exe` | Windows 10/11; offline WebView2 installer included |
| macOS Intel | `macos-x64.dmg` | macOS 14+ |
| macOS Apple Silicon | `macos-arm64.dmg` | macOS 14+ |
| Linux x64 | `.AppImage` | Ubuntu 22.04-compatible glibc, GTK/WebKit environment, FUSE 2; X11 for capture |

Windows 双击安装程序，macOS 打开 DMG 并将应用拖入 Applications。
Linux 首次使用需 `chmod +x ChickReomte_*.AppImage`，再启动文件；无 FUSE 时可尝试 `--appimage-extract-and-run`。

Double-click the Windows installer. On macOS, open the DMG and drag the app to Applications.
On Linux, mark the AppImage executable first; `--appimage-extract-and-run` is an alternative when FUSE is unavailable.

macOS 的应用和 DMG 均为 ad-hoc 签名，不包含开发者证书和 Apple 公证，不能保证免 Gatekeeper 提示。
Windows 未使用商业签名证书，可能显示 SmartScreen。不要全局关闭系统安全保护。

Both the macOS app and DMG are ad-hoc signed, without a Developer ID or notarization. Gatekeeper may require explicit approval.
Windows is not Authenticode-signed and may show SmartScreen. Do not disable system-wide security protections.

## 连接 / Connection

先按[部署指南](startup.md)准备中继。各设备使用不同 ID、同一中继和强随机密钥。
TLS 默认开启并验证证书。目标 ID 留空表示不生成控制规则；填写目标 ID 会创建 VNC 规则。
保存并启动后，点击会话面板进入浏览器。关闭窗口退出应用会停止其客户端。请在退出前关闭远程 Shell 中自行启动的后台任务。

Deploy a relay using the [startup guide](startup.md). Use unique device IDs and the same relay and strong shared secret.
TLS is enabled with certificate verification. An empty target creates no control rule; a target creates a VNC rule.
Save, start, and open the browser dashboard. Exiting the desktop app stops its client. Stop any detached background tasks you started in remote shells before exiting.

配置 / Settings: Tauri `app_data_dir()/settings.json`。密钥以本机配置文件形式保存，不是系统钥匙串；Unix 目录权限 0700、文件 0600。
日志 / Logs: `engine.log` and `logs/` under the same directory.
请勿分享包含密钥的配置文件。Do not share settings containing credentials.

桌面端 HTTP 入口固定监听回环地址。远程会话在系统浏览器中打开，不能调用 Tauri 原生接口。
Local HTTP listeners bind to loopback. Remote session content opens in the system browser, without native Tauri IPC privileges.
中继共享密钥等同主机访问授权。The relay shared secret authorizes host access; share only with trusted peers.

## 开发 / Development

Requirements: Node.js 22, Rust stable, Go 1.25, platform C compiler; Linux also needs WebKitGTK 4.1 development libraries.

```sh
cd desktop
npm ci
npm run prepare:sidecar
npm test
npm run build
cd src-tauri
cargo test --locked --lib
cd ..
npm run desktop:dev
npm run desktop:build -- --bundles dmg
```

Use `nsis` on Windows and `appimage` on Linux. Windows uses MinGW for Go CGO and MSVC for Rust.
Linux packaging requires `/usr/bin/xclip`; the AppImage bundles it for clipboard support.

## 发布 / Release

同步更新 `build`、两个 Go main 文件、`desktop/package.json` / lockfile、Cargo manifest / lockfile 和 `tauri.conf.json`。
Update the version consistently in those files, commit, and push a matching version tag:

```sh
git tag v1.0.20260920
git push origin master
git push origin v1.0.20260920
```

`Desktop packages` 在 push/PR 中验证四个平台，只有版本标签触发 Release。四个平台全部成功后统一上传安装包及 `SHA256SUMS.txt`。
The workflow builds all four targets on push/PR; only a version tag publishes a release, after all builds succeed.
Manual dispatch builds artifacts without publishing. Artifact retention is 14 days; release assets persist.

当前公开版本是 `1.0.20260920`。Windows 的 PE/NSIS 16 位数字字段使用 `1.0.2026.920`，Tauri 内部以 `1.0.2026+920` 表达。
The public version and filenames remain unchanged; Windows uses this numeric mapping to avoid overflowing its 16-bit version fields.
