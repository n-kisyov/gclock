package ui

import (
	"path/filepath"
	"os"

	"github.com/lxn/win"
	"github.com/n-kisyov/glock/internal/settings"
)

type AppState struct {
	Settings *settings.Settings

	AppMode         int
	WinX, WinY      int32
	WinW, WinH      int32

	AlarmActive     bool
	RingingAlarm    int
	LastFireStamp   uint64
	LastSeenStamp   uint64
	AlarmStartedMs  uint64
	AutoSnoozeCount int

	SnoozePending   bool
	SnoozeEndMs     uint64
	SnoozeTotalSec  int

	CdRemainingMs   int
	CdRunning       bool
	CdLastTick      uint64

	SwRunning       bool
	SwStartTick     uint32
	SwAccumulatedMs uint32

	HMainWnd        win.HWND
	Fonts           ClockFonts
	ClockAreaH      int32
	Dpi             int32

	Colors          ThemeColors

	Nid             win.NOTIFYICONDATA
	TrayAdded       bool

	ExeDir          string
}

func NewAppState() *AppState {
	s := &AppState{
		Settings:     settings.Default(),
		RingingAlarm: -1,
	}
	exe, _ := os.Executable()
	s.ExeDir = filepath.Dir(exe)
	return s
}

func (s *AppState) AlarmVolumeFor(idx int) int {
	if idx >= 0 && idx < settings.MaxAlarms && s.Settings.Alarms[idx].Volume >= 0 {
		return s.Settings.Alarms[idx].Volume
	}
	return s.Settings.AlarmVolume
}

func (s *AppState) AlarmSnoozeFor(idx int) int {
	if idx >= 0 && idx < settings.MaxAlarms && s.Settings.Alarms[idx].SnoozeMinutes > 0 {
		return s.Settings.Alarms[idx].SnoozeMinutes
	}
	return s.Settings.SnoozeMinutes
}

func (s *AppState) AlarmSoundFor(idx int) string {
	if idx >= 0 && idx < settings.MaxAlarms && s.Settings.Alarms[idx].Sound != "" {
		return s.Settings.Alarms[idx].Sound
	}
	return ""
}

func (s *AppState) CdTotalMs() int {
	total := int64(s.Settings.CDHours)*3600 + int64(s.Settings.CDMins)*60 + int64(s.Settings.CDSecs)
	total *= 1000
	if total < 0 { total = 0 }
	// int32 max
	if total > 2147483647 { total = 2147483647 }
	return int(total)
}
