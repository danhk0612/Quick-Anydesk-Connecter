//go:build windows

package main

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	WM_UPDATE_PROGRESS = WM_APP + 6
	WM_TIMER           = 0x0113

	ID_UPDATE_CANCEL  = 1020
	ID_UPDATE_RESTART = 1021

	PBM_SETRANGE32 = 0x0406
	PBM_SETPOS     = 0x0402

	ICC_PROGRESS_CLASS = 0x00000020
	updateRestartTimer = 1
)

type initCommonControlsEx struct {
	dwSize uint32
	dwICC  uint32
}

type updateProgressSnapshot struct {
	status     string
	detail     string
	percent    int
	cancelable bool
}

var (
	updateProgressWindow uintptr
	updateStatusLabel    uintptr
	updateProgressBar    uintptr
	updateDetailLabel    uintptr
	updateCancelButton   uintptr
	updateCancelFunc     context.CancelFunc
	updateProgressState  updateProgressSnapshot

	updateRestartWindow         uintptr
	updateRestartCountdownLabel uintptr
	updateRestartTarget         string
	updateRestartSeconds        int
	updateRestartLanguage       string
)

func updateUIText(ko, en string) string {
	if language == "en" {
		return en
	}
	return ko
}

func ensureCommonControls() {
	icc := initCommonControlsEx{dwSize: uint32(unsafe.Sizeof(initCommonControlsEx{})), dwICC: ICC_PROGRESS_CLASS}
	procInitCommonControlsEx.Call(uintptr(unsafe.Pointer(&icc)))
}

func setWindowText(hwnd uintptr, text string) {
	if hwnd == 0 {
		return
	}
	p, _ := syscall.UTF16PtrFromString(text)
	procSetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(p)))
}

func openUpdateProgressWindow(latest string, cancel context.CancelFunc) error {
	if updateProgressWindow != 0 {
		return nil
	}
	ensureCommonControls()
	if dialogBgBrush == 0 {
		dialogBgBrush, _, _ = procGetSysColorBrush.Call(COLOR_BTNFACE)
	}

	className, _ := syscall.UTF16PtrFromString("QuickAnydeskConnectUpdateProgress")
	title, _ := syscall.UTF16PtrFromString(updateUIText("Quick Anydesk Connect 업데이트", "Quick Anydesk Connect Update"))
	hInstance, _, _ := procGetModuleHandleW.Call(0)
	icon := appWindowIcon()
	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   syscall.NewCallback(updateProgressWindowProc),
		hInstance:     hInstance,
		hIcon:         icon,
		hIconSm:       icon,
		hbrBackground: dialogBgBrush,
		lpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	width, height := int32(520), int32(250)
	screenW, _, _ := procGetSystemMetrics.Call(0)
	screenH, _, _ := procGetSystemMetrics.Call(1)
	x := (int32(screenW) - width) / 2
	y := (int32(screenH) - height) / 2
	updateProgressWindow, _, _ = procCreateWindowExW.Call(
		WS_EX_TOPMOST,
		uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(title)),
		WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU,
		uintptr(x), uintptr(y), uintptr(width), uintptr(height),
		0, 0, hInstance, 0,
	)
	if updateProgressWindow == 0 {
		return fmt.Errorf("%s", currentMessages().errDialogCreate)
	}

	font, _, _ := procGetStockObject.Call(DEFAULT_GUI_FONT)
	staticClass, _ := syscall.UTF16PtrFromString("STATIC")
	progressClass, _ := syscall.UTF16PtrFromString("msctls_progress32")
	buttonClass, _ := syscall.UTF16PtrFromString("BUTTON")

	versionText := fmt.Sprintf(updateUIText("새 버전 v%s을(를) 설치하고 있습니다.", "Installing v%s."), latest)
	versionW, _ := syscall.UTF16PtrFromString(versionText)
	versionLabel, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(staticClass)), uintptr(unsafe.Pointer(versionW)), WS_CHILD|WS_VISIBLE, 24, 20, 455, 22, updateProgressWindow, 0, hInstance, 0)
	setFont(versionLabel, font)

	updateStatusLabel, _, _ = procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(staticClass)), 0, WS_CHILD|WS_VISIBLE, 24, 58, 455, 22, updateProgressWindow, 0, hInstance, 0)
	setFont(updateStatusLabel, font)

	updateProgressBar, _, _ = procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(progressClass)), 0, WS_CHILD|WS_VISIBLE, 24, 88, 455, 24, updateProgressWindow, 0, hInstance, 0)
	procSendMessageW.Call(updateProgressBar, PBM_SETRANGE32, 0, 100)
	procSendMessageW.Call(updateProgressBar, PBM_SETPOS, 0, 0)

	updateDetailLabel, _, _ = procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(staticClass)), 0, WS_CHILD|WS_VISIBLE, 24, 122, 455, 22, updateProgressWindow, 0, hInstance, 0)
	setFont(updateDetailLabel, font)

	cancelText, _ := syscall.UTF16PtrFromString(updateUIText("업데이트 중단", "Cancel Update"))
	updateCancelButton, _, _ = procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(buttonClass)), uintptr(unsafe.Pointer(cancelText)), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 344, 164, 135, 30, updateProgressWindow, ID_UPDATE_CANCEL, hInstance, 0)
	setFont(updateCancelButton, font)

	updateCancelFunc = cancel
	dialogActive = true
	procShowWindow.Call(updateProgressWindow, SW_SHOW)
	procUpdateWindow.Call(updateProgressWindow)
	procSetForegroundWindow.Call(updateProgressWindow)
	return nil
}

