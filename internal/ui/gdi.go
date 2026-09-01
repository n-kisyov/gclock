package ui

import (
	"unsafe"

	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

var (
	procCreatePen  = modgdi32.NewProc("CreatePen")
	procRoundRect  = modgdi32.NewProc("RoundRect")
	procRectangle  = modgdi32.NewProc("Rectangle")
	procExtTextOut = modgdi32.NewProc("ExtTextOutW")
	procDrawText   = windows.NewLazySystemDLL("user32.dll").NewProc("DrawTextW")
	procGetDateFormatEx = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetDateFormatEx")
)

const (
	PS_SOLID = 0
)

func CreatePen(fnPenStyle int32, nWidth int32, crColor win.COLORREF) win.HPEN {
	ret, _, _ := procCreatePen.Call(uintptr(fnPenStyle), uintptr(nWidth), uintptr(crColor))
	return win.HPEN(ret)
}

func ExtTextOut(hdc win.HDC, x, y int32, options uint32, rect *win.RECT, str *uint16, count int32, dx *int32) bool {
	ret, _, _ := procExtTextOut.Call(
		uintptr(hdc),
		uintptr(x),
		uintptr(y),
		uintptr(options),
		uintptr(unsafe.Pointer(rect)),
		uintptr(unsafe.Pointer(str)),
		uintptr(count),
		uintptr(unsafe.Pointer(dx)),
	)
	return ret != 0
}

func DrawText(hdc win.HDC, str *uint16, count int32, rect *win.RECT, format uint32) int32 {
	ret, _, _ := procDrawText.Call(
		uintptr(hdc),
		uintptr(unsafe.Pointer(str)),
		uintptr(count),
		uintptr(unsafe.Pointer(rect)),
		uintptr(format),
	)
	return int32(ret)
}

func RoundRect(hdc win.HDC, left, top, right, bottom, width, height int32) bool {
	ret, _, _ := procRoundRect.Call(
		uintptr(hdc),
		uintptr(left),
		uintptr(top),
		uintptr(right),
		uintptr(bottom),
		uintptr(width),
		uintptr(height),
	)
	return ret != 0
}

func Rectangle(hdc win.HDC, left, top, right, bottom int32) bool {
	ret, _, _ := procRectangle.Call(
		uintptr(hdc),
		uintptr(left),
		uintptr(top),
		uintptr(right),
		uintptr(bottom),
	)
	return ret != 0
}

func GetDateFormatEx(localeName *uint16, dwFlags uint32, lpDate *windows.Systemtime, lpFormat *uint16, lpDateStr *uint16, cchDate int32) int32 {
	ret, _, _ := procGetDateFormatEx.Call(
		uintptr(unsafe.Pointer(localeName)),
		uintptr(dwFlags),
		uintptr(unsafe.Pointer(lpDate)),
		uintptr(unsafe.Pointer(lpFormat)),
		uintptr(unsafe.Pointer(lpDateStr)),
		uintptr(cchDate),
	)
	return int32(ret)
}
