package power

import (
	"math"
	"syscall"
	"unsafe"

	"github.com/n-kisyov/gclock/internal/alarms"
	"github.com/n-kisyov/gclock/internal/settings"
	"golang.org/x/sys/windows"
)

var (
	modkernel32               = windows.NewLazySystemDLL("kernel32.dll")
	procCreateWaitableTimerEx = modkernel32.NewProc("CreateWaitableTimerExW")
	procCreateWaitableTimer   = modkernel32.NewProc("CreateWaitableTimerW")
	procSetWaitableTimer      = modkernel32.NewProc("SetWaitableTimer")
	procCancelWaitableTimer   = modkernel32.NewProc("CancelWaitableTimer")
	procSetThreadExecutionState = modkernel32.NewProc("SetThreadExecutionState")

	modpowrprof               = windows.NewLazySystemDLL("powrprof.dll")
	procPowerGetActiveScheme  = modpowrprof.NewProc("PowerGetActiveScheme")
	procPowerReadACValueIndex = modpowrprof.NewProc("PowerReadACValueIndex")
)

const (
	CREATE_WAITABLE_TIMER_MANUAL_RESET    = 0x00000001
	CREATE_WAITABLE_TIMER_HIGH_RESOLUTION = 0x00000002
	TIMER_ALL_ACCESS                      = 0x1F0003

	ES_CONTINUOUS        = 0x80000000
	ES_SYSTEM_REQUIRED   = 0x00000001
	ES_DISPLAY_REQUIRED  = 0x00000002

	WAKE_LEAD_SECONDS = 20
)

var (
	kSleepSubgroup = windows.GUID{Data1: 0x238c9fa8, Data2: 0x0aad, Data3: 0x41ed, Data4: [8]byte{0x83, 0xf4, 0x97, 0xbe, 0x24, 0x2c, 0x8f, 0x20}}
	kAllowRtcWake  = windows.GUID{Data1: 0xbd3b718a, Data2: 0x0680, Data3: 0x4d9d, Data4: [8]byte{0x8a, 0xb2, 0xe1, 0xd2, 0xb4, 0xac, 0x80, 0x6d}}

	wakeTimer windows.Handle
)

func ensureWakeTimer() windows.Handle {
	if wakeTimer != 0 {
		return wakeTimer
	}

	ret, _, _ := procCreateWaitableTimerEx.Call(
		0,
		0,
		uintptr(CREATE_WAITABLE_TIMER_MANUAL_RESET|CREATE_WAITABLE_TIMER_HIGH_RESOLUTION),
		uintptr(TIMER_ALL_ACCESS),
	)
	if ret == 0 {
		ret, _, _ = procCreateWaitableTimer.Call(0, 1, 0)
	}
	wakeTimer = windows.Handle(ret)
	return wakeTimer
}

func SecondsUntilWake(s *settings.Settings, now *windows.Systemtime) int64 {
	if !s.AlarmsEnabled {
		return -1
	}

	best := math.MaxInt32
	for i := 0; i < settings.MaxAlarms; i++ {
		delta, ok := alarms.NextDeltaMinutes(now, &s.Alarms[i])
		if ok && delta < best {
			best = delta
		}
	}
	if best == math.MaxInt32 {
		return -1
	}

	secs := int64(best)*60 - WAKE_LEAD_SECONDS
	if secs < 1 {
		secs = 1
	}
	return secs
}

func ArmWakeTimer(s *settings.Settings) bool {
	h := ensureWakeTimer()
	if h == 0 {
		return false
	}

	var now windows.Systemtime
	windows.GetLocalTime(&now)

	secs := SecondsUntilWake(s, &now)
	if secs < 0 {
		procCancelWaitableTimer.Call(uintptr(h))
		return false
	}

	// negative is relative to now
	li := -secs * 10000000

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
	if wakeTimer != 0 {
		procCancelWaitableTimer.Call(uintptr(wakeTimer))
	}
}

func WakeTimersAllowed() bool {
	var scheme *windows.GUID
	value := uint32(1)

	ret, _, _ := procPowerGetActiveScheme.Call(0, uintptr(unsafe.Pointer(&scheme)))
	if ret == 0 && scheme != nil {
		var v uint32
		retRead, _, _ := procPowerReadACValueIndex.Call(
			0,
			uintptr(unsafe.Pointer(scheme)),
			uintptr(unsafe.Pointer(&kSleepSubgroup)),
			uintptr(unsafe.Pointer(&kAllowRtcWake)),
			uintptr(unsafe.Pointer(&v)),
		)
		if retRead == 0 {
			value = v
		}
		windows.LocalFree(windows.Handle(unsafe.Pointer(scheme)))
	}
	return value != 0
}

func KeepAwake(awake bool) {
	flags := uint32(ES_CONTINUOUS)
	if awake {
		flags |= ES_SYSTEM_REQUIRED | ES_DISPLAY_REQUIRED
	}
	procSetThreadExecutionState.Call(uintptr(flags))
}

func Cleanup() {
	KeepAwake(false)
	if wakeTimer != 0 {
		procCancelWaitableTimer.Call(uintptr(wakeTimer))
		syscall.CloseHandle(syscall.Handle(wakeTimer))
		wakeTimer = 0
	}
}
