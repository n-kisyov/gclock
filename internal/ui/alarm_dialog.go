package ui

import (
	"strconv"
	"syscall"
	"unsafe"
	"path/filepath"

	"github.com/lxn/win"
)

var dayIds = []int32{
	IDC_DAY_SUN, IDC_DAY_MON, IDC_DAY_TUE, IDC_DAY_WED,
	IDC_DAY_THU, IDC_DAY_FRI, IDC_DAY_SAT,
}

type AlarmEditData struct {
	Hour         int
	Minute       int
	Enabled      bool
	SkipNext     bool
	RepeatDays   uint8
	Volume       int
	SnoozeMinutes int
	Label        string
	Sound        string
	State        *AppState
}

func getAlarmDlgData(hDlg win.HWND) *AlarmEditData {
	return (*AlarmEditData)(unsafe.Pointer(uintptr(win.GetWindowLongPtr(hDlg, win.GWLP_USERDATA))))
}

func fillOverrideCombo(hDlg win.HWND, id int32, values []int, current int, suffix string) {
	h := win.GetDlgItem(hDlg, id)
	str, _ := syscall.UTF16FromString("Default")
	win.SendMessage(h, win.CB_ADDSTRING, 0, uintptr(unsafe.Pointer(&str[0])))

	sel := 0
	for i, v := range values {
		buf := strconv.Itoa(v) + suffix
		str, _ := syscall.UTF16FromString(buf)
		win.SendMessage(h, win.CB_ADDSTRING, 0, uintptr(unsafe.Pointer(&str[0])))
		if v == current {
			sel = i + 1
		}
	}
	win.SendMessage(h, win.CB_SETCURSEL, uintptr(sel), 0)
}

func readOverrideCombo(hDlg win.HWND, id int32, values []int) int {
	sel := int(win.SendDlgItemMessage(hDlg, id, win.CB_GETCURSEL, 0, 0))
	if sel <= 0 || sel > len(values) {
		return -1
	}
	return values[sel-1]
}

func getDlgItemText(hDlg win.HWND, id int32) string {
	var buf [256]uint16
	win.GetDlgItemText(hDlg, id, &buf[0], 256)
	return syscall.UTF16ToString(buf[:])
}

