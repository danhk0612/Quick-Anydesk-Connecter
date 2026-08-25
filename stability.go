package main

const (
	WM_CLIPBOARD_IMAGE_READY  = WM_APP + 3
	WM_UPDATE_CHECK_RESULT    = WM_APP + 4
	WM_UPDATE_DOWNLOAD_RESULT = WM_APP + 5
	ID_TRAY_STARTUP           = ID_TRAY_STARTUP_ADD
	ID_TRAY_CHECK_UPDATE      = ID_TRAY_STARTUP_REMOVE
	KEY_QUERY_VALUE           = 0x0001
)

var (
	clipboardImageBusy  bool
	procRegQueryValueExW = advapi32.NewProc("RegQueryValueExW")
)
