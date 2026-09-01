package alarms

import (
	"time"
	"github.com/n-kisyov/glock/internal/settings"
)

// MinuteStamp returns whole minutes since the Windows FILETIME epoch (1601-01-01).
// This matches the C version exactly so saved JSON 'last_seen' values remain valid.
func MinuteStamp(t time.Time) uint64 {
	// Unix time is seconds since 1970-01-01.
	// Windows epoch is 11644473600 seconds before Unix epoch.
	return uint64(t.Unix()+11644473600) / 60
}

func StampToTime(stamp uint64) time.Time {
	secs := int64(stamp*60) - 11644473600
	return time.Unix(secs, 0)
}

func NextDeltaMinutes(now time.Time, a *settings.Alarm) (int, bool) {
	if !a.Enabled || a.Hour < 0 || a.Minute < 0 {
		return 0, false
	}

	nowMin := now.Hour()*60 + now.Minute()
	alarmMin := a.Hour*60 + a.Minute

	if a.RepeatDays == 0 {
		if alarmMin <= nowMin {
			return 0, false
		}
		return alarmMin - nowMin, true
	}

	// dayOffset up to 7 (same weekday, next week)
	for dayOffset := 0; dayOffset <= 7; dayOffset++ {
		day := (int(now.Weekday()) + dayOffset) % 7
		if (a.RepeatDays & (1 << day)) == 0 {
			continue
		}
		if dayOffset == 0 && alarmMin <= nowMin {
			continue
		}
		return dayOffset*24*60 + (alarmMin - nowMin), true
	}
	return 0, false
}

func DueAt(a *settings.Alarm, st time.Time) bool {
	if !a.Enabled || a.Hour < 0 {
		return false
	}
	if a.Hour != st.Hour() || a.Minute != st.Minute() {
		return false
	}
	if a.RepeatDays != 0 && (a.RepeatDays&(1<<st.Weekday())) == 0 {
		return false
	}
	return true
}

func CatchUp(s *settings.Settings, alarmActive *bool, lastFireStamp *uint64, nowStamp uint64, maxGapMinutes int) (int, time.Time, bool) {
	if !s.AlarmsEnabled || *alarmActive {
		return -1, time.Time{}, false
	}
	
	lastSeen := uint64(s.LastSeen)
	if lastSeen == 0 || nowStamp <= lastSeen {
		return -1, time.Time{}, false
	}
	if nowStamp-lastSeen > uint64(maxGapMinutes) {
		return -1, time.Time{}, false
	}

	found := -1
	var when time.Time

	for m := lastSeen + 1; m < nowStamp; m++ {
		st := StampToTime(m)

		for i := 0; i < settings.MaxAlarms; i++ {
			if !DueAt(&s.Alarms[i], st) {
				continue
			}

			if s.Alarms[i].SkipNext {
				s.Alarms[i].SkipNext = false
				if s.Alarms[i].RepeatDays == 0 {
					s.Alarms[i].Enabled = false
				}
				continue
			}
			found = i
			when = st
		}
	}

	if found < 0 {
		return -1, time.Time{}, false
	}

	if s.Alarms[found].RepeatDays == 0 {
		s.Alarms[found].Enabled = false
	}
	*alarmActive = true
	*lastFireStamp = nowStamp
	
	return found, when, true
}

func Check(s *settings.Settings, st time.Time, alarmActive *bool, lastFireStamp *uint64) (int, bool) {
	if !s.AlarmsEnabled || *alarmActive {
		return -1, false
	}

	stamp := MinuteStamp(st)
	if stamp != 0 && stamp == *lastFireStamp {
		return -1, false
	}

	for i := 0; i < settings.MaxAlarms; i++ {
		if !DueAt(&s.Alarms[i], st) {
			continue
		}

		if s.Alarms[i].SkipNext {
			s.Alarms[i].SkipNext = false
			if s.Alarms[i].RepeatDays == 0 {
				s.Alarms[i].Enabled = false
			}
			*lastFireStamp = stamp
			return -1, false
		}

		if s.Alarms[i].RepeatDays == 0 {
			s.Alarms[i].Enabled = false
		}

		*alarmActive = true
		*lastFireStamp = stamp
		return i, true
	}
	return -1, false
}
