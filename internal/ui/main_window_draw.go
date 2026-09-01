package ui

import (
	"syscall"
	"github.com/lxn/win"
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

func (s *AppState) Scale(v int32) int32 {
	dpi := s.Dpi
	if dpi == 0 { dpi = 96 }
	return (v * dpi) / 96
}

func shade(c win.COLORREF, delta int32) win.COLORREF {
	r := int32(c & 0xFF) + delta
	g := int32((c >> 8) & 0xFF) + delta
	b := int32((c >> 16) & 0xFF) + delta
	
	if r < 0 { r = 0 }
	if r > 255 { r = 255 }
	if g < 0 { g = 0 }
	if g > 255 { g = 255 }
	if b < 0 { b = 0 }
	if b > 255 { b = 255 }
	return RGB(byte(r), byte(g), byte(b))
}

func drawButton(hdc win.HDC, r *win.RECT, text string, bg, fg win.COLORREF, st BtnState, hFont win.HFONT) {
	if st == BTN_HOVER {
		bg = shade(bg, 22)
	} else if st == BTN_PRESSED {
		bg = shade(bg, -28)
	}
	
	hBr := CreateSolidBrush(bg)
	hPn := CreatePen(PS_SOLID, 1, fg)
	hOldBr := win.SelectObject(hdc, win.HGDIOBJ(hBr))
	hOldPn := win.SelectObject(hdc, win.HGDIOBJ(hPn))
	
	// win.RoundRect doesn't exist in lxn/win by default, I should use it from gdi32
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
