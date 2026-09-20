## 小鸡远程 / ChickReomte 1.0.20260920

- 使用项目 `logo.png` 更新三平台应用图标、桌面界面 Logo 和网页 favicon。
- Updated native application icons, desktop branding and favicon from the project `logo.png`.
- 重排中英文 README，统一居中 Logo、编号章节、平台支持表与项目结构。
- Restyled bilingual READMEs with centered branding, numbered sections, platform tables and project structure.
- 客户端、服务端、前端与 Tauri 版本统一升级至 `1.0.20260920`。
- Client, relay, frontend and Tauri versions updated to `1.0.20260920`.
- Windows x64 NSIS EXE with offline WebView2; macOS Intel / Apple Silicon DMGs; Linux x64 AppImage.
- macOS 应用及 DMG 使用 ad-hoc 签名，ScreenCaptureKit 截图后端要求 macOS 14+。
- macOS app and DMG are ad-hoc signed; ScreenCaptureKit requires macOS 14+.

### 安装须知 / Installation Notes

需要自行部署中继，并设置相同的强随机共享密钥。TLS 默认开启，需与中继匹配。
Deploy your relay first, configure a shared strong secret, and match TLS settings.

macOS ad-hoc 不等于 Apple 公证，首次运行可能需在系统设置中允许，并授予屏幕录制与辅助功能权限。
Ad-hoc signing is not notarization. Gatekeeper approval and Screen Recording / Accessibility permissions may be required.
Windows may show SmartScreen. Linux needs executable permission and FUSE 2 (or extract-and-run), and X11 for remote capture.

Verify installers against `SHA256SUMS.txt`. See `docs/desktop.md` for build, install, security and version-mapping details.
