package ui

import (
	"syscall"
	"unsafe"
	"golang.org/x/sys/windows"
	"github.com/lxn/win"
	"github.com/user/gclock/internal/settings"
)

const (
	DWMWA_USE_IMMERSIVE_DARK_MODE = 20
	DWMWA_SYSTEMBACKDROP_TYPE     = 38
	DWMSBT_ACRYLIC                = 3
)

var (
	moduxtheme = windows.NewLazySystemDLL("uxtheme.dll")
	modgdi32   = windows.NewLazySystemDLL("gdi32.dll")
	procSetWindowTheme = moduxtheme.NewProc("SetWindowTheme")
	procCreateSolidBrush = modgdi32.NewProc("CreateSolidBrush")
)

func CreateSolidBrush(color win.COLORREF) win.HBRUSH {
	ret, _, _ := procCreateSolidBrush.Call(uintptr(color))
	return win.HBRUSH(ret)
}

func SetWindowTheme(hwnd win.HWND, appName string, idList string) {
	var pAppName *uint16
	if appName != "" {
		pAppName, _ = syscall.UTF16PtrFromString(appName)
	}
	var pIdList *uint16
	if idList != "" {
		pIdList, _ = syscall.UTF16PtrFromString(idList)
	}
	procSetWindowTheme.Call(uintptr(hwnd), uintptr(unsafe.Pointer(pAppName)), uintptr(unsafe.Pointer(pIdList)))
}

type ThemeColors struct {
	BgColor      win.COLORREF
	PanelBgColor win.COLORREF
	TextColor    win.COLORREF
	AccentColor  win.COLORREF
	ClockColor   win.COLORREF

	HBgBrush    win.HBRUSH
	HPanelBrush win.HBRUSH
}

func RGB(r, g, b byte) win.COLORREF {
	return win.COLORREF(uint32(r) | uint32(g)<<8 | uint32(b)<<16)
}

func UpdateColors(s *settings.Settings, colors *ThemeColors) {
	if colors.HBgBrush != 0 {
		win.DeleteObject(win.HGDIOBJ(colors.HBgBrush))
		colors.HBgBrush = 0
	}
	if colors.HPanelBrush != 0 {
		win.DeleteObject(win.HGDIOBJ(colors.HPanelBrush))
		colors.HPanelBrush = 0
	}

	if s.DarkMode {
		colors.BgColor = RGB(0x1E, 0x1E, 0x1E)
		colors.PanelBgColor = RGB(0x2D, 0x2D, 0x2D)
		colors.TextColor = RGB(0xE0, 0xE0, 0xE0)
		colors.AccentColor = RGB(0x00, 0x78, 0xD7)
		colors.ClockColor = RGB(0x00, 0xFF, 0x33)
	} else {
		colors.BgColor = RGB(0xFF, 0xFF, 0xFF)
		colors.PanelBgColor = RGB(0xF3, 0xF3, 0xF3)
		colors.TextColor = RGB(0x1E, 0x1E, 0x1E)
		colors.AccentColor = RGB(0x00, 0x78, 0xD7)
		colors.ClockColor = RGB(0x22, 0x22, 0x22)
	}

	colors.HBgBrush = CreateSolidBrush(colors.BgColor)
	colors.HPanelBrush = CreateSolidBrush(colors.PanelBgColor)
}

func ApplyTheme(hwnd win.HWND, dark bool, acrylic bool) {
	val := int32(0)
	if dark { val = 1 }
	windows.DwmSetWindowAttribute(windows.HWND(hwnd), DWMWA_USE_IMMERSIVE_DARK_MODE, unsafe.Pointer(&val), 4)

	backdrop := int32(0)
	if acrylic { backdrop = DWMSBT_ACRYLIC }
	windows.DwmSetWindowAttribute(windows.HWND(hwnd), DWMWA_SYSTEMBACKDROP_TYPE, unsafe.Pointer(&backdrop), 4)

	win.InvalidateRect(hwnd, nil, true)
	win.UpdateWindow(hwnd)

	child := win.GetWindow(hwnd, win.GW_CHILD)
	for child != 0 {
		win.InvalidateRect(child, nil, true)
		child = win.GetWindow(child, win.GW_HWNDNEXT)
	}
}
