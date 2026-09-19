# ChickReomte 1.0.20260919 Desktop Release

Approved scope: Tauri 2 local desktop controller, bundled Go client, native
Windows NSIS / macOS Intel and Apple Silicon DMG / Linux AppImage releases.

## Execution

- [ ] Modernize macOS capture and verify the Go client with current SDKs.
- [ ] Implement and test validated desktop settings and owned process lifecycle.
- [ ] Build a compact bilingual desktop control surface with local-only IPC.
- [ ] Add reproducible native packaging, version checks, and tag-triggered CI.
- [ ] Update bilingual documentation and installation/security limitations.
- [ ] Run local tests/builds, push commits and version tag, inspect CI artifacts.

## Boundaries

The Go relay remains a separately deployed service. The desktop controller owns
its client and descendant processes, keeps generated HTTP listeners on loopback,
and opens the existing session dashboard outside the privileged Tauri webview.
macOS ad-hoc signing does not provide Apple notarization. Windows uses an offline
WebView2 installer. Linux remote desktop capture requires an X11 session.

## Verification

Go tests and native builds; Rust validation/configuration/lifecycle tests;
frontend build and desktop/mobile layout checks; version mapping tests;
all four native CI jobs; release filenames, signatures, and SHA-256 checksums.
