package ui

import (
	"strconv"
	"syscall"
	"unsafe"
	"github.com/lxn/win"
	"github.com/n-kisyov/glock/internal/settings"
)

var (
	snoozeItems  = []string{"1", "2", "3", "5", "10", "15", "20", "30"}
	snoozeValues = []int{1, 2, 3, 5, 10, 15, 20, 30}

	volItems  = []string{"10%", "20%", "30%", "40%", "50%", "60%", "70%", "80%", "90%", "100%"}
	volValues = []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	sleepItems  = []string{"15", "30", "45", "60", "90"}
	sleepValues = []int{15, 30, 45, 60, 90}
)

func getDlgState(hDlg win.HWND) *AppState {
	return (*AppState)(unsafe.Pointer(uintptr(win.GetWindowLongPtr(hDlg, win.GWLP_USERDATA))))
}

func checkDlgButton(hDlg win.HWND, id int32, check bool) {
	state := win.BST_UNCHECKED
	if check {
		state = win.BST_CHECKED
	}
	CheckDlgButton(hDlg, id, uint32(state))
}

func isDlgButtonChecked(hDlg win.HWND, id int32) bool {
	return IsDlgButtonChecked(hDlg, id) == win.BST_CHECKED
}

func SettingsDlgProc(hDlg win.HWND, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	s := getDlgState(hDlg)

	switch msg {
	case win.WM_INITDIALOG:
		s = (*AppState)(unsafe.Pointer(lParam))
		win.SetWindowLongPtr(hDlg, win.GWLP_USERDATA, lParam)

		ThemeDialogInit(hDlg, s)

		checkDlgButton(hDlg, IDC_DARKMODE, s.Settings.DarkMode)
		checkDlgButton(hDlg, IDC_ALARMS_ENABLED, s.Settings.AlarmsEnabled)
		checkDlgButton(hDlg, IDC_HOUR24, s.Settings.Hour24)
		checkDlgButton(hDlg, IDC_CRESCENDO, s.Settings.Crescendo)
		checkDlgButton(hDlg, IDC_AUTOSTART, s.Settings.AutoStart)
		checkDlgButton(hDlg, IDC_START_MINIMIZED, s.Settings.StartMinimized)
		checkDlgButton(hDlg, IDC_ACRYLIC, s.Settings.Acrylic)
		checkDlgButton(hDlg, IDC_ALWAYS_ON_TOP, s.Settings.AlwaysOnTop)

		if s.Settings.ClockStyle == "analog" {
			checkDlgButton(hDlg, IDC_CLOCK_ANALOG, true)
		} else {
			checkDlgButton(hDlg, IDC_CLOCK_DIGITAL, true)
		}

		if s.Settings.SoundMode == "mp3" {
			checkDlgButton(hDlg, IDC_SOUND_MP3, true)
		} else {
			checkDlgButton(hDlg, IDC_SOUND_SIMPLE, true)
		}

		hCombo := win.GetDlgItem(hDlg, IDC_ALARM_COUNT)
		for i := 1; i <= settings.MaxAlarms; i++ {
			str, _ := syscall.UTF16FromString(strconv.Itoa(i))
			win.SendMessage(hCombo, win.CB_ADDSTRING, 0, uintptr(unsafe.Pointer(&str[0])))
		}
		win.SendMessage(hCombo, win.CB_SETCURSEL, uintptr(s.Settings.AlarmCount-1), 0)

		hCombo = win.GetDlgItem(hDlg, IDC_SNOOZE_MINUTES)
		sel := 0
		for i, val := range snoozeValues {
			str, _ := syscall.UTF16FromString(snoozeItems[i])
			win.SendMessage(hCombo, win.CB_ADDSTRING, 0, uintptr(unsafe.Pointer(&str[0])))
			if val == s.Settings.SnoozeMinutes {
				sel = i
			}
		}
		win.SendMessage(hCombo, win.CB_SETCURSEL, uintptr(sel), 0)

		hCombo = win.GetDlgItem(hDlg, IDC_ALARM_VOLUME)
		sel = 0
		for i, val := range volValues {
			str, _ := syscall.UTF16FromString(volItems[i])
			win.SendMessage(hCombo, win.CB_ADDSTRING, 0, uintptr(unsafe.Pointer(&str[0])))
			if val == s.Settings.AlarmVolume {
				sel = i
			}
		}
		win.SendMessage(hCombo, win.CB_SETCURSEL, uintptr(sel), 0)

		SetDlgItemText(hDlg, IDC_WAKE_NOTE, "")

		hCombo = win.GetDlgItem(hDlg, IDC_SLEEP_MINUTES)
		sel = 1 // 30 minutes default
		for i, val := range sleepValues {
			str, _ := syscall.UTF16FromString(sleepItems[i])
			win.SendMessage(hCombo, win.CB_ADDSTRING, 0, uintptr(unsafe.Pointer(&str[0])))
			if val == s.Settings.SleepMinutes {
				sel = i
			}
		}
		win.SendMessage(hCombo, win.CB_SETCURSEL, uintptr(sel), 0)

		return 1 // TRUE

	case win.WM_CTLCOLORSTATIC, win.WM_CTLCOLORBTN:
		if s == nil {
			break
		}
		return uintptr(ThemeDialogColors(hDlg, s, win.HWND(lParam), win.HDC(wParam)))

	case win.WM_CTLCOLOREDIT:
		if s == nil {
			break
		}
		win.SetTextColor(win.HDC(wParam), s.Colors.TextColor)
		win.SetBkColor(win.HDC(wParam), s.Colors.PanelBgColor)
		return uintptr(s.Colors.HPanelBrush)

	case win.WM_CTLCOLORDLG:
		if s == nil {
			break
		}
		return uintptr(s.Colors.HBgBrush)

	case win.WM_MEASUREITEM:
		mis := (*win.MEASUREITEMSTRUCT)(unsafe.Pointer(lParam))
		if mis.CtlType == ODT_COMBOBOX {
			mis.ItemHeight = 20
			return 1
		}

	case win.WM_DRAWITEM:
		dis := (*win.DRAWITEMSTRUCT)(unsafe.Pointer(lParam))
		if ThemeDrawComboItem(s, dis) {
			return 1
		}

	case win.WM_COMMAND:
		if s == nil {
			break
		}
		switch win.LOWORD(uint32(wParam)) {
		case win.IDOK:
			newDark := isDlgButtonChecked(hDlg, IDC_DARKMODE)
			darkChanged := newDark != s.Settings.DarkMode
			s.Settings.DarkMode = newDark

			s.Settings.Hour24 = isDlgButtonChecked(hDlg, IDC_HOUR24)
			s.Settings.Crescendo = isDlgButtonChecked(hDlg, IDC_CRESCENDO)
			s.Settings.AutoStart = isDlgButtonChecked(hDlg, IDC_AUTOSTART)
			s.Settings.StartMinimized = isDlgButtonChecked(hDlg, IDC_START_MINIMIZED)

			acrylicChanged := isDlgButtonChecked(hDlg, IDC_ACRYLIC) != s.Settings.Acrylic
			s.Settings.Acrylic = isDlgButtonChecked(hDlg, IDC_ACRYLIC)

			topChanged := isDlgButtonChecked(hDlg, IDC_ALWAYS_ON_TOP) != s.Settings.AlwaysOnTop
			s.Settings.AlwaysOnTop = isDlgButtonChecked(hDlg, IDC_ALWAYS_ON_TOP)

			newStyle := "digital"
			if isDlgButtonChecked(hDlg, IDC_CLOCK_ANALOG) {
				newStyle = "analog"
			}
			styleChanged := newStyle != s.Settings.ClockStyle
			s.Settings.ClockStyle = newStyle

			s.Settings.AlarmsEnabled = isDlgButtonChecked(hDlg, IDC_ALARMS_ENABLED)

			s.Settings.SoundMode = "simple"
			if isDlgButtonChecked(hDlg, IDC_SOUND_MP3) {
				s.Settings.SoundMode = "mp3"
			}

			hCombo := win.GetDlgItem(hDlg, IDC_ALARM_COUNT)
			sel := int(win.SendMessage(hCombo, win.CB_GETCURSEL, 0, 0))
			if sel >= 0 {
				s.Settings.AlarmCount = sel + 1
			}

			hCombo = win.GetDlgItem(hDlg, IDC_SNOOZE_MINUTES)
			sel = int(win.SendMessage(hCombo, win.CB_GETCURSEL, 0, 0))
			if sel >= 0 && sel < len(snoozeValues) {
				s.Settings.SnoozeMinutes = snoozeValues[sel]
			}

			hCombo = win.GetDlgItem(hDlg, IDC_ALARM_VOLUME)
			sel = int(win.SendMessage(hCombo, win.CB_GETCURSEL, 0, 0))
			if sel >= 0 && sel < len(volValues) {
				s.Settings.AlarmVolume = volValues[sel]
			}

			hCombo = win.GetDlgItem(hDlg, IDC_SLEEP_MINUTES)
			sel = int(win.SendMessage(hCombo, win.CB_GETCURSEL, 0, 0))
			if sel >= 0 && sel < len(sleepValues) {
				s.Settings.SleepMinutes = sleepValues[sel]
			}

			UpdateColors(s.Settings, &s.Colors)

			if darkChanged || acrylicChanged {
				ApplyTheme(s.HMainWnd, s.Settings.DarkMode, s.Settings.Acrylic)
			}

			if topChanged {
				hwndTopmost := win.HWND(win.HWND_TOPMOST)
				if !s.Settings.AlwaysOnTop {
					hwndTopmost = win.HWND(win.HWND_NOTOPMOST)
				}
				win.SetWindowPos(s.HMainWnd, hwndTopmost, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE)
			}

			if styleChanged {
				dpi := s.Dpi
				if dpi == 0 {
					dpi = 96
				}
				w := int32(500)
				h := int32(280)
				if s.Settings.ClockStyle == "analog" {
					w = 500
					h = 560
				}
				w = (w * dpi) / 96
				h = (h * dpi) / 96
				win.SetWindowPos(s.HMainWnd, 0, 0, 0, w, h, win.SWP_NOMOVE|win.SWP_NOZORDER|win.SWP_NOACTIVATE)
			}

			settingsPath, _ := settings.GetSettingsPath(); settings.Save(settingsPath, s.Settings)
			win.EndDialog(hDlg, win.IDOK)
			return 1

		case win.IDCANCEL:
			win.EndDialog(hDlg, win.IDCANCEL)
			return 1

		case IDC_PREVIEW_SOUND:
			// Let caller handle preview via SendMessage to main window
			win.SendMessage(s.HMainWnd, win.WM_COMMAND, IDC_PREVIEW_SOUND, 0)
			return 1
		}
	}
	return 0
}
