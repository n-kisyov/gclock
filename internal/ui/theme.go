package ui

import (
	"syscall"
	"unsafe"
	"golang.org/x/sys/windows"
	"github.com/lxn/win"
	"github.com/n-kisyov/glock/internal/settings"
)

const (
	DWMWA_USE_IMMERSIVE_DARK_MODE = 20
	DWMWA_SYSTEMBACKDROP_TYPE     = 38
	DWMSBT_ACRYLIC                = 3
)

var (
	moduxtheme = windows.NewLazySystemDLL("uxtheme.dll")
	procSetWindowTheme = moduxtheme.NewProc("SetWindowTheme")
)

// removed CreateSolidBrush

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

func ThemeDialogColors(hDlg win.HWND, s *AppState, ctrl win.HWND, hdc win.HDC) win.HBRUSH {
	if !s.Settings.DarkMode {
		return 0
	}

	var clsBuf [32]uint16
	win.GetClassName(ctrl, &clsBuf[0], 32)
	cls := syscall.UTF16ToString(clsBuf[:])

	if cls == "Static" || cls == "Button" {
		win.SetTextColor(hdc, s.Colors.TextColor)
		win.SetBkColor(hdc, s.Colors.BgColor)
		win.SetBkMode(hdc, win.TRANSPARENT)
	}

	if cls == "Static" {
		style := win.GetWindowLong(ctrl, win.GWL_STYLE)
		if (style&win.SS_TYPEMASK) != win.SS_ICON && (style&win.SS_TYPEMASK) != win.SS_BITMAP {
			return 0
		}
	}

	if cls == "Edit" {
		win.SetTextColor(hdc, s.Colors.TextColor)
		win.SetBkColor(hdc, s.Colors.PanelBgColor)
	}
	
	if cls == "Edit" {
		return s.Colors.HPanelBrush
	}
	return s.Colors.HBgBrush
}

func ThemeDrawComboItem(s *AppState, dis *win.DRAWITEMSTRUCT) bool {
	if s == nil || dis.CtlType != ODT_COMBOBOX {
		return false
	}

	var buf [64]uint16
	if dis.ItemID == -1 || win.SendMessage(dis.HwndItem, win.CB_GETLBTEXT, uintptr(dis.ItemID), uintptr(unsafe.Pointer(&buf[0]))) == win.CB_ERR {
		buf[0] = 0
	}

	isField := (dis.ItemState & win.ODS_COMBOBOXEDIT) != 0
	var bg, fg win.COLORREF
	if (dis.ItemState&win.ODS_SELECTED) != 0 && !isField {
		bg = s.Colors.AccentColor
		fg = RGB(255, 255, 255)
	} else {
		if isField {
			bg = s.Colors.PanelBgColor
		} else {
			bg = s.Colors.BgColor
		}
		fg = s.Colors.TextColor
	}

	hBr := CreateSolidBrush(bg)
	FillRect(dis.HDC, &dis.RcItem, hBr)
	win.DeleteObject(win.HGDIOBJ(hBr))
	win.SetBkMode(dis.HDC, win.TRANSPARENT)
	win.SetTextColor(dis.HDC, fg)
	
	rc := dis.RcItem
	rc.Left += 4
	DrawText(dis.HDC, &buf[0], -1, &rc, win.DT_LEFT|win.DT_VCENTER|win.DT_SINGLELINE)
	
	if (dis.ItemState&win.ODS_FOCUS) != 0 && !isField {
		win.DrawFocusRect(dis.HDC, &dis.RcItem)
	}
	return true
}

func ThemeDialogInit(hDlg win.HWND, s *AppState) {
	ApplyTheme(hDlg, s.Settings.DarkMode, s.Settings.Acrylic)

	if !s.Settings.DarkMode {
		return
	}

	ctrl := win.GetWindow(hDlg, win.GW_CHILD)
	for ctrl != 0 {
		var clsBuf [32]uint16
		win.GetClassName(ctrl, &clsBuf[0], 32)
		cls := syscall.UTF16ToString(clsBuf[:])

		if cls == "Button" {
			style := win.GetWindowLong(ctrl, win.GWL_STYLE)
			typ := style & 0x0F
			if typ == win.BS_AUTOCHECKBOX || typ == win.BS_AUTORADIOBUTTON ||
				typ == win.BS_GROUPBOX || typ == win.BS_3STATE || typ == win.BS_AUTO3STATE {
				SetWindowTheme(ctrl, "", "")
			} else {
				SetWindowTheme(ctrl, "DarkMode_Explorer", "")
			}
		} else if cls == "ComboBox" {
			SetWindowTheme(ctrl, "", "")
		}
		ctrl = win.GetWindow(ctrl, win.GW_HWNDNEXT)
	}
	win.InvalidateRect(hDlg, nil, true)
}
