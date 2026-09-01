package settings

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

const MaxAlarms = 10

type Alarm struct {
	Hour          int    `json:"hour"`
	Minute        int    `json:"minute"`
	Enabled       bool   `json:"enabled"`
	Label         string `json:"label"`
	RepeatDays    uint8  `json:"repeat_days"`
	Volume        int    `json:"volume"`
	SnoozeMinutes int    `json:"snooze_minutes"`
	SkipNext      bool   `json:"skip_next"`
	Sound         string `json:"sound"`
}

type Settings struct {
	DarkMode        bool    `json:"dark_mode"`
	Hour24          bool    `json:"hour24"`
	Crescendo       bool    `json:"crescendo"`
	AutoStart       bool    `json:"autostart"`
	StartMinimized  bool    `json:"start_minimized"`
	Acrylic         bool    `json:"acrylic"`
	AlwaysOnTop     bool    `json:"always_on_top"`
	ClockStyle      string  `json:"clock_style"` // "analog" or "digital"
	AlarmsEnabled   bool    `json:"alarms_enabled"`
	AlarmCount      int     `json:"alarm_count"`
	AlarmsCollapsed bool    `json:"alarms_collapsed"`
	AlarmVolume     int     `json:"alarm_volume"`
	SnoozeMinutes   int     `json:"snooze_minutes"`
	SleepMinutes    int     `json:"sleep_minutes"`
	LastSeen        int     `json:"last_seen"`
	AppMode         int     `json:"app_mode"`
	WinX            int     `json:"win_x"`
	WinY            int     `json:"win_y"`
	WinW            int     `json:"win_w"`
	WinH            int     `json:"win_h"`
	SoundMode       string  `json:"sound_mode"` // "mp3" or "simple"
	CDHours         int     `json:"cd_hours"`
	CDMins          int     `json:"cd_mins"`
	CDSecs          int     `json:"cd_secs"`
	Alarms          []Alarm `json:"alarms"`
}

// Default creates a Settings struct with safe default values.
func Default() *Settings {
	s := &Settings{
		AlarmCount:    1,
		AlarmVolume:   100,
		SnoozeMinutes: 5,
		SleepMinutes:  30,
		Alarms:        make([]Alarm, MaxAlarms),
		ClockStyle:    "digital",
		SoundMode:     "mp3",
	}
	for i := range s.Alarms {
		s.Alarms[i].Hour = -1
		s.Alarms[i].Minute = -1
		s.Alarms[i].Volume = -1
		s.Alarms[i].SnoozeMinutes = -1
	}
	return s
}

// Load reads settings from the given file path.
func Load(path string) (*Settings, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), errors.New("settings file not found")
		}
		return Default(), err
	}

	s := Default()
	// Unmarshal overrides existing values. Since we pre-filled Alarms with defaults,
	// if the JSON array has fewer than 10 alarms, the rest remain defaulted.
	// If it has more, it will grow, which we can truncate.
	if err := json.Unmarshal(data, s); err != nil {
		return s, err
	}

	// Clamp alarm array
	if len(s.Alarms) > MaxAlarms {
		s.Alarms = s.Alarms[:MaxAlarms]
	} else if len(s.Alarms) < MaxAlarms {
		for len(s.Alarms) < MaxAlarms {
			s.Alarms = append(s.Alarms, Alarm{Hour: -1, Minute: -1, Volume: -1, SnoozeMinutes: -1})
		}
	}

	// Validate / Clamp values
	if s.AlarmCount < 1 {
		s.AlarmCount = 1
	} else if s.AlarmCount > MaxAlarms {
		s.AlarmCount = MaxAlarms
	}

	if s.AlarmVolume < 10 {
		s.AlarmVolume = 10
	} else if s.AlarmVolume > 100 {
		s.AlarmVolume = 100
	}

	if s.SnoozeMinutes < 1 {
		s.SnoozeMinutes = 1
	} else if s.SnoozeMinutes > 60 {
		s.SnoozeMinutes = 60
	}

	if s.SleepMinutes < 1 {
		s.SleepMinutes = 1
	} else if s.SleepMinutes > 240 {
		s.SleepMinutes = 240
	}

	for i := range s.Alarms {
		a := &s.Alarms[i]
		if a.Hour < 0 || a.Hour > 23 || a.Minute < 0 || a.Minute > 59 {
			a.Hour = -1
			a.Minute = -1
			a.Enabled = false
			a.SkipNext = false
		}
	}

	return s, nil
}

// Save writes settings to a .tmp file and renames it over the target path.
func Save(path string, s *Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	bakPath := path + ".bak"

	// Write to tmp
	if err := ioutil.WriteFile(tmpPath, append(data, '\n'), 0644); err != nil {
		return err
	}

	// Try to backup existing
	if _, err := os.Stat(path); err == nil {
		_ = os.Rename(path, bakPath)
	}

	// Move tmp to path
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	return nil
}

func GetSettingsPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), "alarmclock_settings.json"), nil
}