func closeUpdateProgressWindow() {
	if updateProgressWindow != 0 {
		procDestroyWindow.Call(updateProgressWindow)
	}
	updateProgressWindow = 0
	updateStatusLabel = 0
	updateProgressBar = 0
	updateDetailLabel = 0
	updateCancelButton = 0
	updateCancelFunc = nil
	dialogActive = false
}

func publishUpdateProgress(status, detail string, percent int, cancelable bool) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	updateMu.Lock()
	updateProgressState = updateProgressSnapshot{status: status, detail: detail, percent: percent, cancelable: cancelable}
	updateMu.Unlock()
	if updateProgressWindow != 0 {
		procPostMessageW.Call(updateProgressWindow, WM_UPDATE_PROGRESS, 0, 0)
	}
}

func handleUpdateProgressUI() {
	updateMu.Lock()
	state := updateProgressState
	updateMu.Unlock()
	applyUpdateProgressState(state)
}

func applyUpdateProgressState(state updateProgressSnapshot) {
	setWindowText(updateStatusLabel, state.status)
	setWindowText(updateDetailLabel, state.detail)
	if updateProgressBar != 0 {
		procSendMessageW.Call(updateProgressBar, PBM_SETPOS, uintptr(state.percent), 0)
	}
	if updateCancelButton != 0 {
		enabled := uintptr(0)
		if state.cancelable {
			enabled = 1
		}
		procEnableWindow.Call(updateCancelButton, enabled)
	}
}

func requestUpdateCancel() {
	updateMu.Lock()
	cancelable := updateProgressState.cancelable
	cancel := updateCancelFunc
	if cancelable {
		updateProgressState.cancelable = false
		updateProgressState.status = updateUIText("업데이트를 중단하는 중...", "Cancelling update...")
	}
	state := updateProgressState
	updateMu.Unlock()
	applyUpdateProgressState(state)
	if cancelable && cancel != nil {
		cancel()
	}
}

func updateProgressWindowProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case WM_UPDATE_PROGRESS:
		handleUpdateProgressUI()
		return 0
	case WM_COMMAND:
		if uint16(wParam&0xffff) == ID_UPDATE_CANCEL {
			requestUpdateCancel()
			return 0
		}
	case WM_CLOSE:
		requestUpdateCancel()
		return 0
	case WM_DESTROY:
		if hwnd == updateProgressWindow {
			updateProgressWindow = 0
		}
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return r
}

