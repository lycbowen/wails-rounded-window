//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	wmDestroy       = 0x0002
	wmSize          = 0x0005
	wmDisplayChange = 0x007E
	wmEnterSizeMove = 0x0231
	wmExitSizeMove  = 0x0232
	wmDpiChanged    = 0x02E0

	sizeMinimized = 1
	sizeMaximized = 2

	gwlStyle     = ^uintptr(15)
	wsMaximize   = 0x01000000
	wsMinimize   = 0x20000000
	baseDPI      = 96
	cornerRadius = 24
	resizeHandle = 16
	rgnOr        = 2
)

var gwlWndProc = ^uintptr(3)

type rect struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

type monitorInfo struct {
	cbSize    uint32
	rcMonitor rect
	rcWork    rect
	dwFlags   uint32
}

var (
	user32                 = syscall.NewLazyDLL("user32.dll")
	gdi32                  = syscall.NewLazyDLL("gdi32.dll")
	procFindWindowW        = user32.NewProc("FindWindowW")
	procGetWindowRect      = user32.NewProc("GetWindowRect")
	procSetWindowRgn       = user32.NewProc("SetWindowRgn")
	procGetDpiForWindow    = user32.NewProc("GetDpiForWindow")
	procGetWindowLongPtrW  = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW  = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProcW    = user32.NewProc("CallWindowProcW")
	procMonitorFromWindow  = user32.NewProc("MonitorFromWindow")
	procGetMonitorInfoW    = user32.NewProc("GetMonitorInfoW")
	procCreateRoundRectRgn = gdi32.NewProc("CreateRoundRectRgn")
	procCreateRectRgn      = gdi32.NewProc("CreateRectRgn")
	procCombineRgn         = gdi32.NewProc("CombineRgn")
	procDeleteObject       = gdi32.NewProc("DeleteObject")
)

var (
	roundedWindowHandle uintptr
	previousWindowProc  uintptr
	windowProcCallback  = syscall.NewCallback(roundedWindowProc)
	isInteractiveSizing bool
	lastRegionWidth     int32
	lastRegionHeight    int32
	lastRegionDPI       uintptr
	regionIsCleared     bool
)

func applyNativeWindowStyle() {
	title, err := syscall.UTF16PtrFromString(appTitle)
	if err != nil {
		return
	}

	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(title)))
	if hwnd == 0 {
		return
	}

	roundedWindowHandle = hwnd
	applyRoundedWindowRegion(hwnd)

	if previousWindowProc == 0 {
		previousWindowProc, _, _ = procGetWindowLongPtrW.Call(hwnd, gwlWndProc)
		procSetWindowLongPtrW.Call(hwnd, gwlWndProc, windowProcCallback)
	}
}

func roundedWindowProc(hwnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	switch msg {
	case wmEnterSizeMove:
		isInteractiveSizing = true
		clearRoundedWindowRegion(hwnd)
	case wmSize:
		if wParam == sizeMinimized || wParam == sizeMaximized {
			clearRoundedWindowRegion(hwnd)
			break
		}
		if isInteractiveSizing {
			break
		}
		applyRoundedWindowRegion(hwnd)
	case wmExitSizeMove:
		isInteractiveSizing = false
		applyRoundedWindowRegion(hwnd)
	case wmDpiChanged, wmDisplayChange:
		applyRoundedWindowRegion(hwnd)
	case wmDestroy:
		clearRoundedWindowRegion(hwnd)
		roundedWindowHandle = 0
		isInteractiveSizing = false
	}

	if previousWindowProc != 0 {
		result, _, _ := procCallWindowProcW.Call(previousWindowProc, hwnd, uintptr(msg), wParam, lParam)
		return result
	}

	return 0
}

