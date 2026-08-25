package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

func main() {
	if len(os.Args) == 4 && os.Args[1] == "--apply-update" {
		if err := runApplyUpdate(os.Args[2], os.Args[3]); err != nil {
			showError(fmt.Sprintf("업데이트 적용에 실패했습니다.\n\n%v", err))
		}
		return
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	instanceHandle, acquired, err := acquireSingleInstance()
	if err != nil {
		showError(err.Error())
		return
	}
	if !acquired {
		return
	}
	defer releaseSingleInstance(instanceHandle)

	exePath, err := os.Executable()
	if err != nil {
		showError(currentMessages().errExecutablePath)
		return
	}
	exeDir = filepath.Dir(exePath)

	dialogBgBrush, _, _ = procGetSysColorBrush.Call(COLOR_BTNFACE)

	configPath, err = resolveConfigPath(exeDir)
	if err != nil {
		showError(err.Error())
		return
	}
	password, language, imageAnalysisEnabled, openRouterModel, err = loadOrCreateConfig(configPath)
	if err != nil {
		if !errors.Is(err, errCancelled) {
			showError(err.Error())
		}
		return
	}

	if _, err := findAnyDesk(); err != nil {
		showError(err.Error())
		return
	}

	if err := runTrayApp(); err != nil {
		showError(err.Error())
	}
}

func runTrayApp() error {
	className, _ := syscall.UTF16PtrFromString("QuickAnydeskConnectTrayWindow")
	windowName, _ := syscall.UTF16PtrFromString("Quick Anydesk Connect")
	hInstance, _, _ := procGetModuleHandleW.Call(0)

	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   syscall.NewCallback(mainWindowProc),
		hInstance:     hInstance,
		hbrBackground: dialogBgBrush,
		lpszClassName: className,
	}

	if r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		return fmt.Errorf("%s", currentMessages().errTrayClass)
	}

	mainWindow, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		WS_OVERLAPPED,
		0, 0, 0, 0,
		0, 0, hInstance, 0,
	)
	if mainWindow == 0 {
		return fmt.Errorf("%s", currentMessages().errTrayWindow)
	}

	procShowWindow.Call(mainWindow, SW_HIDE)

	if r, _, _ := procAddClipboardFormatListener.Call(mainWindow); r == 0 {
		return fmt.Errorf("%s", currentMessages().errClipboardWatch)
	}
	defer procRemoveClipboardFormatListener.Call(mainWindow)

	if err := addTrayIcon(); err != nil {
		return err
	}
	defer removeTrayIcon()

	var m msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	return nil
}
