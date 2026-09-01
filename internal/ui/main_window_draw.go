package ui

import (
	"fmt"
	"syscall"
	"time"
	"github.com/lxn/win"
	"github.com/n-kisyov/gclock/internal/settings"
)

const (
	ALARM_PAD_X      = 10
	ALARM_PAD_Y      = 8
	ALARM_ROW_H      = 30
	ALARM_HEADER_H   = 22
	ALARM_CHK_SIZE   = 18
	ALARM_BTN_W      = 52
	ALARM_BTN_H      = 22
	ALARM_BTN_GAP    = 5
	SEP_MARGIN       = 8

	MODE_BAR_H       = 32
	MODE_BAR_GAP     = 10
	MODE_BAR_BOTTOM  = 16
)

type HitKind int

const (
	HT_NONE HitKind = iota
	HT_MODE
	HT_SETTINGS
	HT_COLLAPSE
	HT_ALARM_CHECK
	HT_ALARM_EDIT
	HT_ALARM_CLEAR
)

type HitTarget struct {
	Kind  HitKind
	Index int
	Rect  win.RECT
}

type BtnState int

const (
	BTN_NORMAL BtnState = iota
	BTN_HOVER
	BTN_PRESSED
)

const (
	MB_NONE = iota
	MB_TO_CLOCK
	MB_TO_TIMER
	MB_TO_STOPWATCH
	MB_SNOOZE
	MB_DISMISS
	MB_CD_START
	MB_CD_PAUSE
	MB_CD_SET
	MB_CD_RESET
	MB_SW_START
	MB_SW_STOP
	MB_SW_RESET
	MB_SLEEP
)

type ModeItem struct {
	Id        int
	Text      string
	Width     int32
	Bg        win.COLORREF
	Fg        win.COLORREF
	IsLabel   bool
	Highlight bool
	Rect      win.RECT
}

type ModeBar struct {
	Items []ModeItem
}

func (s *AppState) Scale(v int32) int32 {
	dpi := s.Dpi
	if dpi == 0 {
		dpi = 96
	}
	return (v * int32(dpi)) / 96
}

func shade(c win.COLORREF, delta int32) win.COLORREF {
	r := int32(c&0xFF) + delta
	g := int32((c>>8)&0xFF) + delta
	b := int32((c>>16)&0xFF) + delta

	if r < 0 { r = 0 }
	if r > 255 { r = 255 }
	if g < 0 { g = 0 }
	if g > 255 { g = 255 }
	if b < 0 { b = 0 }
	if b > 255 { b = 255 }
	return RGB(byte(r), byte(g), byte(b))
}

func btnState(s *AppState, kind HitKind, index int) BtnState {
	if s.PressedTarget.Kind == kind && s.PressedTarget.Index == index {
		if s.PressedIn {
			return BTN_PRESSED
		}
		return BTN_HOVER
	}
	if s.PressedTarget.Kind == HT_NONE && s.HoverTarget.Kind == kind && s.HoverTarget.Index == index {
		return BTN_HOVER
	}
	return BTN_NORMAL
}

func drawButton(hdc win.HDC, r *win.RECT, text string, bg, fg win.COLORREF, st BtnState, hFont win.HFONT) {
	if st == BTN_HOVER {
		bg = shade(bg, 22)
	} else if st == BTN_PRESSED {
		bg = shade(bg, -28)
	}

	hBr := CreateSolidBrush(bg)
	hPn := CreatePen(win.PS_SOLID, 1, fg)
	hOldBr := win.SelectObject(hdc, win.HGDIOBJ(hBr))
	hOldPn := win.SelectObject(hdc, win.HGDIOBJ(hPn))

	RoundRect(hdc, r.Left, r.Top, r.Right, r.Bottom, 4, 4)

	win.SelectObject(hdc, hOldBr)
	win.SelectObject(hdc, hOldPn)
	win.DeleteObject(win.HGDIOBJ(hBr))
	win.DeleteObject(win.HGDIOBJ(hPn))

	win.SetBkMode(hdc, win.TRANSPARENT)
	win.SetTextColor(hdc, fg)

	hOldFont := win.SelectObject(hdc, win.HGDIOBJ(hFont))
	str, _ := syscall.UTF16FromString(text)
	DrawText(hdc, &str[0], -1, r, win.DT_CENTER|win.DT_VCENTER|win.DT_SINGLELINE)
	win.SelectObject(hdc, hOldFont)
}