func AlarmDlgProc(hDlg win.HWND, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	data := getAlarmDlgData(hDlg)

	switch msg {
	case win.WM_INITDIALOG:
		data = (*AlarmEditData)(unsafe.Pointer(lParam))
		win.SetWindowLongPtr(hDlg, win.GWLP_USERDATA, lParam)

		s := data.State
		ThemeDialogInit(hDlg, s)

		if data.Label != "" {
			win.SetDlgItemText(hDlg, IDC_ALARM_LABEL, data.Label)
		}

		if data.Hour >= 0 {
			win.SetDlgItemText(hDlg, IDC_ALARM_HOUR, strconv.Itoa(data.Hour))
		} else {
			win.SetDlgItemText(hDlg, IDC_ALARM_HOUR, "")
		}

		if data.Minute >= 0 {
			win.SetDlgItemText(hDlg, IDC_ALARM_MINUTE, strconv.Itoa(data.Minute))
		} else {
			win.SetDlgItemText(hDlg, IDC_ALARM_MINUTE, "")
		}

		for i := 0; i < 7; i++ {
			check := false
			if (data.RepeatDays & (1 << i)) != 0 {
				check = true
			}
			checkDlgButton(hDlg, dayIds[i], check)
		}

		checkDlgButton(hDlg, IDC_ALARM_ENABLED, data.Enabled)
		checkDlgButton(hDlg, IDC_ALARM_SKIP, data.SkipNext)

		win.SetDlgItemText(hDlg, IDC_ALARM_SOUND, data.Sound)

		fillOverrideCombo(hDlg, IDC_ALARM_VOL, volValues, data.Volume, "%")
		fillOverrideCombo(hDlg, IDC_ALARM_SNOOZE, snoozeValues, data.SnoozeMinutes, " min")
		
		return 1

	case win.WM_MEASUREITEM:
		mis := (*win.MEASUREITEMSTRUCT)(unsafe.Pointer(lParam))
		if mis.CtlType == win.ODT_COMBOBOX {
			mis.ItemHeight = 20
			return 1
		}

	case win.WM_DRAWITEM:
		dis := (*win.DRAWITEMSTRUCT)(unsafe.Pointer(lParam))
		if data != nil && ThemeDrawComboItem(data.State, dis) {
			return 1
		}

	case win.WM_CTLCOLORSTATIC, win.WM_CTLCOLORBTN:
		if data != nil {
			return uintptr(ThemeDialogColors(hDlg, data.State, win.HWND(lParam), win.HDC(wParam)))
		}

	case win.WM_CTLCOLOREDIT:
		if data != nil {
			win.SetTextColor(win.HDC(wParam), data.State.Colors.TextColor)
			win.SetBkColor(win.HDC(wParam), data.State.Colors.PanelBgColor)
			return uintptr(data.State.Colors.HPanelBrush)
		}

	case win.WM_CTLCOLORDLG:
		if data != nil {
			return uintptr(data.State.Colors.HBgBrush)
		}

	case win.WM_COMMAND:
		switch win.LOWORD(uint32(wParam)) {
		case IDC_ALARM_HOUR:
			if win.HIWORD(uint32(wParam)) == win.EN_KILLFOCUS {
				txt := getDlgItemText(hDlg, IDC_ALARM_HOUR)
				if v, err := strconv.Atoi(txt); err == nil && v > 23 {
					win.SetDlgItemText(hDlg, IDC_ALARM_HOUR, "23")
				}
			}
		case IDC_ALARM_MINUTE:
			if win.HIWORD(uint32(wParam)) == win.EN_KILLFOCUS {
				txt := getDlgItemText(hDlg, IDC_ALARM_MINUTE)
				if v, err := strconv.Atoi(txt); err == nil && v > 59 {
					win.SetDlgItemText(hDlg, IDC_ALARM_MINUTE, "59")
				}
			}
		case IDC_DAY_ALL:
			for i := 0; i < 7; i++ {
				checkDlgButton(hDlg, dayIds[i], true)
			}
			return 1
		case IDC_DAY_NONE:
			for i := 0; i < 7; i++ {
				checkDlgButton(hDlg, dayIds[i], false)
			}
			return 1
		case IDC_ALARM_SOUND_PICK:
			if data != nil {
				// We don't have GetOpenFileName hooked up natively in lxn/win easily,
				// but let's assume we can use a basic dialog or leave it for later if it gets complex.
				// For now, let's just do nothing here to keep it simple, or implement it using win32.
				// Wait! Go-ole or win32 has it?
				// To keep it simple, let's skip the file picker in this step since it requires OPENFILENAME struct
				// which is not fully mapped in lxn/win.
				win.MessageBox(hDlg, syscall.SyscallN_String("Sound picker is not yet implemented in Go port."), syscall.SyscallN_String("Not Implemented"), win.MB_OK)
			}
			return 1
		case IDC_ALARM_SOUND_CLEAR:
			if data != nil {
				data.Sound = ""
				win.SetDlgItemText(hDlg, IDC_ALARM_SOUND, "")
			}
			return 1
		case win.IDOK:
			if data == nil {
				return 1
			}
			data.Label = getDlgItemText(hDlg, IDC_ALARM_LABEL)
			hStr := getDlgItemText(hDlg, IDC_ALARM_HOUR)
			mStr := getDlgItemText(hDlg, IDC_ALARM_MINUTE)

			if hStr == "" || mStr == "" {
				win.MessageBox(hDlg, syscall.SyscallN_String("Enter both an hour (0-23) and a minute (0-59)."),
					syscall.SyscallN_String("Invalid Time"), win.MB_OK|win.MB_ICONWARNING)
				return 1
			}

			h, _ := strconv.Atoi(hStr)
			m, _ := strconv.Atoi(mStr)

			if h < 0 || h > 23 || m < 0 || m > 59 {
				win.MessageBox(hDlg, syscall.SyscallN_String("Please enter valid hour (0-23) and minute (0-59)."),
					syscall.SyscallN_String("Invalid Time"), win.MB_OK|win.MB_ICONWARNING)
				return 1
			}

			data.Hour = h
			data.Minute = m
			data.Enabled = isDlgButtonChecked(hDlg, IDC_ALARM_ENABLED)
			data.SkipNext = isDlgButtonChecked(hDlg, IDC_ALARM_SKIP)

			data.RepeatDays = 0
			for i := 0; i < 7; i++ {
				if isDlgButtonChecked(hDlg, dayIds[i]) {
					data.RepeatDays |= (1 << i)
				}
			}

			data.Volume = readOverrideCombo(hDlg, IDC_ALARM_VOL, volValues)
			data.SnoozeMinutes = readOverrideCombo(hDlg, IDC_ALARM_SNOOZE, snoozeValues)

			win.EndDialog(hDlg, win.IDOK)
			return 1

		case win.IDCANCEL:
			win.EndDialog(hDlg, win.IDCANCEL)
			return 1
		}
	}
	return 0
}

func SyscallN_String(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}
