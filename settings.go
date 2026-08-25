//go:build windows

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const (
	settingsDirName        = "QuickAnydeskConnect"
	settingsBackupFormat   = "QuickAnydeskConnectBackup"
	settingsBackupVersion  = 1
	settingsBackupExt      = "qacbackup"
	ID_TRAY_SETTINGS_RESET = 2010
	ID_TRAY_SETTINGS_BACKUP = 2011
	ID_TRAY_SETTINGS_RESTORE = 2012

	OFN_OVERWRITEPROMPT = 0x00000002
	OFN_NOCHANGEDIR     = 0x00000008
	OFN_PATHMUSTEXIST   = 0x00000800
	OFN_FILEMUSTEXIST   = 0x00001000
	OFN_EXPLORER        = 0x00080000
)

var (
	comdlg32                 = syscall.NewLazyDLL("comdlg32.dll")
	procGetOpenFileNameW     = comdlg32.NewProc("GetOpenFileNameW")
	procGetSaveFileNameW     = comdlg32.NewProc("GetSaveFileNameW")
	procCommDlgExtendedError = comdlg32.NewProc("CommDlgExtendedError")
	procCredDeleteW          = advapi32.NewProc("CredDeleteW")
)

type openFilenameW struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        uintptr
	dwReserved        uint32
	flagsEx           uint32
}

type settingsBackup struct {
	Format           string `json:"format"`
	Version          int    `json:"version"`
	Config           string `json:"config"`
	OpenRouterAPIKey string `json:"openrouter_api_key"`
}

type settingsMessages struct {
	backupMenu, restoreMenu, resetMenu string
	backupTitle, restoreTitle, title   string
	backupWarning, backupSuccess       string
	restoreConfirm, restoreSuccess     string
	resetConfirm, resetSuccess         string
	backupFailed, restoreFailed        string
	resetFailed, invalidBackup         string
}

func currentSettingsMessages() settingsMessages {
	if language == "en" {
		return settingsMessages{
			backupMenu: "Back Up Settings", restoreMenu: "Restore Settings", resetMenu: "Reset Settings",
			backupTitle: "Save Quick Anydesk Connect settings backup", restoreTitle: "Open Quick Anydesk Connect settings backup", title: "Quick Anydesk Connect Settings",
			backupWarning: "The backup file contains the AnyDesk default password and OpenRouter API Key in readable form.\n\nStore the backup file securely. Continue?",
			backupSuccess: "Settings backup completed.",
			restoreConfirm: "Restore settings from this backup?\n\nThe current AnyDesk password, language, image-analysis setting, OpenRouter model, and API Key will be replaced.",
			restoreSuccess: "Settings were restored successfully.",
			resetConfirm: "Reset all application settings?\n\nThe AnyDesk password, language, image-analysis setting, OpenRouter model, and API Key will be reset. The startup registration is not changed.\n\nYou will be asked to enter a new default AnyDesk password.",
			resetSuccess: "Settings were reset successfully.",
			backupFailed: "Failed to back up settings.\n\n%v", restoreFailed: "Failed to restore settings.\n\n%v", resetFailed: "Failed to reset settings.\n\n%v",
			invalidBackup: "The selected file is not a valid Quick Anydesk Connect settings backup.",
		}
	}
	return settingsMessages{
		backupMenu: "설정 백업", restoreMenu: "설정 복원", resetMenu: "설정 초기화",
		backupTitle: "Quick Anydesk Connect 설정 백업 저장", restoreTitle: "Quick Anydesk Connect 설정 백업 열기", title: "Quick Anydesk Connect 설정",
		backupWarning: "백업 파일에는 AnyDesk 기본 암호와 OpenRouter API Key가 읽을 수 있는 형태로 포함됩니다.\n\n백업 파일을 안전하게 보관해주세요. 계속하시겠습니까?",
		backupSuccess: "설정 백업이 완료되었습니다.",
		restoreConfirm: "이 백업 파일의 설정을 복원하시겠습니까?\n\n현재 AnyDesk 암호, 언어, 이미지 분석 설정, OpenRouter 모델 및 API Key가 교체됩니다.",
		restoreSuccess: "설정 복원이 완료되었습니다.",
		resetConfirm: "모든 앱 설정을 초기화하시겠습니까?\n\nAnyDesk 암호, 언어, 이미지 분석 설정, OpenRouter 모델 및 API Key가 초기화됩니다. 시작프로그램 등록 상태는 변경하지 않습니다.\n\n초기화 후 새 기본 AnyDesk 암호를 입력하게 됩니다.",
		resetSuccess: "설정 초기화가 완료되었습니다.",
		backupFailed: "설정 백업에 실패했습니다.\n\n%v", restoreFailed: "설정 복원에 실패했습니다.\n\n%v", resetFailed: "설정 초기화에 실패했습니다.\n\n%v",
		invalidBackup: "선택한 파일은 올바른 Quick Anydesk Connect 설정 백업 파일이 아닙니다.",
	}
}

