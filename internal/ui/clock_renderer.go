package ui

import (
	"fmt"
	"math"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/win"
	"golang.org/x/sys/windows"
	"github.com/n-kisyov/gclock/internal/settings"
)

type ClockFonts struct {
	HClockFont win.HFONT
	HDateFont  win.HFONT
	HGuiFont   win.HFONT
}

var (
	gAmpmFont win.HFONT
	gAmpmH    int
)

func EnsureAmpmFont(h int) win.HFONT {
	if gAmpmFont != 0 && gAmpmH == h {
		return gAmpmFont
	}
	if gAmpmFont != 0 {
		win.DeleteObject(win.HGDIOBJ(gAmpmFont))
	}
	
	name, _ := syscall.UTF16FromString("Segoe UI")
	lf := win.LOGFONT{
		LfHeight:         int32(h),
		LfWeight:         win.FW_NORMAL,
		LfQuality:        win.CLEARTYPE_QUALITY,
		LfPitchAndFamily: win.DEFAULT_PITCH | win.FF_DONTCARE,
	}
	copy(lf.LfFaceName[:], name)
	
	gAmpmFont = win.CreateFontIndirect(&lf)
	gAmpmH = h
	if gAmpmFont == 0 {
		gAmpmH = 0
	}
	return gAmpmFont
}

func ClockCleanup() {
	if gAmpmFont != 0 {
		win.DeleteObject(win.HGDIOBJ(gAmpmFont))
		gAmpmFont = 0
		gAmpmH = 0
	}
}

func fitTextFont(hdc win.HDC, s *settings.Settings, text string, maxW int32, sz *win.SIZE, baseFont win.HFONT) win.HFONT {
	prev := win.SelectObject(hdc, win.HGDIOBJ(baseFont))
	
	strPtr, _ := syscall.UTF16FromString(text)
	win.GetTextExtentPoint32(hdc, &strPtr[0], int32(len(strPtr)-1), sz)
	
	if sz.CX <= maxW || sz.CX <= 0 || maxW <= 0 {
		win.SelectObject(hdc, prev)
		return 0
	}
	
	var lf win.LOGFONT
	win.GetObject(win.HGDIOBJ(baseFont), uintptr(unsafe.Sizeof(lf)), unsafe.Pointer(&lf))
	
	h := lf.LfHeight
	if h < 0 { h = -h }
	
	scaled := int32(int64(h) * int64(maxW) / int64(sz.CX))
	if scaled < 10 { scaled = 10 }
	
	if lf.LfHeight < 0 {
		lf.LfHeight = -scaled
	} else {
		lf.LfHeight = scaled
	}
	
	fitted := win.CreateFontIndirect(&lf)
	if fitted == 0 {
		win.SelectObject(hdc, prev)
		return 0
	}
	
	win.SelectObject(hdc, win.HGDIOBJ(fitted))
	win.GetTextExtentPoint32(hdc, &strPtr[0], int32(len(strPtr)-1), sz)
	win.SelectObject(hdc, prev)
	return fitted
}

