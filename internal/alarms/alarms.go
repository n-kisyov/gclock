package alarms

import (
	"github.com/n-kisyov/gclock/internal/settings"
	"golang.org/x/sys/windows"
)

func MinuteStamp(st *windows.Systemtime) uint64 {
	var ft windows.Filetime
	windows.SystemTimeToFileTime(st, &ft)
	// uint64 combining low and high
	u := (uint64(ft.HighDateTime) << 32) | uint64(ft.LowDateTime)
	return u / 600000000 // 100ns ticks in one minute
}

func StampToSystemtime(stamp uint64, st *windows.Systemtime) bool {
	u := stamp * 600000000
	var ft windows.Filetime
	ft.HighDateTime = uint32(u >> 32)
	ft.LowDateTime = uint32(u & 0xFFFFFFFF)
	err := windows.FileTimeToSystemTime(&ft, st)
	return err == nil
}

func NextDeltaMinutes(st *windows.Systemtime, a *settings.Alarm) (int, bool) {
	if !a.Enabled || a.Hour == settings.AlarmUnset || a.Minute == settings.AlarmUnset {
		return 0, false
	}

	nowMin := int(st.Hour)*60 + int(st.Minute)
	alarmMin := a.Hour*60 + a.Minute

	if a.RepeatDays == 0 {
		if alarmMin <= nowMin {
			return 0, false
		}
		return alarmMin - nowMin, true
	}

	for dayOffset := 0; dayOffset <= 7; dayOffset++ {
		day := (int(st.DayOfWeek) + dayOffset) % 7
		if (a.RepeatDays & (1 << uint(day))) == 0 {
			continue
		}
		if dayOffset == 0 && alarmMin <= nowMin {
			continue
		}
		return dayOffset*24*60 + (alarmMin - nowMin), true
	}
	return 0, false
}

func DueAt(a *settings.Alarm, st *windows.Systemtime) bool {
	if !a.Enabled || a.Hour == settings.AlarmUnset {
		return false
	}
	if a.Hour != int(st.Hour) || a.Minute != int(st.Minute) {
		return false
	}
	if a.RepeatDays != 0 && (a.RepeatDays&(1<<uint(st.DayOfWeek))) == 0 {
		return false
	}
	return true
}

func Check(s *settings.Settings, alarmsEnabled bool, alarmActive *bool, lastFireStamp *uint64, st *windows.Systemtime) (int, bool) {
	if !alarmsEnabled || *alarmActive {
		return -1, false
	}

	stamp := MinuteStamp(st)
	if stamp != 0 && stamp == *lastFireStamp {
		return -1, false
	}

	for i := 0; i < settings.MaxAlarms; i++ {
		a := &s.Alarms[i]
		if !DueAt(a, st) {
			continue
		}

		if a.SkipNext {
			a.SkipNext = false
			if a.RepeatDays == 0 {
				a.Enabled = false
			}
			*lastFireStamp = stamp
			s.SaveToDisk()
			return -1, false
		}

		if a.RepeatDays == 0 {
			a.Enabled = false
			s.SaveToDisk()
		}

		*alarmActive = true
		*lastFireStamp = stamp
		return i, true
	}
	return -1, false
}

func CatchUp(s *settings.Settings, alarmsEnabled bool, alarmActive *bool, lastFireStamp *uint64, lastSeenStamp uint64, nowStamp uint64, maxGapMinutes int) (int, windows.Systemtime, bool) {
	var when windows.Systemtime
	if !alarmsEnabled || *alarmActive {
		return -1, when, false
	}
	if lastSeenStamp == 0 || nowStamp <= lastSeenStamp {
		return -1, when, false
	}
	if nowStamp-lastSeenStamp > uint64(maxGapMinutes) {
		return -1, when, false
	}

	found := -1

	for m := lastSeenStamp + 1; m < nowStamp; m++ {
		var st windows.Systemtime
		if !StampToSystemtime(m, &st) {
			continue
		}

		for i := 0; i < settings.MaxAlarms; i++ {
			a := &s.Alarms[i]
			if !DueAt(a, &st) {
				continue
			}

			if a.SkipNext {
				a.SkipNext = false
				if a.RepeatDays == 0 {
					a.Enabled = false
				}
				continue
			}
			found = i
			when = st
		}
	}
	if found < 0 {
		return -1, when, false
	}

	if s.Alarms[found].RepeatDays == 0 {
		s.Alarms[found].Enabled = false
	}
	*alarmActive = true
	*lastFireStamp = nowStamp
	s.SaveToDisk()

	return found, when, true
}
