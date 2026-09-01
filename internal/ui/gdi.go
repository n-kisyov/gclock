package ui

import (
	"unsafe"

	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

var (
	procCreatePen  = modgdi32.NewProc("CreatePen")
	procExtTextOut = modgdi32.NewProc("ExtTextOutW")
	procDrawText   = windows.NewLazySystemDLL("user32.dll").NewProc("DrawTextW")
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
