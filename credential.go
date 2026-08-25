//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const openRouterCredentialTarget = "QuickAnydeskConnect/OpenRouter"

func saveOpenRouterAPIKey(key string) error {
	if key == "" {
		return fmt.Errorf("%s", currentMessages().errOpenRouterKeyEmpty)
	}
	target, _ := syscall.UTF16PtrFromString(openRouterCredentialTarget)
	username, _ := syscall.UTF16PtrFromString("Quick Anydesk Connect")
	blob := []byte(key)

	cred := credential{
		Type:               CRED_TYPE_GENERIC,
		TargetName:         target,
		CredentialBlobSize: uint32(len(blob)),
		Persist:            CRED_PERSIST_LOCAL_MACHINE,
		UserName:           username,
	}
	if len(blob) > 0 {
		cred.CredentialBlob = &blob[0]
	}

	r, _, err := procCredWriteW.Call(uintptr(unsafe.Pointer(&cred)), 0)
	if r == 0 {
		return fmt.Errorf(currentMessages().errCredentialSave, err)
	}
	return nil
}

func loadOpenRouterAPIKey() (string, error) {
	target, _ := syscall.UTF16PtrFromString(openRouterCredentialTarget)
	var pCred uintptr
	r, _, err := procCredReadW.Call(
		uintptr(unsafe.Pointer(target)),
		CRED_TYPE_GENERIC,
		0,
		uintptr(unsafe.Pointer(&pCred)),
	)
	if r == 0 {
		if errno, ok := err.(syscall.Errno); ok && errno == 1168 { // ERROR_NOT_FOUND
			return "", nil
		}
		return "", fmt.Errorf(currentMessages().errCredentialRead, err)
	}
	defer procCredFree.Call(pCred)

	cred := (*credential)(unsafe.Pointer(pCred))
	if cred.CredentialBlobSize == 0 || cred.CredentialBlob == nil {
		return "", nil
	}
	data := unsafe.Slice(cred.CredentialBlob, int(cred.CredentialBlobSize))
	return string(append([]byte(nil), data...)), nil
}