func ClockDrawDigital(hdc win.HDC, rc *win.RECT, st *windows.Systemtime, s *settings.Settings, fonts *ClockFonts, alarmActive bool, clockColor, textColor win.COLORREF) {
	h := st.Hour
	ampm := ""
	if !s.Hour24 {
		if h >= 12 {
			ampm = "PM"
		} else {
			ampm = "AM"
		}
		if h == 0 {
			h = 12
		} else if h > 12 {
			h -= 12
		}
	}
	
	timeStr := fmt.Sprintf("%02d:%02d:%02d", h, st.Minute, st.Second)
	dateStr := fmt.Sprintf("%04d-%02d-%02d", st.Year, st.Month, st.Day)
	
	// Try to get long date format
	var dateBuf [128]uint16
	ret := GetDateFormatEx(nil, 0x00000002, st, nil, &dateBuf[0], int32(len(dateBuf)))
	if ret != 0 {
		dateStr = syscall.UTF16ToString(dateBuf[:])
	}
	
	win.SetBkMode(hdc, win.TRANSPARENT)
	tc := clockColor
	if alarmActive {
		tc = RGB(0xFF, 0x40, 0x40)
	}
	
	hOldFont := win.SelectObject(hdc, win.HGDIOBJ(fonts.HClockFont))
	var tmClock win.TEXTMETRIC
	win.GetTextMetrics(hdc, &tmClock)
	
	win.SelectObject(hdc, win.HGDIOBJ(fonts.HDateFont))
	var tmDate win.TEXTMETRIC
	win.GetTextMetrics(hdc, &tmDate)
	
	gap := tmClock.TmHeight / 10
	startY := rc.Top + 2
	timeBaseline := startY + tmClock.TmAscent
	
	win.SelectObject(hdc, win.HGDIOBJ(fonts.HClockFont))
	var timeSize win.SIZE
	timeUTF16, _ := syscall.UTF16FromString(timeStr)
	win.GetTextExtentPoint32(hdc, &timeUTF16[0], int32(len(timeUTF16)-1), &timeSize)
	
	ampmW := int32(0)
	ampmH := int32(0)
	var hAmpmFont win.HFONT
	var tmAm win.TEXTMETRIC
	
	if ampm != "" {
		ampmH = tmClock.TmHeight / 5
		if ampmH < 12 { ampmH = 12 }
		hAmpmFont = EnsureAmpmFont(int(ampmH))
		if hAmpmFont != 0 {
			hPrev := win.SelectObject(hdc, win.HGDIOBJ(hAmpmFont))
			var amSize win.SIZE
			amUTF16, _ := syscall.UTF16FromString(ampm)
			win.GetTextExtentPoint32(hdc, &amUTF16[0], 2, &amSize)
			ampmW = amSize.CX
			win.GetTextMetrics(hdc, &tmAm)
			win.SelectObject(hdc, hPrev)
		}
	}
	
	totalW := timeSize.CX
	if ampm != "" {
		totalW += 6 + ampmW
	}
	cx := (rc.Left + rc.Right) / 2
	blockX := cx - totalW / 2
	
	win.SetTextColor(hdc, tc)
	win.SelectObject(hdc, win.HGDIOBJ(fonts.HClockFont))
	ExtTextOut(hdc, blockX, startY, 0, nil, &timeUTF16[0], int32(len(timeUTF16)-1), nil)
	
	if ampm != "" && hAmpmFont != 0 {
		hOldAmpm := win.SelectObject(hdc, win.HGDIOBJ(hAmpmFont))
		win.SetTextColor(hdc, tc)
		amUTF16, _ := syscall.UTF16FromString(ampm)
		ExtTextOut(hdc, blockX + timeSize.CX + 6, timeBaseline - tmAm.TmAscent, 0, nil, &amUTF16[0], 2, nil)
		win.SelectObject(hdc, hOldAmpm)
	}
	
	win.SelectObject(hdc, win.HGDIOBJ(fonts.HDateFont))
	win.SetTextColor(hdc, textColor)
	tr := *rc
	tr.Top = startY + tmClock.TmHeight + gap
	
	dateUTF16, _ := syscall.UTF16FromString(dateStr)
	DrawText(hdc, &dateUTF16[0], -1, &tr, win.DT_CENTER | win.DT_SINGLELINE | win.DT_TOP)
	
	win.SelectObject(hdc, hOldFont)
}