func drawHighlightedButton(hdc win.HDC, r *win.RECT, text string, bg, fg win.COLORREF, st BtnState, hFont win.HFONT) {
	if st == BTN_HOVER {
		bg = shade(bg, 22)
	} else if st == BTN_PRESSED {
		bg = shade(bg, -28)
	}

	hBr := CreateSolidBrush(bg)
	hPn := CreatePen(win.PS_SOLID, 2, fg)
	hOldBr := win.SelectObject(hdc, win.HGDIOBJ(hBr))
	hOldPn := win.SelectObject(hdc, win.HGDIOBJ(hPn))

	RoundRect(hdc, r.Left, r.Top, r.Right, r.Bottom, 6, 6)

	win.SelectObject(hdc, hOldBr)
	win.SelectObject(hdc, hOldPn)
	win.DeleteObject(win.HGDIOBJ(hBr))
	win.DeleteObject(win.HGDIOBJ(hPn))

	win.SetBkMode(hdc, win.TRANSPARENT)
	win.SetTextColor(hdc, fg)

	hOldFont := win.SelectObject(hdc, win.HGDIOBJ(hFont))
	str, _ := syscall.UTF16FromString(text)
	DrawText(hdc, &str[0], -1, r, win.DT_CENTER|win.DT_VCENTER|win.DT_SINGLELINE)
	win.SelectObject(hdc, hOldFont)
}

func calcAlarmRects(s *AppState, hwnd win.HWND, panel, header, srcClock *win.RECT) {
	var cr win.RECT
	win.GetClientRect(hwnd, &cr)
	cw := cr.Right - cr.Left
	sepY := srcClock.Bottom
	panW := cw - s.Scale(ALARM_PAD_X)*2
	panX := s.Scale(ALARM_PAD_X)
	panY := sepY + s.Scale(4)
	availH := cr.Bottom - panY - s.Scale(ALARM_PAD_Y)

	rowCount := s.Settings.AlarmCount
	if s.AlarmsCollapsed {
		rowCount = 0
	}
	minH := s.Scale(ALARM_HEADER_H + 17)
	maxH := s.Scale(ALARM_HEADER_H+6) + int32(rowCount)*s.Scale(ALARM_ROW_H) + s.Scale(21)

	panH := availH
	if panH < minH {
		panH = minH
	}
	if panH > maxH {
		panH = maxH
	}

	panel.Left = panX
	panel.Top = panY
	panel.Right = panX + panW
	panel.Bottom = panY + panH
	header.Left = panX + s.Scale(12)
	header.Top = panY + s.Scale(6)
	header.Right = panX + panW - s.Scale(12)
	header.Bottom = header.Top + s.Scale(ALARM_HEADER_H)
}

func alarmVisibleRows(s *AppState, panel, header *win.RECT) int {
	avail := panel.Bottom - (header.Bottom + s.Scale(2))
	n := 0
	if avail > 0 {
		n = int(avail / s.Scale(ALARM_ROW_H))
	}
	if n > s.Settings.AlarmCount {
		n = s.Settings.AlarmCount
	}
	return n
}

func getAlarmRowRect(s *AppState, panel, header *win.RECT, idx int) win.RECT {
	var r win.RECT
	baseY := header.Bottom + s.Scale(2)
	r.Left = panel.Left + s.Scale(8)
	r.Top = baseY + int32(idx)*s.Scale(ALARM_ROW_H)
	r.Right = panel.Right - s.Scale(8)
	r.Bottom = r.Top + s.Scale(ALARM_ROW_H)
	return r
}

