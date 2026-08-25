//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const singleInstanceMutexName = `Local\QuickAnydeskConnect`

var (
	procCreateMutexW = kernel32.NewProc("CreateMutexW")
	procCloseHandle  = kernel32.NewProc("CloseHandle")
	procFindWindowW  = user32.NewProc("FindWindowW")
)

func acquireSingleInstance() (uintptr, bool, error) {
	className, _ := syscall.UTF16PtrFromString("QuickAnydeskConnectTrayWindow")
	if hwnd, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(className)), 0); hwnd != 0 {
		return 0, false, nil
	}

	name, _ := syscall.UTF16PtrFromString(singleInstanceMutexName)
	handle, _, callErr := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if handle == 0 {
		return 0, false, fmt.Errorf("failed to create single-instance mutex: %v", callErr)
	}
	if errno, ok := callErr.(syscall.Errno); ok && errno == 183 { // ERROR_ALREADY_EXISTS
		procCloseHandle.Call(handle)
		return 0, false, nil
	}
	return handle, true, nil
}

func releaseSingleInstance(handle uintptr) {
	if handle != 0 {
		procCloseHandle.Call(handle)
	}
}
