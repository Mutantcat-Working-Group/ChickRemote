# Local compatibility patches

Based on github.com/lwch/rdesktop v1.2.2, under the accompanying MIT license.
macOS capture uses ScreenCaptureKit (macOS 14+) instead of APIs removed from
modern SDKs. Capture failures propagate to callers, image conversion uses a
bounded RGBA bitmap context, and display sizing no longer requires a capture.
The display is not released because this library never captures it exclusively.
Capture, sizing and input consistently target the main display. Pointer input
converts captured pixels to CoreGraphics points for Retina display scaling.