func getCheckRect(s *AppState, row *win.RECT) win.RECT {
	var r win.RECT
	cy := (row.Top + row.Bottom) / 2
	r.Left = row.Left + s.Scale(ALARM_PAD_X)
	r.Top = cy - s.Scale(ALARM_CHK_SIZE)/2
	r.Right = r.Left + s.Scale(ALARM_CHK_SIZE)
	r.Bottom = r.Top + s.Scale(ALARM_CHK_SIZE)
	return r
}

func getClearRect(s *AppState, row *win.RECT) win.RECT {
	var r win.RECT
	r.Right = row.Right - s.Scale(8)
	r.Left = r.Right - s.Scale(ALARM_BTN_W)
	r.Top = row.Top + (s.Scale(ALARM_ROW_H)-s.Scale(ALARM_BTN_H))/2
	r.Bottom = r.Top + s.Scale(ALARM_BTN_H)
	return r
}

func getEditRect(s *AppState, row *win.RECT) win.RECT {
	clear := getClearRect(s, row)
	var r win.RECT
	r.Right = clear.Left - s.Scale(ALARM_BTN_GAP)
	r.Left = r.Right - s.Scale(ALARM_BTN_W)
	r.Top = clear.Top
	r.Bottom = clear.Bottom
	return r
}

func getSettingsRect(s *AppState, header *win.RECT) win.RECT {
	var r win.RECT
	r.Right = header.Right - s.Scale(26)
	r.Left = r.Right - s.Scale(62)
	r.Top = header.Top
	r.Bottom = header.Top + s.Scale(ALARM_HEADER_H)
	return r
}

