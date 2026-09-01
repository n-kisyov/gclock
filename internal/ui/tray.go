package ui

import (
	"syscall"
	"unsafe"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

var (
	msgTaskbarCreated uint32
	procShellNotifyIcon = windows.NewLazySystemDLL("shell32.dll").NewProc("Shell_NotifyIconW")
	procAppendMenu      = windows.NewLazySystemDLL("user32.dll").NewProc("AppendMenuW")
)

func AppendMenu(hMenu win.HMENU, uFlags uint32, uIDNewItem uintptr, lpNewItem *uint16) bool {
	ret, _, _ := procAppendMenu.Call(uintptr(hMenu), uintptr(uFlags), uIDNewItem, uintptr(unsafe.Pointer(lpNewItem)))
	return ret != 0
}

const (
	NIM_ADD    = 0x00000000
	NIM_MODIFY = 0x00000001
	NIM_DELETE = 0x00000002
	
	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004
	
	IDI_APPICON = 101 // From resource.h
	
	WM_TRAYICON = win.WM_APP + 1
	
	IDM_TRAY_SHOW = 40001
	IDM_SETTINGS  = 40002
	IDM_ABOUT     = 40003
	IDM_TRAY_EXIT = 40004
)

func shellNotifyIcon(dwMessage uint32, pnid *win.NOTIFYICONDATA) bool {
	ret, _, _ := procShellNotifyIcon.Call(uintptr(dwMessage), uintptr(unsafe.Pointer(pnid)))
	return ret != 0
}

func TaskbarCreatedMsg() uint32 {
	if msgTaskbarCreated == 0 {
		str, _ := syscall.UTF16PtrFromString("TaskbarCreated")
		msgTaskbarCreated = win.RegisterWindowMessage(str)
	}
	return msgTaskbarCreated
}

func TrayCreate(hwnd win.HWND, nid *win.NOTIFYICONDATA) bool {
	nid.CbSize = uint32(unsafe.Sizeof(*nid))
	nid.HWnd = hwnd
	nid.UID = 1
	nid.UFlags = NIF_ICON | NIF_MESSAGE | NIF_TIP
	nid.UCallbackMessage = WM_TRAYICON
	nid.HIcon = win.LoadIcon(win.GetModuleHandle(nil), win.MAKEINTRESOURCE(IDI_APPICON))
	
	tip, _ := syscall.UTF16FromString("AlarmClock")
	copy(nid.SzTip[:], tip)
	
	return shellNotifyIcon(NIM_ADD, nid)
}

func TrayRemove(nid *win.NOTIFYICONDATA) {
	shellNotifyIcon(NIM_DELETE, nid)
}

func TrayShowMenu(hwnd win.HWND) {
	win.SetForegroundWindow(hwnd)
	
	hMenu := win.CreatePopupMenu()
	
	AppendMenu(hMenu, win.MF_STRING, IDM_TRAY_SHOW, syscall.StringToUTF16Ptr("Show"))
	AppendMenu(hMenu, win.MF_SEPARATOR, 0, nil)
	AppendMenu(hMenu, win.MF_STRING, IDM_SETTINGS, syscall.StringToUTF16Ptr("Settings"))
	AppendMenu(hMenu, win.MF_STRING, IDM_ABOUT, syscall.StringToUTF16Ptr("About"))
	AppendMenu(hMenu, win.MF_SEPARATOR, 0, nil)
	AppendMenu(hMenu, win.MF_STRING, IDM_TRAY_EXIT, syscall.StringToUTF16Ptr("Exit"))
	
	var pt win.POINT
	win.GetCursorPos(&pt)
	
	win.TrackPopupMenu(hMenu, win.TPM_RIGHTALIGN | win.TPM_BOTTOMALIGN, pt.X, pt.Y, 0, hwnd, nil)
	win.DestroyMenu(hMenu)
}

func TrayUpdateTooltip(nid *win.NOTIFYICONDATA, text string) {
	tip, _ := syscall.UTF16FromString(text)
	
	// Check if unchanged
	same := true
	for i, c := range tip {
		if nid.SzTip[i] != c {
			same = false
			break
		}
	}
	if same {
		return
	}
	
	for i := range nid.SzTip {
		nid.SzTip[i] = 0
	}
	copy(nid.SzTip[:], tip)
	shellNotifyIcon(NIM_MODIFY, nid)
}