func applyRoundedWindowRegion(hwnd uintptr) {
	if isWindowMinimized(hwnd) || isWindowMaximized(hwnd) || isWindowFullscreen(hwnd) {
		clearRoundedWindowRegion(hwnd)
		return
	}

	var windowRect rect
	ok, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&windowRect)))
	if ok == 0 {
		return
	}

	width := windowRect.right - windowRect.left
	height := windowRect.bottom - windowRect.top
	if width <= 0 || height <= 0 {
		return
	}

	dpi := windowDPI(hwnd)
	if !regionIsCleared && width == lastRegionWidth && height == lastRegionHeight && dpi == lastRegionDPI {
		return
	}

	roundingSize := scaledRoundingSize(dpi)
	hitSize := scaledResizeHandleSize(dpi)
	region, _, _ := procCreateRoundRectRgn.Call(
		0,
		0,
		uintptr(width+1),
		uintptr(height+1),
		roundingSize,
		roundingSize,
	)
	if region == 0 {
		return
	}
	addCornerResizeRegions(region, width, height, hitSize)

	result, _, _ := procSetWindowRgn.Call(hwnd, region, 1)
	if result == 0 {
		procDeleteObject.Call(region)
		return
	}

	lastRegionWidth = width
	lastRegionHeight = height
	lastRegionDPI = dpi
	regionIsCleared = false
}

func clearRoundedWindowRegion(hwnd uintptr) {
	if regionIsCleared {
		return
	}

	procSetWindowRgn.Call(hwnd, 0, 1)
	lastRegionWidth = 0
	lastRegionHeight = 0
	lastRegionDPI = 0
	regionIsCleared = true
}

func windowDPI(hwnd uintptr) uintptr {
	dpi, _, _ := procGetDpiForWindow.Call(hwnd)
	if dpi == 0 {
		dpi = baseDPI
	}

	return dpi
}

func scaledRoundingSize(dpi uintptr) uintptr {
	return uintptr(cornerRadius * 2 * int32(dpi) / baseDPI)
}

func scaledResizeHandleSize(dpi uintptr) int32 {
	return resizeHandle * int32(dpi) / baseDPI
}

func addCornerResizeRegions(region uintptr, width, height, hitSize int32) {
	if hitSize <= 0 || hitSize > width || hitSize > height {
		return
	}

	unionRectRegion(region, 0, 0, hitSize, hitSize)
	unionRectRegion(region, width-hitSize, 0, width+1, hitSize)
	unionRectRegion(region, 0, height-hitSize, hitSize, height+1)
	unionRectRegion(region, width-hitSize, height-hitSize, width+1, height+1)
}

func unionRectRegion(region uintptr, left, top, right, bottom int32) {
	rectRegion, _, _ := procCreateRectRgn.Call(
		uintptr(left),
		uintptr(top),
		uintptr(right),
		uintptr(bottom),
	)
	if rectRegion == 0 {
		return
	}

	procCombineRgn.Call(region, region, rectRegion, rgnOr)
	procDeleteObject.Call(rectRegion)
}

func isWindowMaximized(hwnd uintptr) bool {
	return windowStyle(hwnd)&wsMaximize != 0
}

func isWindowMinimized(hwnd uintptr) bool {
	return windowStyle(hwnd)&wsMinimize != 0
}

func windowStyle(hwnd uintptr) uintptr {
	style, _, _ := procGetWindowLongPtrW.Call(hwnd, gwlStyle)
	return style
}

func isWindowFullscreen(hwnd uintptr) bool {
	var windowRect rect
	ok, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&windowRect)))
	if ok == 0 {
		return false
	}

	monitor, _, _ := procMonitorFromWindow.Call(hwnd, 2)
	if monitor == 0 {
		return false
	}

	var info monitorInfo
	info.cbSize = uint32(unsafe.Sizeof(info))
	ok, _, _ = procGetMonitorInfoW.Call(monitor, uintptr(unsafe.Pointer(&info)))
	if ok == 0 {
		return false
	}

	return windowRect.left == info.rcMonitor.left &&
		windowRect.top == info.rcMonitor.top &&
		windowRect.right == info.rcMonitor.right &&
		windowRect.bottom == info.rcMonitor.bottom
}
