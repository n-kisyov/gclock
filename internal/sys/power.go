package sys

import (
	"golang.org/x/sys/windows"
	"unsafe"
)

var (
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	powrprof = windows.NewLazySystemDLL("powrprof.dll")

	procCreateWaitableTimerExW = kernel32.NewProc("CreateWaitableTimerExW")
	procCreateWaitableTimerW   = kernel32.NewProc("CreateWaitableTimerW")
	procSetWaitableTimer       = kernel32.NewProc("SetWaitableTimer")
	procCancelWaitableTimer    = kernel32.NewProc("CancelWaitableTimer")
	procSetThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")

	procPowerGetActiveScheme  = powrprof.NewProc("PowerGetActiveScheme")
	procPowerReadACValueIndex = powrprof.NewProc("PowerReadACValueIndex")
)

const (
	CREATE_WAITABLE_TIMER_MANUAL_RESET      = 0x00000001
	CREATE_WAITABLE_TIMER_HIGH_RESOLUTION   = 0x00000002
	TIMER_ALL_ACCESS                        = 0x1F0003
	
	ES_SYSTEM_REQUIRED  = 0x00000001
	ES_DISPLAY_REQUIRED = 0x00000002
	ES_CONTINUOUS       = 0x80000000
)

var (
	kSleepSubgroup = windows.GUID{Data1: 0x238c9fa8, Data2: 0x0aad, Data3: 0x41ed, Data4: [8]byte{0x83, 0xf4, 0x97, 0xbe, 0x24, 0x2c, 0x8f, 0x20}}
	kAllowRtcWake  = windows.GUID{Data1: 0xbd3b718a, Data2: 0x0680, Data3: 0x4d9d, Data4: [8]byte{0x8a, 0xb2, 0xe1, 0xd2, 0xb4, 0xac, 0x80, 0x6d}}
)

var gWakeTimer windows.Handle

func ensureWakeTimer() windows.Handle {
	if gWakeTimer != 0 {
		return gWakeTimer
	}
	
	ret, _, _ := procCreateWaitableTimerExW.Call(
		0,
		0,
		uintptr(CREATE_WAITABLE_TIMER_MANUAL_RESET|CREATE_WAITABLE_TIMER_HIGH_RESOLUTION),
		uintptr(TIMER_ALL_ACCESS),
	)
	if ret == 0 {
		ret, _, _ = procCreateWaitableTimerW.Call(0, 1, 0) // TRUE for manual reset
	}
	gWakeTimer = windows.Handle(ret)
	return gWakeTimer
}

// ArmWakeTimer arms a timer to wake the PC in the specified number of seconds.
// Passes fResume=TRUE to SetWaitableTimer.
func ArmWakeTimer(seconds int64) bool {
	h := ensureWakeTimer()
	if h == 0 {
		return false
	}
	
	if seconds < 0 {
		CancelWakeTimer()
		return false
	}
	
	// negative is relative to now
	li := -int64(seconds * 10000000)
	
	ret, _, _ := procSetWaitableTimer.Call(
		uintptr(h),
		uintptr(unsafe.Pointer(&li)),
		0,
		0,
		0,
		1, // fResume = TRUE
	)
	return ret != 0
}

func CancelWakeTimer() {
	if gWakeTimer != 0 {
		procCancelWaitableTimer.Call(uintptr(gWakeTimer))
	}
}

func WakeTimersAllowed() bool {
	var scheme *windows.GUID
	ret, _, _ := procPowerGetActiveScheme.Call(0, uintptr(unsafe.Pointer(&scheme)))
	if ret != 0 || scheme == nil {
		return true // assume allowed if we can't read it
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(scheme)))
	
	var value uint32 = 1
	ret2, _, _ := procPowerReadACValueIndex.Call(
		0,
		uintptr(unsafe.Pointer(scheme)),
		uintptr(unsafe.Pointer(&kSleepSubgroup)),
		uintptr(unsafe.Pointer(&kAllowRtcWake)),
		uintptr(unsafe.Pointer(&value)),
	)
	if ret2 != 0 {
		value = 1
	}
	return value != 0
}

func KeepAwake(awake bool) {
	flags := uintptr(ES_CONTINUOUS)
	if awake {
		flags |= ES_SYSTEM_REQUIRED | ES_DISPLAY_REQUIRED
	}
	procSetThreadExecutionState.Call(flags)
}

func Cleanup() {
	KeepAwake(false)
	if gWakeTimer != 0 {
		CancelWakeTimer()
		windows.CloseHandle(gWakeTimer)
		gWakeTimer = 0
	}
}