func runUpdateCompleteCountdown(targetPath, oldVersion, newVersion, lang string) error {
	ensureCommonControls()
	if dialogBgBrush == 0 {
		dialogBgBrush, _, _ = procGetSysColorBrush.Call(COLOR_BTNFACE)
	}
	updateRestartTarget = targetPath
	updateRestartSeconds = 5
	updateRestartLanguage = lang

	className, _ := syscall.UTF16PtrFromString("QuickAnydeskConnectUpdateComplete")
	titleText := "Quick Anydesk Connect 업데이트"
	if lang == "en" {
		titleText = "Quick Anydesk Connect Update"
	}
	title, _ := syscall.UTF16PtrFromString(titleText)
	hInstance, _, _ := procGetModuleHandleW.Call(0)
	icon := appWindowIcon()
	wc := wndClassEx{cbSize: uint32(unsafe.Sizeof(wndClassEx{})), lpfnWndProc: syscall.NewCallback(updateCompleteWindowProc), hInstance: hInstance, hIcon: icon, hIconSm: icon, hbrBackground: dialogBgBrush, lpszClassName: className}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	width, height := int32(470), int32(235)
	screenW, _, _ := procGetSystemMetrics.Call(0)
	screenH, _, _ := procGetSystemMetrics.Call(1)
	x := (int32(screenW) - width) / 2
	y := (int32(screenH) - height) / 2
	updateRestartWindow, _, _ = procCreateWindowExW.Call(WS_EX_TOPMOST, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(title)), WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU, uintptr(x), uintptr(y), uintptr(width), uintptr(height), 0, 0, hInstance, 0)
	if updateRestartWindow == 0 {
		return fmt.Errorf("could not create update completion window")
	}

	font, _, _ := procGetStockObject.Call(DEFAULT_GUI_FONT)
	staticClass, _ := syscall.UTF16PtrFromString("STATIC")
	buttonClass, _ := syscall.UTF16PtrFromString("BUTTON")
	completeText := "업데이트가 완료되었습니다."
	versionText := fmt.Sprintf("v%s → v%s", oldVersion, newVersion)
	restartText := "지금 재시작"
	if lang == "en" {
		completeText = "The update is complete."
		restartText = "Restart Now"
	}
	p, _ := syscall.UTF16PtrFromString(completeText)
	label, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(staticClass)), uintptr(unsafe.Pointer(p)), WS_CHILD|WS_VISIBLE, 24, 24, 410, 24, updateRestartWindow, 0, hInstance, 0)
	setFont(label, font)
	p, _ = syscall.UTF16PtrFromString(versionText)
	label, _, _ = procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(staticClass)), uintptr(unsafe.Pointer(p)), WS_CHILD|WS_VISIBLE, 24, 58, 410, 22, updateRestartWindow, 0, hInstance, 0)
	setFont(label, font)
	updateRestartCountdownLabel, _, _ = procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(staticClass)), 0, WS_CHILD|WS_VISIBLE, 24, 98, 410, 22, updateRestartWindow, 0, hInstance, 0)
	setFont(updateRestartCountdownLabel, font)
	p, _ = syscall.UTF16PtrFromString(restartText)
	button, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(buttonClass)), uintptr(unsafe.Pointer(p)), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_DEFPUSHBUTTON, 294, 142, 140, 32, updateRestartWindow, ID_UPDATE_RESTART, hInstance, 0)
	setFont(button, font)
	refreshRestartCountdownText()
	procSetTimer.Call(updateRestartWindow, updateRestartTimer, 1000, 0)
	procShowWindow.Call(updateRestartWindow, SW_SHOW)
	procUpdateWindow.Call(updateRestartWindow)
	procSetForegroundWindow.Call(updateRestartWindow)

	var m msg
	for updateRestartWindow != 0 {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	return nil
}

func refreshRestartCountdownText() {
	text := fmt.Sprintf("자동 재시작까지 %d초", updateRestartSeconds)
	if updateRestartLanguage == "en" {
		text = fmt.Sprintf("Restarting automatically in %d seconds", updateRestartSeconds)
	}
	setWindowText(updateRestartCountdownLabel, text)
}

func restartUpdatedApp() {
	if updateRestartWindow != 0 {
		procKillTimer.Call(updateRestartWindow, updateRestartTimer)
	}
	if updateRestartTarget != "" {
		_ = exec.Command(updateRestartTarget).Start()
	}
	if updateRestartWindow != 0 {
		procDestroyWindow.Call(updateRestartWindow)
	}
	updateRestartWindow = 0
	procPostQuitMessage.Call(0)
}

func updateCompleteWindowProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case WM_COMMAND:
		if uint16(wParam&0xffff) == ID_UPDATE_RESTART {
			restartUpdatedApp()
			return 0
		}
	case WM_TIMER:
		if wParam == updateRestartTimer {
			updateRestartSeconds--
			if updateRestartSeconds <= 0 {
				restartUpdatedApp()
			} else {
				refreshRestartCountdownText()
			}
			return 0
		}
	case WM_CLOSE:
		restartUpdatedApp()
		return 0
	case WM_DESTROY:
		if hwnd == updateRestartWindow {
			updateRestartWindow = 0
		}
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return r
}
