package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

func askAnyDeskID() (string, error) {
	m := currentMessages()
	return showInputDialog(
		modeAnyDeskID,
		m.connectTitle,
		m.connectField,
		m.connectDescription,
		false,
		m.connectButton,
	)
}

func askPassword() (string, error) {
	m := currentMessages()
	return showInputDialog(
		modePassword,
		m.setupTitle,
		m.setupPassword,
		m.setupDescription,
		true,
		m.saveButton,
	)
}

func askNewPassword() (string, error) {
	m := currentMessages()
	return showInputDialog(modePassword, m.changePasswordTitle, m.setupPassword, m.changePasswordDescription, true, m.saveButton)
}

func askOpenRouterKey() (string, error) {
	m := currentMessages()
	return showInputDialog(modeOpenRouterKey, m.openRouterTitle, m.openRouterKeyLabel, m.openRouterKeyDescription, true, m.verifyButton)
}

func showInputDialog(mode dialogMode, title, fieldLabel, description string, passwordField bool, okText string) (string, error) {
	if dialogActive {
		return "", errCancelled
	}
	dialogActive = true
	defer func() { dialogActive = false }()

	className, _ := syscall.UTF16PtrFromString("QuickAnydeskConnectInputDialog")
	titleW, _ := syscall.UTF16PtrFromString(title)
	hInstance, _, _ := procGetModuleHandleW.Call(0)

	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   syscall.NewCallback(dialogWindowProc),
		hInstance:     hInstance,
		hbrBackground: dialogBgBrush,
		lpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	width := int32(460)
	height := int32(225)

	screenW, _, _ := procGetSystemMetrics.Call(0)
	screenH, _, _ := procGetSystemMetrics.Call(1)
	x := (int32(screenW) - width) / 2
	y := (int32(screenH) - height) / 2

	dialogWindow, _, _ = procCreateWindowExW.Call(
		WS_EX_TOPMOST,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(titleW)),
		WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU,
		uintptr(x), uintptr(y), uintptr(width), uintptr(height),
		0, 0, hInstance, 0,
	)
	if dialogWindow == 0 {
		return "", fmt.Errorf("%s", currentMessages().errDialogCreate)
	}

	font, _, _ := procGetStockObject.Call(DEFAULT_GUI_FONT)

	staticClass, _ := syscall.UTF16PtrFromString("STATIC")
	descW, _ := syscall.UTF16PtrFromString(description)
	labelW, _ := syscall.UTF16PtrFromString(fieldLabel)

	desc, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(staticClass)), uintptr(unsafe.Pointer(descW)),
		WS_CHILD|WS_VISIBLE, 24, 20, 400, 38, dialogWindow, 0, hInstance, 0,
	)
	setFont(desc, font)

	label, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(staticClass)), uintptr(unsafe.Pointer(labelW)),
		WS_CHILD|WS_VISIBLE, 24, 66, 400, 20, dialogWindow, 0, hInstance, 0,
	)
	setFont(label, font)

	editStyle := uintptr(WS_CHILD | WS_VISIBLE | WS_TABSTOP | WS_BORDER | ES_AUTOHSCROLL)
	if passwordField {
		editStyle |= ES_PASSWORD
	}

	editClass, _ := syscall.UTF16PtrFromString("EDIT")
	empty, _ := syscall.UTF16PtrFromString("")
	dialogEdit, _, _ = procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(editClass)), uintptr(unsafe.Pointer(empty)),
		editStyle, 24, 90, 400, 28, dialogWindow, ID_EDIT, hInstance, 0,
	)
	setFont(dialogEdit, font)

	buttonClass, _ := syscall.UTF16PtrFromString("BUTTON")
	okW, _ := syscall.UTF16PtrFromString(okText)
	cancelW, _ := syscall.UTF16PtrFromString(currentMessages().cancelButton)

	okBtn, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(buttonClass)), uintptr(unsafe.Pointer(okW)),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_DEFPUSHBUTTON,
		254, 140, 80, 30, dialogWindow, ID_OK, hInstance, 0,
	)
	setFont(okBtn, font)

	cancelBtn, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(buttonClass)), uintptr(unsafe.Pointer(cancelW)),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON,
		344, 140, 80, 30, dialogWindow, ID_CANCEL, hInstance, 0,
	)
	setFont(cancelBtn, font)

	dialogValue = ""
	dialogCancel = false
	dialogModeNow = mode

	procShowWindow.Call(dialogWindow, SW_SHOW)
	procUpdateWindow.Call(dialogWindow)
	procSetForegroundWindow.Call(dialogWindow)
	procSetFocus.Call(dialogEdit)

	var m msg
	for dialogWindow != 0 {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			dialogCancel = true
			break
		}

		if m.message == WM_KEYDOWN {
			switch m.wParam {
			case VK_RETURN:
				submitDialog()
				continue
			case VK_ESCAPE:
				dialogCancel = true
				procDestroyWindow.Call(dialogWindow)
				dialogWindow = 0
				continue
			}
		}

		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	if dialogCancel || dialogValue == "" {
		return "", errCancelled
	}

	return dialogValue, nil
}

