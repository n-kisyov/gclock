package ui

import "github.com/lxn/win"

const (
	IDI_APPICON          = 101
	IDR_DIGITALFONT      = 102

	IDR_APP_MANIFEST     = 1

	IDD_SETTINGS         = 200
	IDC_DARKMODE         = 201
	IDC_CLOCK_DIGITAL    = 202
	IDC_CLOCK_ANALOG     = 203
	IDC_ALARMS_ENABLED   = 204
	IDC_SOUND_SIMPLE     = 205
	IDC_SOUND_MP3        = 206
	IDC_ALARM_COUNT      = 207
	IDC_HOUR24           = 208
	IDC_SNOOZE_MINUTES   = 209
	IDC_CRESCENDO        = 210
	IDC_AUTOSTART        = 211
	IDC_START_MINIMIZED  = 212
	IDC_ACRYLIC          = 213
	IDC_ALWAYS_ON_TOP    = 214
	IDC_ALARM_VOLUME     = 215
	IDC_PREVIEW_SOUND    = 216
	IDC_SLEEP_MINUTES    = 217
	IDC_WAKE_NOTE        = 218

	IDD_ALARM            = 300
	IDC_ALARM_HOUR       = 301
	IDC_ALARM_MINUTE     = 302
	IDC_ALARM_ENABLED    = 303
	IDC_ALARM_LABEL      = 304
	IDC_ALARM_REPEAT     = 305
	IDC_ALARM_SOUND      = 306
	IDC_ALARM_SOUND_PICK = 307
	IDC_ALARM_SOUND_CLEAR = 308
	IDC_ALARM_VOL        = 309
	IDC_DAY_SUN          = 310
	IDC_DAY_MON          = 311
	IDC_DAY_TUE          = 312
	IDC_DAY_WED          = 313
	IDC_DAY_THU          = 314
	IDC_DAY_FRI          = 315
	IDC_DAY_SAT          = 316
	IDC_DAY_ALL          = 317
	IDC_DAY_NONE         = 318
	IDC_ALARM_SNOOZE     = 320
	IDC_ALARM_SKIP       = 321

	IDD_COUNTDOWN_SET    = 400
	IDC_CD_HOURS         = 401
	IDC_CD_MINS          = 402
	IDC_CD_SECS          = 403
	IDC_CD_PRESET_5      = 404
	IDC_CD_PRESET_10     = 405
	IDC_CD_PRESET_25     = 406

	APP_MODE_CLOCK       = 0
	APP_MODE_COUNTDOWN   = 1
	APP_MODE_STOPWATCH   = 2

	// IDM_SETTINGS is already defined in tray.go as 40002.
	// Oh wait, resource.h has:
	// IDM_SETTINGS         1001
	// IDM_EXIT             1002
	// IDM_ABOUT            1003
	// IDM_TRAY_SHOW        2001
	// IDM_TRAY_EXIT        2002
	// Let's match resource.h strictly so dialogs and menus match.
)

const (
	WM_AUDIO_TRACK_DONE = win.WM_APP + 3
	TIMER_CLOCK         = 1
	// TIMER_SOUND_PREVIEW = 2 (defined in audio package as 1001, we'll keep 1001)
)
