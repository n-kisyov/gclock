package audio

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lxn/win"
	"github.com/n-kisyov/glock/internal/settings"
)

const (
	CrescendoMs    = 15000
	CrescendoFloor = 0.10
	PreviewMs      = 3000

	TimerSoundPreview = 1001 // Match this with UI timer ID
)

var audioExtensions = []string{".mp3", ".wav", ".flac", ".m4a", ".wma", ".aac", ".mp4"}

type SoundManager struct {
	ExeDir        string
	NotifyWnd     win.HWND
	Settings      *settings.Settings
	
	AlarmActive   bool
	RingingAlarm  int
	SoundPreview  bool
	SleepRunning  bool
	SleepEnd      time.Time
	
	tracks        []string
	trackIdx      int
	songsMTime    time.Time
	crescendoEnd  time.Time
}

func NewSoundManager(exeDir string, notifyWnd win.HWND, set *settings.Settings) *SoundManager {
	return &SoundManager{
		ExeDir:    exeDir,
		NotifyWnd: notifyWnd,
		Settings:  set,
	}
}

func hasAudioExtension(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	for _, e := range audioExtensions {
		if ext == e {
			return true
		}
	}
	return false
}

func (sm *SoundManager) targetGain() float32 {
	v := sm.Settings.AlarmVolume
	if !sm.SoundPreview && sm.RingingAlarm >= 0 && sm.RingingAlarm < settings.MaxAlarms {
		if av := sm.Settings.Alarms[sm.RingingAlarm].Volume; av >= 0 {
			v = av
		}
	}
	if v < 0 { v = 0 }
	if v > 100 { v = 100 }
	return float32(v) / 100.0
}

func (sm *SoundManager) shuffleTracks() {
	rand.Shuffle(len(sm.tracks), func(i, j int) {
		sm.tracks[i], sm.tracks[j] = sm.tracks[j], sm.tracks[i]
	})
	sm.trackIdx = 0
}

func (sm *SoundManager) scanTracks() bool {
	sm.tracks = nil
	sm.trackIdx = 0

	dir := filepath.Join(sm.ExeDir, "songs")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, e := range entries {
		if !e.IsDir() && hasAudioExtension(e.Name()) {
			sm.tracks = append(sm.tracks, filepath.Join(dir, e.Name()))
		}
	}
	if len(sm.tracks) == 0 {
		return false
	}
	sm.shuffleTracks()
	return true
}

func (sm *SoundManager) ensureTracks() bool {
	dir := filepath.Join(sm.ExeDir, "songs")
	info, err := os.Stat(dir)
	haveStamp := false
	var mtime time.Time
	if err == nil && info.IsDir() {
		haveStamp = true
		mtime = info.ModTime()
		if len(sm.tracks) > 0 && mtime.Equal(sm.songsMTime) {
			return true
		}
	}

	if !sm.scanTracks() {
		return false
	}
	if haveStamp {
		sm.songsMTime = mtime
	}
	return true
}

func (sm *SoundManager) playNextTrack() bool {
	for tried := 0; tried < len(sm.tracks); tried++ {
		if sm.trackIdx >= len(sm.tracks) {
			sm.shuffleTracks()
		}
		path := sm.tracks[sm.trackIdx]
		sm.trackIdx++
		if PlayFile(path, sm.NotifyWnd) {
			return true
		}
	}
	return false
}

func (sm *SoundManager) PlayAlarm() {
	started := false
	sm.SleepRunning = false
	sm.SleepEnd = time.Time{}

	var ownPath string
	if !sm.SoundPreview && sm.RingingAlarm >= 0 && sm.RingingAlarm < settings.MaxAlarms {
		ownPath = sm.Settings.Alarms[sm.RingingAlarm].Sound
	}

	if ownPath != "" {
		if !filepath.IsAbs(ownPath) {
			ownPath = filepath.Join(sm.ExeDir, "songs", ownPath)
		}
		started = PlayFile(ownPath, sm.NotifyWnd)
	}

	if !started && sm.Settings.SoundMode == "mp3" && sm.ensureTracks() {
		started = sm.playNextTrack()
	}

	if !started {
		started = PlayTone(sm.NotifyWnd)
	}
	if !started {
		return
	}

	target := sm.targetGain()
	if sm.Settings.Crescendo && !sm.SoundPreview {
		RampGain(CrescendoFloor*target, target, CrescendoMs)
		sm.crescendoEnd = time.Now().Add(CrescendoMs * time.Millisecond)
	} else {
		SetGain(target)
		sm.crescendoEnd = time.Time{}
	}

	if sm.SoundPreview && sm.NotifyWnd != 0 {
		win.SetTimer(sm.NotifyWnd, TimerSoundPreview, PreviewMs, 0)
	}
}

func (sm *SoundManager) StopAlarm() {
	if sm.NotifyWnd != 0 {
		win.KillTimer(sm.NotifyWnd, TimerSoundPreview)
	}
	sm.SoundPreview = false
	sm.crescendoEnd = time.Time{}
	Stop()
}

func (sm *SoundManager) StartSleepTimer() bool {
	if sm.AlarmActive {
		return false
	}
	if !sm.ensureTracks() || !sm.playNextTrack() {
		return false
	}

	mins := sm.Settings.SleepMinutes
	if mins <= 0 {
		mins = 30
	}
	RampGain(sm.targetGain(), 0.0, uint32(mins)*60*1000)

	sm.SleepRunning = true
	sm.SleepEnd = time.Now().Add(time.Duration(mins) * time.Minute)
	return true
}

func (sm *SoundManager) StopSleepTimer() {
	sm.SleepRunning = false
	sm.SleepEnd = time.Time{}
	Stop()
}

func (sm *SoundManager) OnTrackDone() {
	if sm.SleepRunning && !sm.AlarmActive {
		carried := GetGain()
		now := time.Now()

		if !sm.SleepEnd.After(now) || !sm.playNextTrack() {
			sm.StopSleepTimer()
			return
		}
		RampGain(carried, 0.0, uint32(sm.SleepEnd.Sub(now).Milliseconds()))
		return
	}

	if !sm.AlarmActive {
		return
	}

	var ownPath string
	if sm.RingingAlarm >= 0 && sm.RingingAlarm < settings.MaxAlarms {
		ownPath = sm.Settings.Alarms[sm.RingingAlarm].Sound
	}

	carried := GetGain()
	started := false

	if ownPath != "" {
		if !filepath.IsAbs(ownPath) {
			ownPath = filepath.Join(sm.ExeDir, "songs", ownPath)
		}
		started = PlayFile(ownPath, sm.NotifyWnd)
	} else if sm.Settings.SoundMode == "mp3" {
		started = sm.playNextTrack()
	} else {
		return
	}

	if !started {
		return
	}

	now := time.Now()
	if sm.crescendoEnd.After(now) {
		RampGain(carried, sm.targetGain(), uint32(sm.crescendoEnd.Sub(now).Milliseconds()))
	} else {
		SetGain(sm.targetGain())
	}
}

func (sm *SoundManager) Cleanup() {
	sm.tracks = nil
	sm.songsMTime = time.Time{}
}