func ClockDrawAnalog(hdc win.HDC, rc *win.RECT, s *settings.Settings, textColor, accentColor win.COLORREF) {
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top
	cx := float64(rc.Left + w/2)
	cy := float64(rc.Top + h/2)
	
	minDim := w
	if h < w { minDim = h }
	radius := float64(minDim/2) - 8.0
	if radius < 30.0 { radius = 30.0 }
	
	now := time.Now()
	msFrac := float64(now.Nanosecond()) / 1e9
	secFrac := float64(now.Second()) + msFrac
	minFrac := float64(now.Minute()) + secFrac/60.0
	hourFrac := float64(now.Hour()%12) + minFrac/60.0
	
	hourAngle := hourFrac * (math.Pi / 6.0) - math.Pi/2.0
	minuteAngle := minFrac * (math.Pi / 30.0) - math.Pi/2.0
	secondAngle := secFrac * (math.Pi / 30.0) - math.Pi/2.0
	
	// Create GDI Objects
	var faceColor, rimColor, tickColor, minColor, secColor win.COLORREF
	if s.DarkMode {
		faceColor = RGB(0x26, 0x1F, 0x1B) // 0xFF1B1F26 BGR vs RGB? GDI is COLORREF (0x00bbggrr)
		faceColor = RGB(0x1B, 0x1F, 0x26)
		rimColor  = RGB(0x5A, 0x60, 0x70)
		tickColor = RGB(0xD2, 0xD7, 0xE0)
		minColor  = RGB(0x7F, 0xC0, 0xEA)
		secColor  = RGB(0xFF, 0x6B, 0x6B)
	} else {
		faceColor = RGB(0xFA, 0xFA, 0xFF)
		rimColor  = RGB(0xA0, 0xA0, 0xA8)
		tickColor = textColor
		minColor  = RGB(0x5B, 0xA0, 0xD0)
		secColor  = RGB(0xFF, 0x40, 0x40)
	}
	
	faceBrush := CreateSolidBrush(faceColor)
	rimPen := CreatePen(PS_SOLID, 2, rimColor)
	
	win.SelectObject(hdc, win.HGDIOBJ(faceBrush))
	win.SelectObject(hdc, win.HGDIOBJ(rimPen))
	win.Ellipse(hdc, int32(cx-radius), int32(cy-radius), int32(cx+radius), int32(cy+radius))
	
	win.DeleteObject(win.HGDIOBJ(faceBrush))
	win.DeleteObject(win.HGDIOBJ(rimPen))
	
	majorTick := CreatePen(PS_SOLID, 3, tickColor)
	minorTick := CreatePen(PS_SOLID, 1, tickColor)
	
	for i := 0; i < 60; i++ {
		a := float64(i) * math.Pi / 30.0 - math.Pi/2.0
		ca := math.Cos(a)
		sa := math.Sin(a)
		
		isMajor := (i % 5 == 0)
		inner := radius - 10
		if isMajor {
			inner = radius - 18
		}
		
		if isMajor {
			win.SelectObject(hdc, win.HGDIOBJ(majorTick))
		} else {
			win.SelectObject(hdc, win.HGDIOBJ(minorTick))
		}
		
		win.MoveToEx(hdc, int(cx+(radius-4)*ca), int(cy+(radius-4)*sa), nil)
		win.LineTo(hdc, int32(cx+inner*ca), int32(cy+inner*sa))
	}
	win.DeleteObject(win.HGDIOBJ(majorTick))
	win.DeleteObject(win.HGDIOBJ(minorTick))
	
	// Numerals (using GDI text)
	numH := int32(radius / 8)
	if numH < 12 { numH = 12 }
	
	numFont := EnsureAmpmFont(int(numH))
	if numFont != 0 {
		oldFont := win.SelectObject(hdc, win.HGDIOBJ(numFont))
		win.SetBkMode(hdc, win.TRANSPARENT)
		
		numColor := tickColor
		if s.DarkMode { numColor = RGB(0xF2, 0xE6, 0xCC) }
		win.SetTextColor(hdc, numColor)
		
		for i := 1; i <= 12; i++ {
			a := float64(i) * math.Pi / 6.0 - math.Pi/2.0
			str := fmt.Sprintf("%d", i)
			strPtr, _ := syscall.UTF16FromString(str)
			
			var sz win.SIZE
			win.GetTextExtentPoint32(hdc, &strPtr[0], int32(len(strPtr)-1), &sz)
			
			x := int32(cx + (radius-30)*math.Cos(a)) - sz.CX/2
			y := int32(cy + (radius-30)*math.Sin(a)) - sz.CY/2
			ExtTextOut(hdc, x, y, 0, nil, &strPtr[0], int32(len(strPtr)-1), nil)
		}
		
		if !s.Hour24 {
			ap := "AM"
			if now.Hour() >= 12 { ap = "PM" }
			apPtr, _ := syscall.UTF16FromString(ap)
			
			var sz win.SIZE
			win.GetTextExtentPoint32(hdc, &apPtr[0], 2, &sz)
			x := int32(cx) - sz.CX/2
			y := int32(cy + radius*0.34)
			win.SetTextColor(hdc, tickColor)
			ExtTextOut(hdc, x, y, 0, nil, &apPtr[0], 2, nil)
		}
		win.SelectObject(hdc, oldFont)
	}
	
	hLen := radius * 0.50
	mLen := radius * 0.75
	sLen := radius * 0.82
	
	hourPen := CreatePen(PS_SOLID, 5, accentColor)
	win.SelectObject(hdc, win.HGDIOBJ(hourPen))
	win.MoveToEx(hdc, int(cx), int(cy), nil)
	win.LineTo(hdc, int32(cx+hLen*math.Cos(hourAngle)), int32(cy+hLen*math.Sin(hourAngle)))
	win.DeleteObject(win.HGDIOBJ(hourPen))
	
	minPen := CreatePen(PS_SOLID, 3, minColor)
	win.SelectObject(hdc, win.HGDIOBJ(minPen))
	win.MoveToEx(hdc, int(cx), int(cy), nil)
	win.LineTo(hdc, int32(cx+mLen*math.Cos(minuteAngle)), int32(cy+mLen*math.Sin(minuteAngle)))
	win.DeleteObject(win.HGDIOBJ(minPen))
	
	secPen := CreatePen(PS_SOLID, 1, secColor)
	win.SelectObject(hdc, win.HGDIOBJ(secPen))
	win.MoveToEx(hdc, int(cx), int(cy), nil)
	win.LineTo(hdc, int32(cx+sLen*math.Cos(secondAngle)), int32(cy+sLen*math.Sin(secondAngle)))
	win.DeleteObject(win.HGDIOBJ(secPen))
	
	dotBrush := CreateSolidBrush(accentColor)
	win.SelectObject(hdc, win.HGDIOBJ(dotBrush))
	win.SelectObject(hdc, win.HGDIOBJ(win.GetStockObject(win.NULL_PEN)))
	win.Ellipse(hdc, int32(cx-4), int32(cy-4), int32(cx+4), int32(cy+4))
	win.DeleteObject(win.HGDIOBJ(dotBrush))
}
