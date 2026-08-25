//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

func removeStaleStartupRegistration() error {
	exists, matches, err := startupRegistrationMatchesCurrent()
	if err != nil || !exists || matches {
		return err
	}
	return unregisterStartup()
}

func startupRegistrationMatchesCurrent() (bool, bool, error) {
	keyPath, _ := syscall.UTF16PtrFromString(`Software\Microsoft\Windows\CurrentVersion\Run`)
	var hKey uintptr
	result, _, _ := procRegOpenKeyExW.Call(HKEY_CURRENT_USER, uintptr(unsafe.Pointer(keyPath)), 0, KEY_QUERY_VALUE, uintptr(unsafe.Pointer(&hKey)))
	if result == 2 { // ERROR_FILE_NOT_FOUND
		return false, false, nil
	}
	if result != 0 {
		return false, false, fmt.Errorf(currentMessages().errRegistryOpen, result)
	}
	defer procRegCloseKey.Call(hKey)

	valueName, _ := syscall.UTF16PtrFromString("QuickAnydeskConnect")
	var valueType, size uint32
	result, _, _ = procRegQueryValueExW.Call(hKey, uintptr(unsafe.Pointer(valueName)), 0, uintptr(unsafe.Pointer(&valueType)), 0, uintptr(unsafe.Pointer(&size)))
	if result == 2 {
		return false, false, nil
	}
	if result != 0 {
		return false, false, fmt.Errorf(currentMessages().errRegistryOpen, result)
	}
	if valueType != REG_SZ || size < 2 {
		return true, false, nil
	}

	buf := make([]uint16, (size+1)/2)
	result, _, _ = procRegQueryValueExW.Call(
		hKey,
		uintptr(unsafe.Pointer(valueName)),
		0,
		uintptr(unsafe.Pointer(&valueType)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if result != 0 {
		return true, false, fmt.Errorf(currentMessages().errRegistryOpen, result)
	}

	registered := executableFromStartupCommand(syscall.UTF16ToString(buf))
	if registered == "" {
		return true, false, nil
	}
	current, err := os.Executable()
	if err != nil {
		return true, false, err
	}
	registered, err = filepath.Abs(registered)
	if err != nil {
		return true, false, err
	}
	current, err = filepath.Abs(current)
	if err != nil {
		return true, false, err
	}
	return true, strings.EqualFold(filepath.Clean(registered), filepath.Clean(current)), nil
}

func executableFromStartupCommand(command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return ""
	}
	if command[0] == '"' {
		if end := strings.Index(command[1:], `"`); end >= 0 {
			return command[1 : end+1]
		}
		return ""
	}
	if fields := strings.Fields(command); len(fields) > 0 {
		return fields[0]
	}
	return ""
}
