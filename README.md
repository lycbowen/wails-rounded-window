# wails-rounded-window

A Wails v2 demo for building a frameless macOS desktop window with rounded
corners, custom window controls, resize support, and native glass-style effects.

## Features

- Frameless Wails window with a custom React titlebar.
- macOS native vibrancy blur using `NSVisualEffectView`.
- macOS rounded native glass layer without breaking resize hit testing.
- Transparent WebView background so the native window shape can show through.

## Platform Notes

### macOS

The app enables Wails' native translucent window and then applies a small
Objective-C helper after DOM ready:

- keeps the `NSWindow` background transparent;
- clips only the `NSVisualEffectView` layer to 24px rounded corners;
- leaves the content view unmasked so native resize hit testing still works.

## Development

Install frontend dependencies:

```bash
cd frontend
npm install
```

Run in live development mode:

```bash
wails dev
```

Build the frontend:

```bash
cd frontend
npm run build
```

Check the Go side:

```bash
go build ./...
```

## Important Files

- `main.go` configures the Wails window and platform options.
- `native_window_darwin.go` contains the macOS native vibrancy/rounding helper.
- `native_window_other.go` keeps non-macOS platforms as a no-op.
- `frontend/src/App.tsx` implements the custom titlebar and window controls.
- `frontend/src/App.css` contains the rounded shell and visual styling.
