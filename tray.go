package main

import (
	"fmt"
	"image"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type clipboardImageResult struct {
	img     image.Image
	pngData []byte
	hash    [32]byte
	err     error
}

var (
	analysisResultMu      sync.Mutex
	clipboardImageMu      sync.Mutex
	clipboardImagePending clipboardImageResult
	lastClipboardImageAt  time.Time
)

func mainWindowProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case WM_CLIPBOARDUPDATE:
		handleClipboardUpdate()
		return 0

	case WM_IMAGE_ANALYSIS_RESULT:
		handleImageAnalysisResult()
		return 0

	case WM_CLIPBOARD_IMAGE_READY:
		handleClipboardImageReady()
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
	if dialogActive || imageAnalysisBusy || clipboardImageBusy {
		return
	}

	if text, err := readClipboardText(); err == nil {
		target := normalizeID(text)
		if isAnyDeskID(target) {
			confirmAndConnect(target)
			return
		}
	}
	if !imageAnalysisEnabled || !clipboardHasImage() {
		return
	}

	// Clipboard image decoding, resizing and PNG encoding can be expensive for
	// large screenshots/photos. Keep that work off the Win32 UI thread so tray
	// and window messages remain responsive.
	clipboardImageBusy = true
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		img, pngData, hash, err := readClipboardImage()
		clipboardImageMu.Lock()
		clipboardImagePending = clipboardImageResult{img: img, pngData: pngData, hash: hash, err: err}
		clipboardImageMu.Unlock()
		procPostMessageW.Call(mainWindow, WM_CLIPBOARD_IMAGE_READY, 0, 0)
	}()
}

func handleClipboardImageReady() {
	clipboardImageBusy = false

	clipboardImageMu.Lock()
	pending := clipboardImagePending
	clipboardImagePending = clipboardImageResult{}
	clipboardImageMu.Unlock()

	if pending.err != nil {
		showError(pending.err.Error())
		return
	}
	if pending.img == nil || len(pending.pngData) == 0 {
		showError(currentMessages().errClipboardImageRead)
		return
	}
	if hashEqual(pending.hash, lastClipboardImageHash) && time.Since(lastClipboardImageAt) < 2*time.Second {
		return
	}
	lastClipboardImageHash = pending.hash
	lastClipboardImageAt = time.Now()

	approved, err := showImagePreviewConfirm(pending.img)
	if err != nil {
		showError(err.Error())
		return
	}
	if !approved {
		return
	}

	imageAnalysisBusy = true
	go func(data []byte) {
		result, err := analyzeImageForAnyDesk(data)
		analysisResultMu.Lock()
		analysisResult = result
		analysisErr = err
		analysisResultMu.Unlock()
		procPostMessageW.Call(mainWindow, WM_IMAGE_ANALYSIS_RESULT, 0, 0)
	}(append([]byte(nil), pending.pngData...))
}

func handleImageAnalysisResult() {
	imageAnalysisBusy = false
	analysisResultMu.Lock()
	result, err := analysisResult, analysisErr
	analysisResult = ""
	analysisErr = nil
	analysisResultMu.Unlock()
	if err != nil {
		showImageAnalysisError(err)
		return
	}
	if !isAnyDeskID(result) {
		showImageAnalysisError(errOpenRouterInvalidResponse)
		return
	}
	confirmAndConnect(result)
}

