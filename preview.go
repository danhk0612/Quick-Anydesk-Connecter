//go:build windows

package main

import (
	"fmt"
	"image"
	"syscall"
	"unsafe"
)

var (
	previewWindow   uintptr
	previewDecision bool
	previewClosed   bool
	previewBitmap   uintptr
)

func showImagePreviewConfirm(img image.Image) (bool, error) {
	if dialogActive {
		return false, errCancelled
	}
	dialogActive = true
	defer func() { dialogActive = false }()

	hbmp, imgW, imgH, err := createHBITMAPFromImage(img)
	if err != nil {
		return false, err
	}
	previewBitmap = hbmp
	defer func() {
		if previewBitmap != 0 {
			procDeleteObject.Call(previewBitmap)
			previewBitmap = 0
		}
	}()

	className := utf16Ptr("QuickAnydeskConnectImagePreview")
	title := utf16Ptr(currentMessages().imagePreviewTitle)
	hInstance, _, _ := procGetModuleHandleW.Call(0)
	wc := wndClassEx{cbSize: uint32(unsafe.Sizeof(wndClassEx{})), lpfnWndProc: syscall.NewCallback(previewWindowProc), hInstance: hInstance, hbrBackground: dialogBgBrush, lpszClassName: className}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	width := imgW + 70
	if width < 540 {
		width = 540
	}
	height := imgH + 185
	screenW, _, _ := procGetSystemMetrics.Call(0)
	screenH, _, _ := procGetSystemMetrics.Call(1)
	x := (int32(screenW) - int32(width)) / 2
	y := (int32(screenH) - int32(height)) / 2
	previewWindow, _, _ = procCreateWindowExW.Call(WS_EX_TOPMOST, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(title)), WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU, uintptr(x), uintptr(y), uintptr(width), uintptr(height), 0, 0, hInstance, 0)
	if previewWindow == 0 {
		return false, fmt.Errorf("%s", currentMessages().errDialogCreate)
	}

	font, _, _ := procGetStockObject.Call(DEFAULT_GUI_FONT)
	staticClass := utf16Ptr("STATIC")
	imageCtl, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(staticClass)), 0, WS_CHILD|WS_VISIBLE|SS_BITMAP, uintptr((width-imgW)/2), 24, uintptr(imgW), uintptr(imgH), previewWindow, 0, hInstance, 0)
	procSendMessageW.Call(imageCtl, STM_SETIMAGE, IMAGE_BITMAP, hbmp)
	text := utf16Ptr(currentMessages().imagePreviewQuestion)
	desc, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(staticClass)), uintptr(unsafe.Pointer(text)), WS_CHILD|WS_VISIBLE, 20, uintptr(imgH+48), uintptr(width-40), 24, previewWindow, 0, hInstance, 0)
	setFont(desc, font)
	buttonClass := utf16Ptr("BUTTON")
	analyze := utf16Ptr(currentMessages().analyzeButton)
	ignore := utf16Ptr(currentMessages().ignoreButton)
	analyzeBtn, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(buttonClass)), uintptr(unsafe.Pointer(analyze)), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_DEFPUSHBUTTON, uintptr(width-210), uintptr(imgH+88), 85, 30, previewWindow, ID_ANALYZE, hInstance, 0)
	ignoreBtn, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(buttonClass)), uintptr(unsafe.Pointer(ignore)), WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, uintptr(width-115), uintptr(imgH+88), 85, 30, previewWindow, ID_IGNORE, hInstance, 0)
	setFont(analyzeBtn, font)
	setFont(ignoreBtn, font)

	previewDecision = false
	previewClosed = false
	procShowWindow.Call(previewWindow, SW_SHOW)
	procUpdateWindow.Call(previewWindow)
	procSetForegroundWindow.Call(previewWindow)
	var m msg
	for previewWindow != 0 {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			previewClosed = true
			break
		}
		if m.message == WM_KEYDOWN && m.wParam == VK_ESCAPE {
			previewClosed = true
			procDestroyWindow.Call(previewWindow)
			previewWindow = 0
			continue
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	if previewClosed {
		return false, nil
	}
	return previewDecision, nil
}

func previewWindowProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case WM_CTLCOLORSTATIC:
		procSetBkMode.Call(wParam, TRANSPARENT)
		if dialogBgBrush != 0 {
			return dialogBgBrush
		}
		return 0
	case WM_COMMAND:
		switch uint16(wParam & 0xffff) {
		case ID_ANALYZE:
			previewDecision = true
			procDestroyWindow.Call(hwnd)
			previewWindow = 0
			return 0
		case ID_IGNORE:
			previewDecision = false
			procDestroyWindow.Call(hwnd)
			previewWindow = 0
			return 0
		}
	case WM_CLOSE:
		previewClosed = true
		procDestroyWindow.Call(hwnd)
		previewWindow = 0
		return 0
	case WM_DESTROY:
		previewWindow = 0
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return r
}