func resolveConfigPath(executableDir string) (string, error) {
	base := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
	if base == "" {
		var err error
		base, err = os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("could not determine user settings directory: %w", err)
		}
	}

	dir := filepath.Join(base, settingsDirName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("could not create user settings directory: %w", err)
	}
	target := filepath.Join(dir, "config.ini")
	marker := filepath.Join(dir, ".legacy-config-migration-v1")

	if _, err := os.Stat(target); err == nil {
		return target, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	if _, err := os.Stat(marker); os.IsNotExist(err) {
		legacy := filepath.Join(executableDir, "config.ini")
		if _, err := os.Stat(legacy); err == nil {
			if _, _, _, _, err := readConfig(legacy); err == nil {
				data, readErr := os.ReadFile(legacy)
				if readErr != nil {
					return "", readErr
				}
				if writeErr := os.WriteFile(target, data, 0600); writeErr != nil {
					return "", writeErr
				}
				_ = os.Remove(legacy)
			}
		}
		_ = os.WriteFile(marker, []byte("migrated\r\n"), 0600)
	}

	return target, nil
}

func backupSettings() error {
	m := currentSettingsMessages()
	if messageBoxResult(mainWindow, m.backupWarning, m.title, MB_OKCANCEL|MB_ICONQUESTION|MB_TOPMOST|MB_SETFOREGROUND) != IDOK {
		return errCancelled
	}

	path, err := chooseSettingsBackupPath(true, m.backupTitle)
	if err != nil {
		return err
	}
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	apiKey, err := loadOpenRouterAPIKey()
	if err != nil {
		return err
	}
	backup := settingsBackup{Format: settingsBackupFormat, Version: settingsBackupVersion, Config: string(configData), OpenRouterAPIKey: apiKey}
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\r', '\n')
	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}
	messageBox(mainWindow, m.backupSuccess, m.title, MB_OK|MB_TOPMOST|MB_SETFOREGROUND)
	return nil
}

func restoreSettings() error {
	m := currentSettingsMessages()
	path, err := chooseSettingsBackupPath(false, m.restoreTitle)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var backup settingsBackup
	if err := json.Unmarshal(data, &backup); err != nil || backup.Format != settingsBackupFormat || backup.Version != settingsBackupVersion || strings.TrimSpace(backup.Config) == "" {
		return errors.New(m.invalidBackup)
	}

	p, lang, imageEnabled, model, err := validateBackupConfig(backup.Config)
	if err != nil {
		return fmt.Errorf("%s\n\n%v", m.invalidBackup, err)
	}
	if messageBoxResult(mainWindow, m.restoreConfirm, m.title, MB_OKCANCEL|MB_ICONQUESTION|MB_TOPMOST|MB_SETFOREGROUND) != IDOK {
		return errCancelled
	}

	oldConfig, oldConfigErr := os.ReadFile(configPath)
	oldKey, oldKeyErr := loadOpenRouterAPIKey()
	if oldKeyErr != nil {
		return oldKeyErr
	}

	if backup.OpenRouterAPIKey == "" {
		if err := deleteOpenRouterAPIKey(); err != nil {
			return err
		}
	} else if err := saveOpenRouterAPIKey(backup.OpenRouterAPIKey); err != nil {
		return err
	}

	if err := saveConfig(configPath, p, lang, imageEnabled, model); err != nil {
		if oldKey == "" {
			_ = deleteOpenRouterAPIKey()
		} else {
			_ = saveOpenRouterAPIKey(oldKey)
		}
		if oldConfigErr == nil {
			_ = os.WriteFile(configPath, oldConfig, 0600)
		}
		return err
	}

	password = p
	language = lang
	imageAnalysisEnabled = imageEnabled
	openRouterModel = model
	messageBox(mainWindow, m.restoreSuccess, m.title, MB_OK|MB_TOPMOST|MB_SETFOREGROUND)
	return nil
}