func confirmAndConnect(target string) {
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

	m := currentMessages()
	ptr := func(v string) *uint16 { p, _ := syscall.UTF16PtrFromString(v); return p }

	koFlags, enFlags := uintptr(MF_STRING), uintptr(MF_STRING)
	if language == "ko" {
		koFlags |= MF_CHECKED
	} else {
		enFlags |= MF_CHECKED
	}
	imageFlags := uintptr(MF_STRING)
	if imageAnalysisEnabled {
		imageFlags |= MF_CHECKED
	}

	procAppendMenuW.Call(languageMenu, koFlags, ID_TRAY_LANGUAGE_KO, uintptr(unsafe.Pointer(ptr("한국어"))))
	procAppendMenuW.Call(languageMenu, enFlags, ID_TRAY_LANGUAGE_EN, uintptr(unsafe.Pointer(ptr("English"))))

	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_CONNECT, uintptr(unsafe.Pointer(ptr(m.trayConnect))))
	procAppendMenuW.Call(menu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(menu, imageFlags, ID_TRAY_IMAGE_ANALYSIS, uintptr(unsafe.Pointer(ptr(m.trayImageAnalysis))))
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_CHANGE_PASSWORD, uintptr(unsafe.Pointer(ptr(m.trayChangePassword))))
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_OPENROUTER, uintptr(unsafe.Pointer(ptr(m.trayOpenRouter))))
	procAppendMenuW.Call(menu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_STARTUP_ADD, uintptr(unsafe.Pointer(ptr(m.trayStartupAdd))))
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_STARTUP_REMOVE, uintptr(unsafe.Pointer(ptr(m.trayStartupRemove))))
	procAppendMenuW.Call(menu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(menu, MF_POPUP, languageMenu, uintptr(unsafe.Pointer(ptr(m.trayLanguage))))
	procAppendMenuW.Call(menu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_EXIT, uintptr(unsafe.Pointer(ptr(m.trayExit))))

	var pt point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procSetForegroundWindow.Call(hwnd)
	cmd, _, _ := procTrackPopupMenu.Call(menu, TPM_RIGHTBUTTON|TPM_RETURNCMD, uintptr(pt.x), uintptr(pt.y), 0, hwnd, 0)

	switch cmd {
	case ID_TRAY_CONNECT:
		startManualConnection()
	case ID_TRAY_IMAGE_ANALYSIS:
		if imageAnalysisEnabled {
			if err := setImageAnalysisEnabled(false); err != nil {
				showError(err.Error())
			}
		} else {
			if messageBoxResult(mainWindow, currentMessages().imagePrivacyNotice, "Quick Anydesk Connect", MB_OKCANCEL|MB_ICONQUESTION|MB_TOPMOST|MB_SETFOREGROUND) != IDOK {
				break
			}
			if ensureOpenRouterReady() {
				if err := setImageAnalysisEnabled(true); err != nil {
					showError(err.Error())
				}
			}
		}
	case ID_TRAY_CHANGE_PASSWORD:
		newPassword, err := askNewPassword()
		if err == nil {
			old := password
			password = newPassword
			if err := persistConfig(); err != nil {
				password = old
				showError(err.Error())
			}
		}
	case ID_TRAY_OPENROUTER:
		configureOpenRouter()
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

	result, _, _ := procRegOpenKeyExW.Call(
		HKEY_CURRENT_USER,
		uintptr(unsafe.Pointer(keyPath)),
		0,
		KEY_SET_VALUE,
		uintptr(unsafe.Pointer(&hKey)),
	)
	if result != 0 {
		return fmt.Errorf(currentMessages().errRegistryOpen, result)
	}
	defer procRegCloseKey.Call(hKey)

	valueName, _ := syscall.UTF16PtrFromString("QuickAnydeskConnect")

	// 경로에 공백이 있을 수 있으므로 Run 값에는 따옴표를 포함해서 저장한다.
	value := `"` + exePath + `"`
	valueUTF16, err := syscall.UTF16FromString(value)
	if err != nil {
		return fmt.Errorf(currentMessages().errPathConvert, err)
	}

	result, _, _ = procRegSetValueExW.Call(
		hKey,
		uintptr(unsafe.Pointer(valueName)),
		0,
		REG_SZ,
		uintptr(unsafe.Pointer(&valueUTF16[0])),
		uintptr(len(valueUTF16)*2),
	)
	if result != 0 {
		return fmt.Errorf(currentMessages().errRegistrySet, result)
	}

	return nil
}

func unregisterStartup() error {
	keyPath, _ := syscall.UTF16PtrFromString(`Software\Microsoft\Windows\CurrentVersion\Run`)
	var hKey uintptr

	result, _, _ := procRegOpenKeyExW.Call(
		HKEY_CURRENT_USER,
		uintptr(unsafe.Pointer(keyPath)),
		0,
		KEY_SET_VALUE,
		uintptr(unsafe.Pointer(&hKey)),
	)
	if result != 0 {
		return fmt.Errorf(currentMessages().errRegistryOpen, result)
	}
	defer procRegCloseKey.Call(hKey)

	valueName, _ := syscall.UTF16PtrFromString("QuickAnydeskConnect")
	result, _, _ = procRegDeleteValueW.Call(
		hKey,
		uintptr(unsafe.Pointer(valueName)),
	)

	// ERROR_FILE_NOT_FOUND(2)는 이미 제거된 상태로 간주한다.
	if result != 0 && result != 2 {
		return fmt.Errorf(currentMessages().errRegistryDelete, result)
	}

	return nil
}

func addTrayIcon() error {
	icon := loadAppIcon()

	trayIcon = notifyIconData{
		cbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		hWnd:             mainWindow,
		uID:              1,
		uFlags:           NIF_MESSAGE | NIF_ICON | NIF_TIP,
		uCallbackMessage: WM_TRAYICON,
		hIcon:            icon,
	}

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
	// app.ico는 go:embed로 EXE 내부에 포함된다.
	// ICO 디렉터리에서 32x32에 가장 가까운 이미지 엔트리를 선택한 뒤
	// CreateIconFromResourceEx로 HICON을 생성한다.
	if len(embeddedIcon) >= 6 {
		count := int(uint16(embeddedIcon[4]) | uint16(embeddedIcon[5])<<8)
		bestOffset := -1
		bestSize := 0
		bestScore := 1 << 30

		for i := 0; i < count; i++ {
			base := 6 + i*16
			if base+16 > len(embeddedIcon) {
				break
			}

			w := int(embeddedIcon[base])
			h := int(embeddedIcon[base+1])
			if w == 0 {
				w = 256
			}
			if h == 0 {
				h = 256
			}

			size := int(uint32(embeddedIcon[base+8]) |
				uint32(embeddedIcon[base+9])<<8 |
				uint32(embeddedIcon[base+10])<<16 |
				uint32(embeddedIcon[base+11])<<24)
			offset := int(uint32(embeddedIcon[base+12]) |
				uint32(embeddedIcon[base+13])<<8 |
				uint32(embeddedIcon[base+14])<<16 |
				uint32(embeddedIcon[base+15])<<24)

			if size <= 0 || offset < 0 || offset+size > len(embeddedIcon) {
				continue
			}

			// 트레이 기본 크기인 32x32에 가까운 엔트리 우선
			score := absInt(w-32) + absInt(h-32)
			if score < bestScore {
				bestScore = score
				bestOffset = offset
				bestSize = size
			}
		}

		if bestOffset >= 0 && bestSize > 0 {
			data := embeddedIcon[bestOffset : bestOffset+bestSize]
			icon, _, _ := procCreateIconFromResourceEx.Call(
				uintptr(unsafe.Pointer(&data[0])),
				uintptr(len(data)),
				1, // fIcon = TRUE
				0x00030000,
				32,
				32,
				LR_DEFAULTCOLOR,
			)
			if icon != 0 {
				return icon
			}
		}
	}

	// 최종 fallback
	icon, _, _ := procLoadIconW.Call(0, 32512) // IDI_APPLICATION
	return icon
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
