package ui

import (
	
	"unsafe"

	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

var (
	modgdi32       = windows.NewLazySystemDLL("gdi32.dll")
	procCreatePen  = modgdi32.NewProc("CreatePen")
	procRoundRect  = modgdi32.NewProc("RoundRect")
	procRectangle  = modgdi32.NewProc("Rectangle")
	procExtTextOut = modgdi32.NewProc("ExtTextOutW")
	
	moduser32      = windows.NewLazySystemDLL("user32.dll")
	procDrawText   = moduser32.NewProc("DrawTextW")
	procSetDlgItemText = moduser32.NewProc("SetDlgItemTextW")
	procGetDlgItemText = moduser32.NewProc("GetDlgItemTextW")
	procSendDlgItemMessage = moduser32.NewProc("SendDlgItemMessageW")
	procCheckDlgButton = moduser32.NewProc("CheckDlgButton")
	procIsDlgButtonChecked = moduser32.NewProc("IsDlgButtonChecked")
	procFillRect       = moduser32.NewProc("FillRect")

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

func SetDlgItemText(hDlg win.HWND, nIDDlgItem int32, lpString string) bool {
	str, _ := windows.UTF16FromString(lpString)
	ret, _, _ := procSetDlgItemText.Call(
		uintptr(hDlg),
		uintptr(nIDDlgItem),
		uintptr(unsafe.Pointer(&str[0])),
	)
	return ret != 0
}

func GetDlgItemText(hDlg win.HWND, nIDDlgItem int32, lpString *uint16, cchMax int32) int32 {
	ret, _, _ := procGetDlgItemText.Call(
		uintptr(hDlg),
		uintptr(nIDDlgItem),
		uintptr(unsafe.Pointer(lpString)),
		uintptr(cchMax),
	)
	return int32(ret)
}

func SendDlgItemMessage(hDlg win.HWND, nIDDlgItem int32, Msg uint32, wParam uintptr, lParam uintptr) uintptr {
	ret, _, _ := procSendDlgItemMessage.Call(
		uintptr(hDlg),
		uintptr(nIDDlgItem),
		uintptr(Msg),
		uintptr(wParam),
		uintptr(lParam),
	)
	return ret
}

func CheckDlgButton(hDlg win.HWND, nIDButton int32, uCheck uint32) bool {
	ret, _, _ := procCheckDlgButton.Call(
		uintptr(hDlg),
		uintptr(nIDButton),
		uintptr(uCheck),
	)
	return ret != 0
}

func IsDlgButtonChecked(hDlg win.HWND, nIDButton int32) uint32 {
	ret, _, _ := procIsDlgButtonChecked.Call(
		uintptr(hDlg),
		uintptr(nIDButton),
	)
	return uint32(ret)
}

func FillRect(hdc win.HDC, lprc *win.RECT, hbr win.HBRUSH) int32 {
	ret, _, _ := procFillRect.Call(
		uintptr(hdc),
		uintptr(unsafe.Pointer(lprc)),
		uintptr(hbr),
	)
	return int32(ret)
}

func PtInRect(rect *win.RECT, pt win.POINT) bool {
	return pt.X >= rect.Left && pt.X < rect.Right && pt.Y >= rect.Top && pt.Y < rect.Bottom
}

var procCreateSolidBrush = modgdi32.NewProc("CreateSolidBrush")

func CreateSolidBrush(color win.COLORREF) win.HBRUSH {
	ret, _, _ := procCreateSolidBrush.Call(uintptr(color))
	return win.HBRUSH(ret)
}

