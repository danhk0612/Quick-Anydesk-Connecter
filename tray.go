package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func mainWindowProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case WM_CLIPBOARDUPDATE:
		handleClipboardUpdate()
		return 0

	case WM_TRAYICON:
		switch uint32(lParam) {
		case WM_RBUTTONUP, WM_CONTEXTMENU:
			showTrayMenu(hwnd)
		case WM_LBUTTONDBLCLK:
			startManualConnection()
		}
		return 0

	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}

	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return r
}

func handleClipboardUpdate() {
	if dialogActive {
		return
	}

	text, err := readClipboardText()
	if err != nil {
		return
	}

	target := normalizeID(text)
	if !isAnyDeskID(target) {
		return
	}

	question := fmt.Sprintf(currentMessages().confirmConnect, target)
	procSetForegroundWindow.Call(mainWindow)
	flags := uintptr(MB_YESNO | MB_ICONQUESTION | MB_SETFOREGROUND | MB_TOPMOST)
	if messageBoxResult(mainWindow, question, currentMessages().connectTitle, flags) != IDYES {
		return
	}

	connectTarget(target)
}

func startManualConnection() {
	if dialogActive {
		return
	}

	target, err := askAnyDeskID()
	if err != nil {
		return
	}

	connectTarget(target)
}

func connectTarget(target string) {
	anyDeskPath, err := findAnyDesk()
	if err != nil {
		showError(err.Error())
		return
	}

	if err := ensureAnyDeskRunning(anyDeskPath); err != nil {
		showError(fmt.Sprintf(currentMessages().errAnyDeskStart, err.Error()))
		return
	}

	if err := connect(anyDeskPath, target, password); err != nil {
		showError(fmt.Sprintf(currentMessages().errAnyDeskRun, err.Error()))
	}
}

func showTrayMenu(hwnd uintptr) {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)

	languageMenu, _, _ := procCreatePopupMenu.Call()
	if languageMenu == 0 {
		return
	}
	defer procDestroyMenu.Call(languageMenu)

	m := currentMessages()
	connectText, _ := syscall.UTF16PtrFromString(m.trayConnect)
	startupAddText, _ := syscall.UTF16PtrFromString(m.trayStartupAdd)
	startupRemoveText, _ := syscall.UTF16PtrFromString(m.trayStartupRemove)
	languageText, _ := syscall.UTF16PtrFromString(m.trayLanguage)
	koreanText, _ := syscall.UTF16PtrFromString("한국어")
	englishText, _ := syscall.UTF16PtrFromString("English")
	exitText, _ := syscall.UTF16PtrFromString(m.trayExit)

	koFlags := uintptr(MF_STRING)
	enFlags := uintptr(MF_STRING)
	if language == "ko" {
		koFlags |= MF_CHECKED
	} else {
		enFlags |= MF_CHECKED
	}

	procAppendMenuW.Call(languageMenu, koFlags, ID_TRAY_LANGUAGE_KO, uintptr(unsafe.Pointer(koreanText)))
	procAppendMenuW.Call(languageMenu, enFlags, ID_TRAY_LANGUAGE_EN, uintptr(unsafe.Pointer(englishText)))

	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_CONNECT, uintptr(unsafe.Pointer(connectText)))
	procAppendMenuW.Call(menu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_STARTUP_ADD, uintptr(unsafe.Pointer(startupAddText)))
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_STARTUP_REMOVE, uintptr(unsafe.Pointer(startupRemoveText)))
	procAppendMenuW.Call(menu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(menu, MF_POPUP, languageMenu, uintptr(unsafe.Pointer(languageText)))
	procAppendMenuW.Call(menu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_EXIT, uintptr(unsafe.Pointer(exitText)))

	var pt point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procSetForegroundWindow.Call(hwnd)

	cmd, _, _ := procTrackPopupMenu.Call(
		menu,
		TPM_RIGHTBUTTON|TPM_RETURNCMD,
		uintptr(pt.x),
		uintptr(pt.y),
		0,
		hwnd,
		0,
	)

	switch cmd {
	case ID_TRAY_CONNECT:
		startManualConnection()
	case ID_TRAY_STARTUP_ADD:
		if err := registerStartup(); err != nil {
			messageBox(mainWindow, fmt.Sprintf(currentMessages().startupAddFailure, err.Error()), "Quick Anydesk Connect", MB_OK|MB_ICONERROR|MB_TOPMOST|MB_SETFOREGROUND)
		} else {
			messageBox(mainWindow, currentMessages().startupAddSuccess, "Quick Anydesk Connect", MB_OK|MB_TOPMOST|MB_SETFOREGROUND)
		}
	case ID_TRAY_STARTUP_REMOVE:
		if err := unregisterStartup(); err != nil {
			messageBox(mainWindow, fmt.Sprintf(currentMessages().startupRemoveFailure, err.Error()), "Quick Anydesk Connect", MB_OK|MB_ICONERROR|MB_TOPMOST|MB_SETFOREGROUND)
		} else {
			messageBox(mainWindow, currentMessages().startupRemoveSuccess, "Quick Anydesk Connect", MB_OK|MB_TOPMOST|MB_SETFOREGROUND)
		}
	case ID_TRAY_LANGUAGE_KO:
		setLanguage("ko")
	case ID_TRAY_LANGUAGE_EN:
		setLanguage("en")
	case ID_TRAY_EXIT:
		procDestroyWindow.Call(hwnd)
	}
}

