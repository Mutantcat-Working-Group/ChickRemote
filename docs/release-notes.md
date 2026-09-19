## 小鸡远程 / ChickReomte 1.0.20260919

- Tauri 2 桌面客户端，内置 Go 引擎，中英文界面及连接配置。
- Tauri 2 desktop client with a bundled Go engine, bilingual UI and connection settings.
- Windows x64 NSIS EXE with offline WebView2; macOS Intel / Apple Silicon DMGs; Linux x64 AppImage.
- macOS 应用及 DMG 使用 ad-hoc 签名，截图后端迁移至 ScreenCaptureKit（macOS 14+）。
- macOS app and DMG are ad-hoc signed; ScreenCaptureKit requires macOS 14+.
- 品牌统一为小鸡远程 / ChickReomte，包名 org.mutantcat.chickreomte，MIT 协议。
- Remote-session stability fixes, MIT licensing and refreshed bilingual documentation.

### 安装须知 / Installation Notes

需要自行部署中继，并设置相同的强随机共享密钥。TLS 默认开启，需与中继匹配。
Deploy your relay first, configure a shared strong secret, and match TLS settings.

macOS ad-hoc 不等于 Apple 公证，首次运行可能需在系统设置中允许，并授予屏幕录制与辅助功能权限。
Ad-hoc signing is not notarization. Gatekeeper approval and Screen Recording / Accessibility permissions may be required.
Windows may show SmartScreen. Linux needs executable permission and FUSE 2 (or extract-and-run), and X11 for remote capture.

Verify installers against `SHA256SUMS.txt`. See `docs/desktop.md` for build, install, security and version-mapping details.