func buildModeBar(s *AppState, clockRect *win.RECT, bar *ModeBar) {
	white := RGB(255, 255, 255)
	neutral := RGB(0xE0, 0xE0, 0xE0)
	if s.Settings.DarkMode {
		neutral = RGB(0x45, 0x45, 0x45)
	}
	resetBg := RGB(0xC0, 0x50, 0x50)

	bar.Items = nil

	if s.AlarmActive {
		bar.Items = append(bar.Items, ModeItem{Id: MB_SNOOZE, Text: "SNOOZE", Width: 102, Bg: RGB(0xDE, 0x87, 0x00), Fg: white})
		bar.Items = append(bar.Items, ModeItem{Id: MB_DISMISS, Text: "DISMISS", Width: 102, Bg: RGB(0xE8, 0x11, 0x23), Fg: white})
	} else if s.SnoozePending {
		now := time.Now()
		remain := s.SnoozeEnd.Sub(now)
		if remain < 0 {
			remain = 0
		}
		rs := int(remain.Seconds())
		buf := fmt.Sprintf("Snoozed  %d:%02d", rs/60, rs%60)
		bar.Items = append(bar.Items, ModeItem{Text: buf, Width: 96, Fg: s.Colors.TextColor, IsLabel: true})
		bar.Items = append(bar.Items, ModeItem{Id: MB_DISMISS, Text: "CANCEL", Width: 104, Bg: RGB(0xE8, 0x11, 0x23), Fg: white})
	} else if s.Settings.AppMode == 1 { // countdown
		bar.Items = append(bar.Items, ModeItem{Id: MB_TO_CLOCK, Text: "Clock", Width: 56, Bg: neutral, Fg: s.Colors.TextColor})
		if s.CdRunning {
			bar.Items = append(bar.Items, ModeItem{Id: MB_CD_PAUSE, Text: "Pause", Width: 62, Bg: s.Colors.AccentColor, Fg: white})
		} else {
			if s.CdRemainingMs > 0 {
				bar.Items = append(bar.Items, ModeItem{Id: MB_CD_START, Text: "Start", Width: 62, Bg: RGB(0x00, 0x88, 0x00), Fg: white})
			}
			bar.Items = append(bar.Items, ModeItem{Id: MB_CD_SET, Text: "Set", Width: 62, Bg: s.Colors.AccentColor, Fg: white})
		}
		bar.Items = append(bar.Items, ModeItem{Id: MB_CD_RESET, Text: "Reset", Width: 62, Bg: resetBg, Fg: white})
	} else if s.Settings.AppMode == 2 { // stopwatch
		bar.Items = append(bar.Items, ModeItem{Id: MB_TO_CLOCK, Text: "Clock", Width: 56, Bg: neutral, Fg: s.Colors.TextColor})
		if s.SwRunning {
			bar.Items = append(bar.Items, ModeItem{Id: MB_SW_STOP, Text: "Stop", Width: 62, Bg: RGB(0xCC, 0x33, 0x00), Fg: white})
		} else {
			bar.Items = append(bar.Items, ModeItem{Id: MB_SW_START, Text: "Start", Width: 62, Bg: RGB(0x00, 0x88, 0x00), Fg: white})
		}
		bar.Items = append(bar.Items, ModeItem{Id: MB_SW_RESET, Text: "Reset", Width: 62, Bg: resetBg, Fg: white})
	} else {
		bar.Items = append(bar.Items, ModeItem{Id: MB_TO_CLOCK, Text: "Clock", Width: 70, Bg: s.Colors.AccentColor, Fg: white, Highlight: true})

		cdFinished := s.CdRemainingMs <= 0 && (s.Settings.CDHours+s.Settings.CDMins+s.Settings.CDSecs > 0)
		if s.CdRunning {
			bar.Items = append(bar.Items, ModeItem{Id: MB_TO_TIMER, Text: "Timer", Width: 68, Bg: RGB(0x20, 0x80, 0x20), Fg: white, Highlight: true})
		} else if cdFinished {
			bar.Items = append(bar.Items, ModeItem{Id: MB_TO_TIMER, Text: "Finished!", Width: 68, Bg: RGB(0xC0, 0x30, 0x30), Fg: white})
		} else {
			bar.Items = append(bar.Items, ModeItem{Id: MB_TO_TIMER, Text: "Timer", Width: 68, Bg: neutral, Fg: s.Colors.TextColor})
		}

		if s.SwRunning {
			bar.Items = append(bar.Items, ModeItem{Id: MB_TO_STOPWATCH, Text: "Stopw.", Width: 67, Bg: RGB(0x20, 0x80, 0x20), Fg: white, Highlight: true})
		} else {
			bar.Items = append(bar.Items, ModeItem{Id: MB_TO_STOPWATCH, Text: "Stopw.", Width: 67, Bg: neutral, Fg: s.Colors.TextColor})
		}

		if s.SleepRunning {
			now := time.Now()
			remain := s.SleepEnd.Sub(now)
			if remain < 0 {
				remain = 0
			}
			rs := int(remain.Seconds())
			buf := fmt.Sprintf("Sleep %d:%02d", rs/60, rs%60)
			bar.Items = append(bar.Items, ModeItem{Id: MB_SLEEP, Text: buf, Width: 92, Bg: RGB(0x5A, 0x3E, 0x8C), Fg: white, Highlight: true})
		} else {
			bar.Items = append(bar.Items, ModeItem{Id: MB_SLEEP, Text: "Sleep", Width: 60, Bg: neutral, Fg: s.Colors.TextColor})
		}
	}

	// Layout
	total := int32(0)
	for i := range bar.Items {
		bar.Items[i].Width = s.Scale(bar.Items[i].Width)
		total += bar.Items[i].Width
		if i > 0 {
			total += s.Scale(MODE_BAR_GAP)
		}
	}
	cx := clockRect.Left + (clockRect.Right-clockRect.Left)/2
	top := clockRect.Bottom - s.Scale(MODE_BAR_H) - s.Scale(MODE_BAR_BOTTOM)
	x := cx - total/2

	for i := range bar.Items {
		bar.Items[i].Rect.Left = x
		bar.Items[i].Rect.Right = x + bar.Items[i].Width
		bar.Items[i].Rect.Top = top
		bar.Items[i].Rect.Bottom = top + s.Scale(MODE_BAR_H)
		x += bar.Items[i].Width + s.Scale(MODE_BAR_GAP)
	}
}

