package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	MaxAlarms = 10
	AlarmUnset = -1

	ClockDigital = 0
	ClockAnalog  = 1

	SoundSimple = 0
	SoundMP3    = 1
)

type Alarm struct {
	Hour          int    `json:"hour"`
	Minute        int    `json:"minute"`
	Enabled       bool   `json:"enabled"`
	Label         string `json:"label"`
	RepeatDays    uint8  `json:"repeat_days"`
	Sound         string `json:"sound"`
	Volume        int    `json:"volume"`
	SnoozeMinutes int    `json:"snooze_minutes"`
	SkipNext      bool   `json:"skip_next"`
}

type Settings struct {
	DarkMode        bool   `json:"dark_mode"`
	ClockStyle      int    `json:"clock_style"`
	AlarmsEnabled   bool   `json:"alarms_enabled"`
	SoundMode       int    `json:"sound_mode"`
	Hour24          bool   `json:"hour24"`
	Crescendo       bool   `json:"crescendo"`
	AutoStart       bool   `json:"autostart"`
	StartMinimized  bool   `json:"start_minimized"`
	Acrylic         bool   `json:"acrylic"`
	AlwaysOnTop     bool   `json:"always_on_top"`
	AlarmVolume     int    `json:"alarm_volume"`
	SnoozeMinutes   int    `json:"snooze_minutes"`
	AppMode         int    `json:"app_mode"`
	SleepMinutes    int    `json:"sleep_minutes"`

	CDHours         int    `json:"cd_hours"`
	CDMins          int    `json:"cd_mins"`
	CDSecs          int    `json:"cd_secs"`

	WinX            int32  `json:"win_x"`
	WinY            int32  `json:"win_y"`
	WinW            int32  `json:"win_w"`
	WinH            int32  `json:"win_h"`

	Alarms          []Alarm `json:"alarms"`
}

func Default() *Settings {
	s := &Settings{
		DarkMode:       true,
		ClockStyle:     ClockDigital,
		AlarmsEnabled:  true,
		SoundMode:      SoundMP3,
		Hour24:         true,
		Crescendo:      true,
		AutoStart:      false,
		StartMinimized: false,
		Acrylic:        true,
		AlwaysOnTop:    false,
		AlarmVolume:    80,
		SnoozeMinutes:  5,
		AppMode:        0,
		SleepMinutes:   30,
		CDHours:        0,
		CDMins:         5,
		CDSecs:         0,
		WinX:           -1,
		WinY:           -1,
		WinW:           -1,
		WinH:           -1,
		Alarms:         make([]Alarm, MaxAlarms),
	}
	for i := 0; i < MaxAlarms; i++ {
		s.Alarms[i] = Alarm{
			Hour:          AlarmUnset,
			Volume:        -1,
			SnoozeMinutes: -1,
		}
	}
	return s
}

func GetSettingsPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), "gclock_settings.json"), nil
}

func Load(path string) *Settings {
	s := Default()
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, s)
	}
	// ensure bounds
	if len(s.Alarms) != MaxAlarms {
		alarms := make([]Alarm, MaxAlarms)
		copy(alarms, s.Alarms)
		for i := len(s.Alarms); i < MaxAlarms; i++ {
			alarms[i] = Alarm{Hour: AlarmUnset, Volume: -1, SnoozeMinutes: -1}
		}
		s.Alarms = alarms
	}
	return s
}

func Save(path string, s *Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	// On Windows, Rename fails if the destination exists. Use Remove first.
	os.Remove(path)
	return os.Rename(tmp, path)
}

func (s *Settings) SaveToDisk() error {
	path, err := GetSettingsPath()
	if err != nil {
		return err
	}
	return Save(path, s)
}