func registerStartup() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf(currentMessages().errExecutablePath+": %w", err)
	}

	keyPath, _ := syscall.UTF16PtrFromString(`Software\Microsoft\Windows\CurrentVersion\Run`)
	var hKey uintptr
	result, _, _ := procRegOpenKeyExW.Call(HKEY_CURRENT_USER, uintptr(unsafe.Pointer(keyPath)), 0, KEY_SET_VALUE, uintptr(unsafe.Pointer(&hKey)))
	if result != 0 {
		return fmt.Errorf(currentMessages().errRegistryOpen, result)
	}
	defer procRegCloseKey.Call(hKey)

	valueName, _ := syscall.UTF16PtrFromString("QuickAnydeskConnect")
	value := `"` + exePath + `"`
	valueUTF16, err := syscall.UTF16FromString(value)
	if err != nil {
		return fmt.Errorf(currentMessages().errPathConvert, err)
	}

	result, _, _ = procRegSetValueExW.Call(hKey, uintptr(unsafe.Pointer(valueName)), 0, REG_SZ, uintptr(unsafe.Pointer(&valueUTF16[0])), uintptr(len(valueUTF16)*2))
	if result != 0 {
		return fmt.Errorf(currentMessages().errRegistrySet, result)
	}
	return nil
}

func unregisterStartup() error {
	keyPath, _ := syscall.UTF16PtrFromString(`Software\Microsoft\Windows\CurrentVersion\Run`)
	var hKey uintptr
	result, _, _ := procRegOpenKeyExW.Call(HKEY_CURRENT_USER, uintptr(unsafe.Pointer(keyPath)), 0, KEY_SET_VALUE, uintptr(unsafe.Pointer(&hKey)))
	if result != 0 {
		return fmt.Errorf(currentMessages().errRegistryOpen, result)
	}
	defer procRegCloseKey.Call(hKey)

	valueName, _ := syscall.UTF16PtrFromString("QuickAnydeskConnect")
	result, _, _ = procRegDeleteValueW.Call(hKey, uintptr(unsafe.Pointer(valueName)))
	if result != 0 && result != 2 {
		return fmt.Errorf(currentMessages().errRegistryDelete, result)
	}
	return nil
}

func addTrayIcon() error {
	icon := loadAppIcon()
	trayIcon = notifyIconData{cbSize: uint32(unsafe.Sizeof(notifyIconData{})), hWnd: mainWindow, uID: 1, uFlags: NIF_MESSAGE | NIF_ICON | NIF_TIP, uCallbackMessage: WM_TRAYICON, hIcon: icon}
	tip := syscall.StringToUTF16("Quick Anydesk Connect")
	copy(trayIcon.szTip[:], tip)
	if r, _, _ := procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&trayIcon))); r == 0 {
		return fmt.Errorf("%s", currentMessages().errTrayIcon)
	}
	return nil
}

func removeTrayIcon() {
	if trayIcon.hWnd != 0 {
		procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&trayIcon)))
	}
}

func loadAppIcon() uintptr {
	if len(embeddedIcon) >= 6 {
		count := int(uint16(embeddedIcon[4]) | uint16(embeddedIcon[5])<<8)
		bestOffset := -1
		bestSize := 0
		bestScore := 1 << 30
		for i := 0; i < count; i++ {
			base := 6 + i*16
			if base+16 > len(embeddedIcon) { break }
			w := int(embeddedIcon[base]); h := int(embeddedIcon[base+1])
			if w == 0 { w = 256 }; if h == 0 { h = 256 }
			size := int(uint32(embeddedIcon[base+8]) | uint32(embeddedIcon[base+9])<<8 | uint32(embeddedIcon[base+10])<<16 | uint32(embeddedIcon[base+11])<<24)
			offset := int(uint32(embeddedIcon[base+12]) | uint32(embeddedIcon[base+13])<<8 | uint32(embeddedIcon[base+14])<<16 | uint32(embeddedIcon[base+15])<<24)
			if size <= 0 || offset < 0 || offset+size > len(embeddedIcon) { continue }
			score := absInt(w-32) + absInt(h-32)
			if score < bestScore { bestScore = score; bestOffset = offset; bestSize = size }
		}
		if bestOffset >= 0 && bestSize > 0 {
			data := embeddedIcon[bestOffset : bestOffset+bestSize]
			icon, _, _ := procCreateIconFromResourceEx.Call(uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)), 1, 0x00030000, 32, 32, LR_DEFAULTCOLOR)
			if icon != 0 { return icon }
		}
	}
	icon, _, _ := procLoadIconW.Call(0, 32512)
	return icon
}

func absInt(v int) int {
	if v < 0 { return -v }
	return v
}