func drawModeBar(hdc win.HDC, s *AppState, clockRect *win.RECT) {
	var bar ModeBar
	buildModeBar(s, clockRect, &bar)

	for _, it := range bar.Items {
		if it.IsLabel {
			win.SetBkMode(hdc, win.TRANSPARENT)
			win.SetTextColor(hdc, it.Fg)
			hOld := win.SelectObject(hdc, win.HGDIOBJ(s.Fonts.HGuiFont))
			str, _ := syscall.UTF16FromString(it.Text)
			DrawText(hdc, &str[0], -1, &it.Rect, win.DT_CENTER|win.DT_VCENTER|win.DT_SINGLELINE)
			win.SelectObject(hdc, hOld)
		} else if it.Highlight {
			drawHighlightedButton(hdc, &it.Rect, it.Text, it.Bg, it.Fg, btnState(s, HT_MODE, it.Id), s.Fonts.HGuiFont)
		} else {
			drawButton(hdc, &it.Rect, it.Text, it.Bg, it.Fg, btnState(s, HT_MODE, it.Id), s.Fonts.HGuiFont)
		}
	}
}

func drawAlarmPanel(hdc win.HDC, hwnd win.HWND, s *AppState, clockRect *win.RECT) {
	var panel, header win.RECT
	calcAlarmRects(s, hwnd, &panel, &header, clockRect)

	hOldBr := win.SelectObject(hdc, win.HGDIOBJ(s.Colors.HPanelBrush))
	hOldPn := win.SelectObject(hdc, win.HGDIOBJ(win.GetStockObject(win.NULL_PEN)))
	Rectangle(hdc, panel.Left, panel.Top, panel.Right, panel.Bottom)

	win.SetBkMode(hdc, win.TRANSPARENT)
	win.SetTextColor(hdc, s.Colors.TextColor)
	hOldFont := win.SelectObject(hdc, win.HGDIOBJ(s.Fonts.HGuiFont))

	settingsR := getSettingsRect(s, &header)
	hdr := header
	hdr.Right = settingsR.Left - s.Scale(8)
	str, _ := syscall.UTF16FromString("Alarms")
	DrawText(hdc, &str[0], -1, &hdr, win.DT_LEFT|win.DT_VCENTER|win.DT_SINGLELINE)
	win.SelectObject(hdc, hOldFont)

	btnBg := RGB(0xE0, 0xE0, 0xE0)
	if s.Settings.DarkMode {
		btnBg = RGB(0x45, 0x45, 0x45)
	}
	drawButton(hdc, &settingsR, "Settings", btnBg, s.Colors.TextColor, btnState(s, HT_SETTINGS, 0), s.Fonts.HGuiFont)

	win.SetBkMode(hdc, win.TRANSPARENT)
	win.SetTextColor(hdc, s.Colors.TextColor)
	win.SelectObject(hdc, win.HGDIOBJ(s.Fonts.HGuiFont))

	arrow := "\u25BC"
	if s.AlarmsCollapsed {
		arrow = "\u25B6"
	}
	hdr = header
	hdr.Left = hdr.Right - s.Scale(22)
	str, _ = syscall.UTF16FromString(arrow)
	DrawText(hdc, &str[0], 1, &hdr, win.DT_CENTER|win.DT_VCENTER|win.DT_SINGLELINE)
	win.SelectObject(hdc, hOldFont)

	if s.AlarmsCollapsed {
		win.SelectObject(hdc, hOldBr)
		win.SelectObject(hdc, hOldPn)
		return
	}

	hChkPen := CreatePen(win.PS_SOLID, 2, s.Colors.TextColor)
	hChkOn := CreateSolidBrush(s.Colors.AccentColor)
	hChkOff := CreateSolidBrush(s.Colors.BgColor)

	faceName, _ := syscall.UTF16FromString("Segoe UI")
	var lf win.LOGFONT
	lf.LfHeight = s.Scale(14)
	lf.LfWeight = win.FW_BOLD
	lf.LfQuality = win.CLEARTYPE_QUALITY
	lf.LfPitchAndFamily = win.DEFAULT_PITCH | win.FF_DONTCARE
	copy(lf.LfFaceName[:], faceName)
	hTickFont := win.CreateFontIndirect(&lf)

	visible := alarmVisibleRows(s, &panel, &header)
	for i := 0; i < visible; i++ {
		rowR := getAlarmRowRect(s, &panel, &header, i)
		chkR := getCheckRect(s, &rowR)
		editR := getEditRect(s, &rowR)
		clrR := getClearRect(s, &rowR)

		armed := s.Settings.Alarms[i].Enabled && s.Settings.Alarms[i].Hour != -1

		if armed {
			win.SelectObject(hdc, win.HGDIOBJ(hChkOn))
		} else {
			win.SelectObject(hdc, win.HGDIOBJ(hChkOff))
		}
		win.SelectObject(hdc, win.HGDIOBJ(hChkPen))
		stChk := btnState(s, HT_ALARM_CHECK, i)
		if stChk == BTN_HOVER {
			// Subtle highlight
			win.SelectObject(hdc, win.HGDIOBJ(CreateSolidBrush(s.Colors.AccentColor)))
		}

		RoundRect(hdc, chkR.Left, chkR.Top, chkR.Right, chkR.Bottom, 3, 3)

		if armed {
			win.SetBkMode(hdc, win.TRANSPARENT)
			win.SetTextColor(hdc, RGB(255, 255, 255))
			win.SelectObject(hdc, win.HGDIOBJ(hTickFont))
			tick, _ := syscall.UTF16FromString("\u2713")
			DrawText(hdc, &tick[0], -1, &chkR, win.DT_CENTER|win.DT_VCENTER|win.DT_SINGLELINE)
		}

		win.SetBkMode(hdc, win.TRANSPARENT)
		if armed {
			win.SetTextColor(hdc, s.Colors.TextColor)
		} else {
			win.SetTextColor(hdc, RGB(128, 128, 128)) // dim
		}
		win.SelectObject(hdc, win.HGDIOBJ(s.Fonts.HGuiFont))

		textR := rowR
		textR.Left = chkR.Right + s.Scale(12)
		textR.Right = editR.Left - s.Scale(8)

		var timeStr string
		if s.Settings.Alarms[i].Hour == -1 {
			timeStr = "Not Set"
		} else {
			h := s.Settings.Alarms[i].Hour
			m := s.Settings.Alarms[i].Minute
			if !s.Settings.Hour24 {
				ampm := "AM"
				if h >= 12 {
					ampm = "PM"
				}
				if h > 12 {
					h -= 12
				}
				if h == 0 {
					h = 12
				}
				timeStr = fmt.Sprintf("%d:%02d %s", h, m, ampm)
			} else {
				timeStr = fmt.Sprintf("%02d:%02d", h, m)
			}
		}

		if s.Settings.Alarms[i].Label != "" {
			timeStr += "  -  " + s.Settings.Alarms[i].Label
		}
		if s.Settings.Alarms[i].SkipNext {
			timeStr += " (Skipped)"
		}

		tStr, _ := syscall.UTF16FromString(timeStr)
		DrawText(hdc, &tStr[0], -1, &textR, win.DT_LEFT|win.DT_VCENTER|win.DT_SINGLELINE)

		drawButton(hdc, &editR, "Edit", btnBg, s.Colors.TextColor, btnState(s, HT_ALARM_EDIT, i), s.Fonts.HGuiFont)
		drawButton(hdc, &clrR, "Clear", btnBg, s.Colors.TextColor, btnState(s, HT_ALARM_CLEAR, i), s.Fonts.HGuiFont)
	}

	win.DeleteObject(win.HGDIOBJ(hChkPen))
	win.DeleteObject(win.HGDIOBJ(hChkOn))
	win.DeleteObject(win.HGDIOBJ(hChkOff))
	win.DeleteObject(win.HGDIOBJ(hTickFont))

	win.SelectObject(hdc, hOldBr)
	win.SelectObject(hdc, hOldPn)
}

