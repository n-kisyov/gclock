package main

import (
	"math/rand"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/win"
	"golang.org/x/sys/windows"

	"github.com/n-kisyov/glock/internal/audio"
	"github.com/n-kisyov/glock/internal/settings"
	"github.com/n-kisyov/glock/internal/ui"
)

var gState *ui.AppState
var audioManager *audio.SoundManager

func primaryDpi() int32 {
	hdc := win.GetDC(0)
	d := int32(96)
	if hdc != 0 {
		d = win.GetDeviceCaps(hdc, win.LOGPIXELSY)
		win.ReleaseDC(0, hdc)
	}
	if d == 0 {
		return 96
	}
	return d
}

func ensureOnScreen(x, y *int32, w, h int32) {
	// Fallback to simple screen limits if MonitorFromRect missing in lxn/win
	var work win.RECT
	if !win.SystemParametersInfo(48 /* SPI_GETWORKAREA */, 0, unsafe.Pointer(&work), 0) {
		work.Left = 0
		work.Top = 0
		work.Right = win.GetSystemMetrics(win.SM_CXSCREEN)
		work.Bottom = win.GetSystemMetrics(win.SM_CYSCREEN)
	}
	*x = work.Left + ((work.Right-work.Left)-w)/2
	*y = work.Top + ((work.Bottom-work.Top)-h)/2
	if *x < work.Left { *x = work.Left }
	if *y < work.Top { *y = work.Top }
}

func wndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_CREATE:
		gState.HMainWnd = hwnd
		
		// Start Audio
		audioManager = audio.NewSoundManager(gState.ExeDir, hwnd, gState.Settings)
		// audioManager.Settings is linked.

		ui.TrayCreate(hwnd, &gState.Nid)
		win.SetTimer(hwnd, ui.TIMER_CLOCK, 250, 0)
		return 0

	case ui.WM_TRAYICON:
		if lParam == win.WM_RBUTTONUP {
			ui.TrayShowMenu(hwnd)
		} else if lParam == win.WM_LBUTTONDBLCLK {
			if win.IsIconic(hwnd) {
				win.ShowWindow(hwnd, win.SW_RESTORE)
			} else {
				win.ShowWindow(hwnd, win.SW_SHOW)
			}
			win.SetForegroundWindow(hwnd)
		}
		return 0
		
	case win.WM_COMMAND:
		id := win.LOWORD(uint32(wParam))
		if id == ui.IDM_TRAY_SHOW {
			if win.IsIconic(hwnd) {
				win.ShowWindow(hwnd, win.SW_RESTORE)
			} else {
				win.ShowWindow(hwnd, win.SW_SHOW)
			}
			win.SetForegroundWindow(hwnd)
		} else if id == ui.IDM_TRAY_EXIT || id == ui.IDM_EXIT {
			win.PostMessage(hwnd, win.WM_CLOSE, 0, 0)
		}
		return 0

	case win.WM_TIMER:
		if wParam == ui.TIMER_CLOCK {
			win.InvalidateRect(hwnd, nil, false)
			ui.TrayUpdateTooltip(&gState.Nid, "GClock (Go Port)")
		} else if wParam == 1001 { // TimerSoundPreview
			audioManager.StopAlarm()
		}
		return 0

	case win.WM_PAINT:
		var ps win.PAINTSTRUCT
		hdc := win.BeginPaint(hwnd, &ps)
		
		var rc win.RECT
		win.GetClientRect(hwnd, &rc)
		
		// Fill background
		hOldBr := win.SelectObject(hdc, win.HGDIOBJ(gState.Colors.HBgBrush))
		hOldPn := win.SelectObject(hdc, win.HGDIOBJ(win.GetStockObject(win.NULL_PEN)))
		ui.Rectangle(hdc, rc.Left, rc.Top, rc.Right, rc.Bottom)
		win.SelectObject(hdc, hOldBr)
		win.SelectObject(hdc, hOldPn)
		
		// Draw Clock
		if gState.Settings.ClockStyle == "analog" {
			ui.ClockDrawAnalog(hdc, &rc, gState.Settings, gState.Colors.TextColor, gState.Colors.AccentColor)
		} else {
			t := time.Now()
			st := windows.Systemtime{
				Year:   uint16(t.Year()),
				Month:  uint16(t.Month()),
				Day:    uint16(t.Day()),
				Hour:   uint16(t.Hour()),
				Minute: uint16(t.Minute()),
				Second: uint16(t.Second()),
			}
			ui.ClockDrawDigital(hdc, &rc, &st, gState.Settings, &gState.Fonts, gState.AlarmActive, gState.Colors.ClockColor, gState.Colors.TextColor)
		}

		win.EndPaint(hwnd, &ps)
		return 0

	case win.WM_CLOSE:
		win.ShowWindow(hwnd, win.SW_HIDE)
		return 0
		
	case win.WM_DESTROY:
		ui.TrayRemove(&gState.Nid)
		if audioManager != nil {
			audioManager.Cleanup()
		}
		ui.ClockCleanup()
		win.PostQuitMessage(0)
		return 0
	}
	return win.DefWindowProc(hwnd, msg, wParam, lParam)
}

