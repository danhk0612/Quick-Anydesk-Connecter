//go:build windows

package main

import (
	_ "embed"
	"errors"
	"regexp"
	"syscall"
)

const (
	CF_BITMAP      = 2
	CF_DIB         = 8
	CF_UNICODETEXT = 13
	CF_DIBV5       = 17

	WM_DESTROY               = 0x0002
	WM_CLOSE                 = 0x0010
	WM_COMMAND               = 0x0111
	WM_KEYDOWN               = 0x0100
	WM_SETFONT               = 0x0030
	WM_CTLCOLOREDIT          = 0x0133
	WM_CTLCOLORSTATIC        = 0x0138
	WM_RBUTTONUP             = 0x0205
	WM_LBUTTONDBLCLK         = 0x0203
	WM_CONTEXTMENU           = 0x007B
	WM_CLIPBOARDUPDATE       = 0x031D
	WM_APP                   = 0x8000
	WM_TRAYICON              = WM_APP + 1
	WM_IMAGE_ANALYSIS_RESULT = WM_APP + 2

	VK_RETURN = 0x0D
	VK_ESCAPE = 0x1B

	WS_EX_TOPMOST  = 0x00000008
	WS_OVERLAPPED  = 0x00000000
	WS_CAPTION     = 0x00C00000
	WS_SYSMENU     = 0x00080000
	WS_VISIBLE     = 0x10000000
	WS_CHILD       = 0x40000000
	WS_TABSTOP     = 0x00010000
	WS_BORDER      = 0x00800000
	ES_AUTOHSCROLL = 0x0080
	ES_PASSWORD    = 0x0020

	BS_PUSHBUTTON    = 0x00000000
	BS_DEFPUSHBUTTON = 0x00000001
	SS_BITMAP        = 0x0000000E
	STM_SETIMAGE     = 0x0172
	IMAGE_BITMAP     = 0
	DIB_RGB_COLORS   = 0

	SW_SHOW       = 5
	SW_HIDE       = 0
	SW_SHOWNORMAL = 1

	MB_OK            = 0x00000000
	MB_OKCANCEL      = 0x00000001
	MB_YESNO         = 0x00000004
	MB_ICONERROR     = 0x00000010
	MB_ICONQUESTION  = 0x00000020
	MB_SETFOREGROUND = 0x00010000
	MB_TOPMOST       = 0x00040000
	IDOK             = 1
	IDCANCEL         = 2
	IDYES            = 6

	DEFAULT_GUI_FONT = 17
	COLOR_BTNFACE    = 15
	TRANSPARENT      = 1

	NIM_ADD     = 0x00000000
	NIM_DELETE  = 0x00000002
	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004

	MF_STRING    = 0x00000000
	MF_CHECKED   = 0x00000008
	MF_POPUP     = 0x00000010
	MF_SEPARATOR = 0x00000800

	TPM_RIGHTBUTTON = 0x0002
	TPM_RETURNCMD   = 0x0100

	ID_EDIT                 = 1001
	ID_OK                   = 1002
	ID_CANCEL               = 1003
	ID_MODEL_EDIT           = 1004
	ID_OPENROUTER_KEYS_LINK = 1005
	ID_ANALYZE              = 1010
	ID_IGNORE               = 1011
	ID_TRAY_CONNECT         = 2001
	ID_TRAY_IMAGE_ANALYSIS  = 2002
	ID_TRAY_CHANGE_PASSWORD = 2003
	ID_TRAY_OPENROUTER      = 2004
	ID_TRAY_STARTUP_ADD     = 2005
	ID_TRAY_STARTUP_REMOVE  = 2006
	ID_TRAY_LANGUAGE_KO     = 2007
	ID_TRAY_LANGUAGE_EN     = 2008
	ID_TRAY_EXIT            = 2009

	LR_DEFAULTCOLOR = 0x0000

	HKEY_CURRENT_USER = 0x80000001
	KEY_SET_VALUE     = 0x0002
	REG_SZ            = 1

	CRED_TYPE_GENERIC          = 1
	CRED_PERSIST_LOCAL_MACHINE = 2
)