func HitTest(s *AppState, hwnd win.HWND, mx, my int32) HitTarget {
	pt := win.POINT{X: mx, Y: my}
	var out HitTarget
	out.Kind = HT_NONE

	var clockRect win.RECT
	// Basic calculation for now (assumes top portion of window)
	var cr win.RECT
	win.GetClientRect(hwnd, &cr)
	
	// Duplicate logic from clock calc to find clock bottom
	dpi := s.Dpi
	if dpi == 0 {
		dpi = 96
	}
	h := (280 * int32(dpi)) / 96
	if s.Settings.ClockStyle == settings.ClockAnalog {
		h = (560 * int32(dpi)) / 96
	}
	clockRect.Left = cr.Left
	clockRect.Top = cr.Top
	clockRect.Right = cr.Right
	clockRect.Bottom = cr.Top + h - s.Scale(70) // roughly bottom of clock

	var bar ModeBar
	buildModeBar(s, &clockRect, &bar)
	for _, it := range bar.Items {
		if !it.IsLabel && PtInRect(&it.Rect, pt) {
			out.Kind = HT_MODE
			out.Index = it.Id
			out.Rect = it.Rect
			return out
		}
	}

	var panel, header win.RECT
	calcAlarmRects(s, hwnd, &panel, &header, &clockRect)

	if PtInRect(&header, pt) {
		settingsR := getSettingsRect(s, &header)
		if PtInRect(&settingsR, pt) {
			out.Kind = HT_SETTINGS
			out.Rect = settingsR
			return out
		}
		var collR win.RECT
		collR.Right = header.Right
		collR.Left = header.Right - s.Scale(40)
		collR.Top = header.Top
		collR.Bottom = header.Bottom
		if PtInRect(&collR, pt) {
			out.Kind = HT_COLLAPSE
			out.Rect = collR
			return out
		}
	}

	if !s.AlarmsCollapsed && PtInRect(&panel, pt) {
		visible := alarmVisibleRows(s, &panel, &header)
		for i := 0; i < visible; i++ {
			rowR := getAlarmRowRect(s, &panel, &header, i)
			if !PtInRect(&rowR, pt) {
				continue
			}

			chkR := getCheckRect(s, &rowR)
			if PtInRect(&chkR, pt) {
				out.Kind = HT_ALARM_CHECK
				out.Index = i
				out.Rect = chkR
				return out
			}
			editR := getEditRect(s, &rowR)
			if PtInRect(&editR, pt) {
				out.Kind = HT_ALARM_EDIT
				out.Index = i
				out.Rect = editR
				return out
			}
			clrR := getClearRect(s, &rowR)
			if PtInRect(&clrR, pt) {
				out.Kind = HT_ALARM_CLEAR
				out.Index = i
				out.Rect = clrR
				return out
			}
		}
	}

	return out
}

func DrawPanel(hdc win.HDC, hwnd win.HWND, s *AppState) {
	// Re-calculate clock rect exactly as in WM_PAINT to ensure mode bar is under it.
	var cr win.RECT
	win.GetClientRect(hwnd, &cr)
	
	dpi := s.Dpi
	if dpi == 0 { dpi = 96 }
	
	analog := s.Settings.ClockStyle == settings.ClockAnalog
	baseW, baseH := int32(500), int32(280)
	if analog {
		baseH = 560
	}
	
	w := (baseW * int32(dpi)) / 96
	h := (baseH * int32(dpi)) / 96
	
	var clockRect win.RECT
	clockRect.Left = (cr.Right - cr.Left - w) / 2
	clockRect.Right = clockRect.Left + w
	clockRect.Top = 0
	clockRect.Bottom = h
	if !analog {
		clockRect.Bottom -= s.Scale(70) // adjustment for digital clock height in C
	}

	drawModeBar(hdc, s, &clockRect)
	drawAlarmPanel(hdc, hwnd, s, &clockRect)
}