func setFont(hwnd, font uintptr) {
	if hwnd != 0 && font != 0 {
		procSendMessageW.Call(hwnd, WM_SETFONT, font, 1)
	}
}

func dialogWindowProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case WM_CTLCOLORSTATIC:
		procSetBkMode.Call(wParam, TRANSPARENT)
		if dialogBgBrush != 0 {
			return dialogBgBrush
		}
		return 0

	case WM_COMMAND:
		id := uint16(wParam & 0xffff)
		switch id {
		case ID_OK:
			submitDialog()
			return 0
		case ID_CANCEL:
			dialogCancel = true
			procDestroyWindow.Call(hwnd)
			dialogWindow = 0
			return 0
		}

	case WM_CLOSE:
		dialogCancel = true
		procDestroyWindow.Call(hwnd)
		dialogWindow = 0
		return 0

	case WM_DESTROY:
		dialogWindow = 0
		return 0
	}

	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return r
}

func submitDialog() {
	raw := getWindowText(dialogEdit)

	switch dialogModeNow {
	case modeAnyDeskID:
		value := normalizeID(raw)
		if !isAnyDeskID(value) {
			messageBox(
				dialogWindow,
				currentMessages().invalidID,
				currentMessages().inputCheckTitle,
				MB_OK|MB_ICONERROR|MB_TOPMOST|MB_SETFOREGROUND,
			)
			procSetFocus.Call(dialogEdit)
			return
		}
		dialogValue = value

	case modePassword, modeOpenRouterKey:
		if raw == "" {
			messageBox(
				dialogWindow,
				func() string {
					if dialogModeNow == modeOpenRouterKey {
						return currentMessages().errOpenRouterKeyEmpty
					}
					return currentMessages().emptyPassword
				}(),
				currentMessages().inputCheckTitle,
				MB_OK|MB_ICONERROR|MB_TOPMOST|MB_SETFOREGROUND,
			)
			procSetFocus.Call(dialogEdit)
			return
		}
		dialogValue = raw
	}

	dialogCancel = false
	procDestroyWindow.Call(dialogWindow)
	dialogWindow = 0
}

func getWindowText(hwnd uintptr) string {
	n, _, _ := procGetWindowTextLenW.Call(hwnd)
	if n == 0 {
		return ""
	}

	buf := make([]uint16, n+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}

func showError(text string) {
	messageBox(0, text, "Quick Anydesk Connect", MB_OK|MB_ICONERROR|MB_TOPMOST|MB_SETFOREGROUND)
}

func messageBox(hwnd uintptr, text, title string, flags uintptr) {
	messageBoxResult(hwnd, text, title, flags)
}

func messageBoxResult(hwnd uintptr, text, title string, flags uintptr) uintptr {
	t, _ := syscall.UTF16PtrFromString(text)
	c, _ := syscall.UTF16PtrFromString(title)

	r, _, _ := procMessageBoxW.Call(
		hwnd,
		uintptr(unsafe.Pointer(t)),
		uintptr(unsafe.Pointer(c)),
		flags,
	)
	return r
}
