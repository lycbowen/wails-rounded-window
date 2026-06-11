//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	wmSize       = 0x0005
	wmDestroy    = 0x0002
	roundingSize = 48
)

var gwlWndProc = ^uintptr(3)

type rect struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

type osVersionInfoEx struct {
	osVersionInfoSize uint32
	majorVersion      uint32
	minorVersion      uint32
	buildNumber       uint32
	platformID        uint32
	csdVersion        [128]uint16
	servicePackMajor  uint16
	servicePackMinor  uint16
	suiteMask         uint16
	productType       byte
	reserved          byte
}

var (
	user32                 = syscall.NewLazyDLL("user32.dll")
	gdi32                  = syscall.NewLazyDLL("gdi32.dll")
	ntdll                  = syscall.NewLazyDLL("ntdll.dll")
	procFindWindowW        = user32.NewProc("FindWindowW")
	procGetWindowRect      = user32.NewProc("GetWindowRect")
	procSetWindowRgn       = user32.NewProc("SetWindowRgn")
	procGetWindowLongPtrW  = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW  = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProcW    = user32.NewProc("CallWindowProcW")
	procCreateRoundRectRgn = gdi32.NewProc("CreateRoundRectRgn")
	procDeleteObject       = gdi32.NewProc("DeleteObject")
	procRtlGetVersion      = ntdll.NewProc("RtlGetVersion")
)

var (
	roundedWindowHandle uintptr
	previousWindowProc  uintptr
	windowProcCallback  = syscall.NewCallback(roundedWindowProc)
)

func applyNativeWindowStyle() {
	if isWindows11OrNewer() {
		return
	}

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
	case wmSize:
		applyRoundedWindowRegion(hwnd)
	case wmDestroy:
		roundedWindowHandle = 0
	}

	if previousWindowProc != 0 {
		result, _, _ := procCallWindowProcW.Call(previousWindowProc, hwnd, uintptr(msg), wParam, lParam)
		return result
	}

	return 0
}

func applyRoundedWindowRegion(hwnd uintptr) {
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

	result, _, _ := procSetWindowRgn.Call(hwnd, region, 1)
	if result == 0 {
		procDeleteObject.Call(region)
	}
}

func isWindows11OrNewer() bool {
	var version osVersionInfoEx
	version.osVersionInfoSize = uint32(unsafe.Sizeof(version))

	result, _, _ := procRtlGetVersion.Call(uintptr(unsafe.Pointer(&version)))
	if result != 0 {
		return false
	}

	return version.majorVersion > 10 || (version.majorVersion == 10 && version.buildNumber >= 22000)
}