func main() {
	runtime.LockOSThread()
	
	var icc win.INITCOMMONCONTROLSEX
	icc.DwSize = uint32(unsafe.Sizeof(icc))
	icc.DwICC = win.ICC_STANDARD_CLASSES
	win.InitCommonControlsEx(&icc)

	gState = ui.NewAppState()
	rand.Seed(time.Now().UnixNano())

	settingsPath, _ := settings.GetSettingsPath()
	s, err := settings.Load(settingsPath)
	if err == nil {
		gState.Settings = s
	}

	ui.UpdateColors(gState.Settings, &gState.Colors)

	// Create Fonts
	var lf win.LOGFONT
	fontName, _ := syscall.UTF16FromString("Segoe UI")
	copy(lf.LfFaceName[:], fontName)
	lf.LfHeight = -70
	gState.Fonts.HClockFont = win.CreateFontIndirect(&lf)
	lf.LfHeight = -20
	gState.Fonts.HDateFont = win.CreateFontIndirect(&lf)

	className, _ := syscall.UTF16PtrFromString("GClockMainWnd")
	appName, _ := syscall.UTF16PtrFromString("GClock")

	var wc win.WNDCLASSEX
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.Style = win.CS_HREDRAW | win.CS_VREDRAW
	wc.LpfnWndProc = syscall.NewCallback(wndProc)
	wc.HInstance = win.GetModuleHandle(nil)
	wc.HIcon = win.LoadIcon(wc.HInstance, win.MAKEINTRESOURCE(ui.IDI_APPICON))
	wc.HCursor = win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_ARROW))
	wc.HbrBackground = win.HBRUSH(win.COLOR_WINDOW + 1)
	wc.LpszClassName = className
	wc.HIconSm = wc.HIcon

	if win.RegisterClassEx(&wc) == 0 {
		return
	}

	dpi := primaryDpi()
	gState.Dpi = dpi

	winW := int32(720 * dpi / 96)
	winH := int32(520 * dpi / 96)
	if gState.Settings.ClockStyle == "analog" {
		winW = int32(500 * dpi / 96)
		winH = int32(710 * dpi / 96)
	}
	winX := (win.GetSystemMetrics(win.SM_CXSCREEN) - winW) / 2
	winY := (win.GetSystemMetrics(win.SM_CYSCREEN) - winH) / 2

	ensureOnScreen(&winX, &winY, winW, winH)

	style := uint32(win.WS_OVERLAPPEDWINDOW)
	exStyle := uint32(0)
	if gState.Settings.AlwaysOnTop {
		exStyle |= win.WS_EX_TOPMOST
	}

	hwnd := win.CreateWindowEx(
		exStyle,
		className, appName,
		style,
		winX, winY, winW, winH,
		0, 0, wc.HInstance, nil)

	if hwnd == 0 {
		return
	}

	ui.ApplyTheme(hwnd, gState.Settings.DarkMode, gState.Settings.Acrylic)

	if !gState.Settings.StartMinimized {
		win.ShowWindow(hwnd, win.SW_SHOW)
	}
	win.UpdateWindow(hwnd)

	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) > 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}
}