var (
	errCancelled                   = errors.New("cancelled")
	errOpenRouterUnauthorized      = errors.New("openrouter unauthorized")
	errOpenRouterForbidden         = errors.New("openrouter forbidden")
	errOpenRouterRateLimit         = errors.New("openrouter rate limit")
	errOpenRouterUpstreamRateLimit = errors.New("openrouter upstream provider rate limit")
	errOpenRouterServer            = errors.New("openrouter server error")
	errOpenRouterTimeout           = errors.New("openrouter timeout")
	errOpenRouterNetwork           = errors.New("openrouter network error")
	errOpenRouterPayment           = errors.New("openrouter payment required")
	errOpenRouterInvalidResponse   = errors.New("openrouter invalid response")
	errAnyDeskNotFound             = errors.New("anydesk address not found")
	errMultipleAnyDeskIDs          = errors.New("multiple anydesk addresses")
	errOpenRouterBadRequest        = errors.New("openrouter bad request")
	errOpenRouterModelUnavailable  = errors.New("openrouter model unavailable")
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")

	procOpenClipboard                 = user32.NewProc("OpenClipboard")
	procCloseClipboard                = user32.NewProc("CloseClipboard")
	procGetClipboardData              = user32.NewProc("GetClipboardData")
	procIsClipboardFormatAvailable    = user32.NewProc("IsClipboardFormatAvailable")
	procRegisterClipboardFormatW      = user32.NewProc("RegisterClipboardFormatW")
	procGetDC                         = user32.NewProc("GetDC")
	procReleaseDC                     = user32.NewProc("ReleaseDC")
	procPostMessageW                  = user32.NewProc("PostMessageW")
	procRegisterClassExW              = user32.NewProc("RegisterClassExW")
	procCreateWindowExW               = user32.NewProc("CreateWindowExW")
	procDefWindowProcW                = user32.NewProc("DefWindowProcW")
	procDestroyWindow                 = user32.NewProc("DestroyWindow")
	procShowWindow                    = user32.NewProc("ShowWindow")
	procUpdateWindow                  = user32.NewProc("UpdateWindow")
	procGetMessageW                   = user32.NewProc("GetMessageW")
	procTranslateMessage              = user32.NewProc("TranslateMessage")
	procDispatchMessageW              = user32.NewProc("DispatchMessageW")
	procPostQuitMessage               = user32.NewProc("PostQuitMessage")
	procGetWindowTextLenW             = user32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW                = user32.NewProc("GetWindowTextW")
	procSetFocus                      = user32.NewProc("SetFocus")
	procMessageBoxW                   = user32.NewProc("MessageBoxW")
	procGetSystemMetrics              = user32.NewProc("GetSystemMetrics")
	procSendMessageW                  = user32.NewProc("SendMessageW")
	procGetSysColorBrush              = user32.NewProc("GetSysColorBrush")
	procAddClipboardFormatListener    = user32.NewProc("AddClipboardFormatListener")
	procRemoveClipboardFormatListener = user32.NewProc("RemoveClipboardFormatListener")
	procCreatePopupMenu               = user32.NewProc("CreatePopupMenu")
	procAppendMenuW                   = user32.NewProc("AppendMenuW")
	procTrackPopupMenu                = user32.NewProc("TrackPopupMenu")
	procDestroyMenu                   = user32.NewProc("DestroyMenu")
	procGetCursorPos                  = user32.NewProc("GetCursorPos")
	procSetForegroundWindow           = user32.NewProc("SetForegroundWindow")
	procLoadIconW                     = user32.NewProc("LoadIconW")
	procCreateIconFromResourceEx      = user32.NewProc("CreateIconFromResourceEx")

	procShellExecuteW = shell32.NewProc("ShellExecuteW")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	procGlobalLock       = kernel32.NewProc("GlobalLock")
	procGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
	procGlobalSize       = kernel32.NewProc("GlobalSize")

	procGetStockObject   = gdi32.NewProc("GetStockObject")
	procSetBkMode        = gdi32.NewProc("SetBkMode")
	procCreateDIBSection = gdi32.NewProc("CreateDIBSection")
	procGetDIBits        = gdi32.NewProc("GetDIBits")
	procGetObjectW       = gdi32.NewProc("GetObjectW")
	procDeleteObject     = gdi32.NewProc("DeleteObject")
	procShellNotifyIconW = shell32.NewProc("Shell_NotifyIconW")

	procRegOpenKeyExW   = advapi32.NewProc("RegOpenKeyExW")
	procRegSetValueExW  = advapi32.NewProc("RegSetValueExW")
	procRegDeleteValueW = advapi32.NewProc("RegDeleteValueW")
	procRegCloseKey     = advapi32.NewProc("RegCloseKey")
	procCredWriteW      = advapi32.NewProc("CredWriteW")
	procCredReadW       = advapi32.NewProc("CredReadW")
	procCredFree        = advapi32.NewProc("CredFree")
)

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type point struct {
	x int32
	y int32
}

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

type notifyIconData struct {
	cbSize           uint32
	hWnd             uintptr
	uID              uint32
	uFlags           uint32
	uCallbackMessage uint32
	hIcon            uintptr
	szTip            [128]uint16
	dwState          uint32
	dwStateMask      uint32
	szInfo           [256]uint16
	uVersion         uint32
	szInfoTitle      [64]uint16
	dwInfoFlags      uint32
	guidItem         [16]byte
	hBalloonIcon     uintptr
}

type bitmapInfoHeader struct {
	biSize          uint32
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

type rgbQuad struct{ blue, green, red, reserved byte }
type bitmapInfo struct {
	bmiHeader bitmapInfoHeader
	bmiColors [1]rgbQuad
}
type bitmap struct {
	bmType       int32
	bmWidth      int32
	bmHeight     int32
	bmWidthBytes int32
	bmPlanes     uint16
	bmBitsPixel  uint16
	bmBits       uintptr
}

type filetime struct {
	dwLowDateTime  uint32
	dwHighDateTime uint32
}

type credential struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        filetime
	CredentialBlobSize uint32
	CredentialBlob     *byte
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

type dialogMode int

const (
	modeAnyDeskID dialogMode = iota
	modePassword
	modeOpenRouterKey
)

var (
	mainWindow             uintptr
	trayIcon               notifyIconData
	exeDir                 string
	configPath             string
	password               string
	language               = "ko"
	imageAnalysisEnabled   bool
	openRouterModel        = defaultOpenRouterModel
	imageAnalysisBusy      bool
	lastClipboardImageHash [32]byte
	analysisResult         string
	analysisErr            error

	dialogWindow     uintptr
	dialogEdit       uintptr
	dialogModelEdit  uintptr
	dialogValue      string
	dialogModelValue string
	dialogCancel     bool
	dialogModeNow    dialogMode
	dialogActive     bool
	dialogBgBrush    uintptr

	idPattern = regexp.MustCompile(`^\d{9,10}$`)
)

//go:embed app.ico
var embeddedIcon []byte
