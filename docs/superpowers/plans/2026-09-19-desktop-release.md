# ChickReomte 1.0.20260919 Desktop Release

Approved scope: Tauri 2 local desktop controller, bundled Go client, native
Windows NSIS / macOS Intel and Apple Silicon DMG / Linux AppImage releases.

## Execution

- [x] Modernize macOS capture and verify the Go client with current SDKs.
- [x] Implement and test validated desktop settings and owned process lifecycle.
- [x] Build a compact bilingual desktop control surface with local-only IPC.
- [x] Add reproducible native packaging, version checks, and tag-triggered CI.
- [x] Update bilingual documentation and installation/security limitations.
- [x] Run local tests/builds, push commits and version tag, inspect CI artifacts.

## Boundaries

The Go relay remains a separately deployed service. The desktop controller owns
its client process, keeps generated HTTP listeners on loopback,
and opens the existing session dashboard outside the privileged Tauri webview.
macOS ad-hoc signing does not provide Apple notarization. Windows uses an offline
WebView2 installer. Linux remote desktop capture requires an X11 session.
Detached jobs launched inside remote shells must be stopped separately.

## Verification

Go tests and native builds; Rust validation/configuration/lifecycle tests;
frontend build and desktop/mobile layout checks; version mapping tests;
all four native CI jobs; release filenames, signatures, and SHA-256 checksums.

## Release Evidence

- Published tag: `v1.0.20260919`, source commit `2b8b547`.
- [Release](https://github.com/Mutantcat-Working-Group/ChickRemote/releases/tag/v1.0.20260919)
- [Native build and release run](https://github.com/Mutantcat-Working-Group/ChickRemote/actions/runs/35431689198): all four targets and publication passed.
- Windows silent installation and installed application launch passed; Linux
  extracted AppImage launch passed under Xvfb; both macOS DMGs passed app launch,
  app/DMG signature verification and disk-image verification.
- All four installer SHA-256 checks passed before publication.
- [CodeQL](https://github.com/Mutantcat-Working-Group/ChickRemote/actions/runs/35431689414)
  passed; no open CodeQL alerts at publication time.
- Local full Go tests, `go vet`, and all three Node packaging tests passed.

Launch checks are smoke tests, not a complete interactive remote-control test.
Screen Recording/Accessibility permissions and physical input/capture still
require verification on user machines. Ad-hoc signatures do not bypass
Gatekeeper, and unsigned Windows installers may trigger SmartScreen.