func resetSettings() error {
	m := currentSettingsMessages()
	if messageBoxResult(mainWindow, m.resetConfirm, m.title, MB_OKCANCEL|MB_ICONQUESTION|MB_TOPMOST|MB_SETFOREGROUND) != IDOK {
		return errCancelled
	}

	oldConfig, oldConfigErr := os.ReadFile(configPath)
	oldKey, oldKeyErr := loadOpenRouterAPIKey()
	if oldKeyErr != nil {
		return oldKeyErr
	}

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := deleteOpenRouterAPIKey(); err != nil {
		if oldConfigErr == nil {
			_ = os.WriteFile(configPath, oldConfig, 0600)
		}
		return err
	}

	p, lang, imageEnabled, model, err := loadOrCreateConfig(configPath)
	if err != nil {
		if oldConfigErr == nil {
			_ = os.WriteFile(configPath, oldConfig, 0600)
		}
		if oldKey != "" {
			_ = saveOpenRouterAPIKey(oldKey)
		}
		return err
	}
	password = p
	language = lang
	imageAnalysisEnabled = imageEnabled
	openRouterModel = model
	messageBox(mainWindow, currentSettingsMessages().resetSuccess, currentSettingsMessages().title, MB_OK|MB_TOPMOST|MB_SETFOREGROUND)
	return nil
}

func validateBackupConfig(content string) (string, string, bool, string, error) {
	dir := filepath.Dir(configPath)
	f, err := os.CreateTemp(dir, "qac-restore-*.ini")
	if err != nil {
		return "", "ko", false, defaultOpenRouterModel, err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		return "", "ko", false, defaultOpenRouterModel, err
	}
	if err := f.Close(); err != nil {
		return "", "ko", false, defaultOpenRouterModel, err
	}
	return readConfig(name)
}

func deleteOpenRouterAPIKey() error {
	target, _ := syscall.UTF16PtrFromString(openRouterCredentialTarget)
	r, _, err := procCredDeleteW.Call(uintptr(unsafe.Pointer(target)), CRED_TYPE_GENERIC, 0)
	if r != 0 {
		return nil
	}
	if errno, ok := err.(syscall.Errno); ok && errno == 1168 {
		return nil
	}
	return fmt.Errorf("could not remove OpenRouter credential: %v", err)
}

func chooseSettingsBackupPath(save bool, title string) (string, error) {
	var fileBuf [32768]uint16
	if save {
		copy(fileBuf[:], syscall.StringToUTF16("QuickAnydeskConnect-backup.qacbackup"))
	}
	filter := utf16NULTerminated("Quick Anydesk Connect Backup (*.qacbackup)\x00*.qacbackup\x00All Files (*.*)\x00*.*\x00")
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	defExt, _ := syscall.UTF16PtrFromString(settingsBackupExt)
	of := openFilenameW{
		lStructSize: uint32(unsafe.Sizeof(openFilenameW{})),
		hwndOwner: mainWindow,
		lpstrFilter: &filter[0],
		nFilterIndex: 1,
		lpstrFile: &fileBuf[0],
		nMaxFile: uint32(len(fileBuf)),
		lpstrTitle: titlePtr,
		lpstrDefExt: defExt,
		flags: OFN_EXPLORER | OFN_NOCHANGEDIR | OFN_PATHMUSTEXIST,
	}
	if save {
		of.flags |= OFN_OVERWRITEPROMPT
	} else {
		of.flags |= OFN_FILEMUSTEXIST
	}
	var r uintptr
	if save {
		r, _, _ = procGetSaveFileNameW.Call(uintptr(unsafe.Pointer(&of)))
	} else {
		r, _, _ = procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&of)))
	}
	if r == 0 {
		code, _, _ := procCommDlgExtendedError.Call()
		if code == 0 {
			return "", errCancelled
		}
		return "", fmt.Errorf("file dialog error: 0x%X", code)
	}
	return syscall.UTF16ToString(fileBuf[:]), nil
}

func utf16NULTerminated(s string) []uint16 {
	v := utf16.Encode([]rune(s))
	return append(v, 0)
}
