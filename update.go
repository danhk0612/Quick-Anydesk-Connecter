//go:build windows

package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	appVersion          = "1.3.1"
	latestReleaseAPI    = "https://api.github.com/repos/danhk0612/Quick-Anydesk-Connecter/releases/latest"
	updateExeAssetName  = "QuickAnydeskConnect.exe"
	updateHashAssetName = "QuickAnydeskConnect.exe.sha256"
)

type githubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type githubRelease struct {
	TagName string               `json:"tag_name"`
	HTMLURL string               `json:"html_url"`
	Assets  []githubReleaseAsset `json:"assets"`
}

type updateCheckResult struct {
	release githubRelease
	err     error
}

type updateDownloadResult struct {
	updaterPath string
	newExePath  string
	targetPath  string
	err         error
}

type updaterMessages struct {
	title, latest, available, downloading                   string
	checkFailure, downloadFailure, applyFailure             string
	invalidRelease, assetMissing, hashInvalid, hashMismatch string
}

func currentUpdaterMessages() updaterMessages {
	if language == "en" {
		return updaterMessages{
			title: "Quick Anydesk Connect Update",
			latest: "You are using the latest version.\n\nCurrent version: %s\nRunning EXE: %s",
			available: "A new version is available.\n\nCurrent version: %s\nLatest version: %s\nRunning EXE: %s\n\nUpdate now?",
			downloading: "Downloading the update.\n\nThe app will restart automatically when it is ready.",
			checkFailure: "Failed to check for updates.\n\n%v",
			downloadFailure: "Failed to download or verify the update.\n\n%v",
			applyFailure: "Failed to start the update installer.\n\n%v",
			invalidRelease: "Could not determine the latest version from GitHub Releases.",
			assetMissing: "The Release does not contain the update EXE or SHA-256 file.",
			hashInvalid: "The Release SHA-256 file is invalid.",
			hashMismatch: "The downloaded EXE SHA-256 does not match the Release value.",
		}
	}
	return updaterMessages{
		title: "Quick Anydesk Connect 업데이트",
		latest: "현재 최신 버전을 사용 중입니다.\n\n현재 버전: %s\n실행 중인 EXE: %s",
		available: "새 버전이 있습니다.\n\n현재 버전: %s\n최신 버전: %s\n실행 중인 EXE: %s\n\n지금 업데이트하시겠습니까?",
		downloading: "업데이트 파일을 다운로드합니다.\n\n완료되면 프로그램이 자동으로 재시작됩니다.",
		checkFailure: "업데이트 확인에 실패했습니다.\n\n%v",
		downloadFailure: "업데이트 다운로드 또는 검증에 실패했습니다.\n\n%v",
		applyFailure: "업데이트 적용 프로그램을 시작하지 못했습니다.\n\n%v",
		invalidRelease: "GitHub Release에서 최신 버전 정보를 확인할 수 없습니다.",
		assetMissing: "Release에 업데이트용 EXE 또는 SHA-256 파일이 없습니다.",
		hashInvalid: "Release의 SHA-256 파일 형식이 올바르지 않습니다.",
		hashMismatch: "다운로드한 EXE의 SHA-256이 Release 값과 일치하지 않습니다.",
	}
}

func trayStartupText() string {
	if language == "en" {
		return "Run at Startup"
	}
	return "시작 프로그램"
}

func trayCheckUpdateText() string {
	if language == "en" {
		return "Check for Updates"
	}
	return "업데이트 확인"
}

var (
	updateMu              sync.Mutex
	updateCheckBusy       bool
	updateDownloadBusy    bool
	pendingUpdateCheck    updateCheckResult
	pendingUpdateDownload updateDownloadResult
)

func startUpdateCheck() {
	updateMu.Lock()
	if updateCheckBusy || updateDownloadBusy {
		updateMu.Unlock()
		return
	}
	updateCheckBusy = true
	updateMu.Unlock()
	go func() {
		release, err := fetchLatestRelease()
		updateMu.Lock()
		pendingUpdateCheck = updateCheckResult{release: release, err: err}
		updateMu.Unlock()
		procPostMessageW.Call(mainWindow, WM_UPDATE_CHECK_RESULT, 0, 0)
	}()
}

func handleUpdateCheckResult() {
	updateMu.Lock()
	result := pendingUpdateCheck
	pendingUpdateCheck = updateCheckResult{}
	updateCheckBusy = false
	updateMu.Unlock()
	m := currentUpdaterMessages()
	if result.err != nil {
		showError(fmt.Sprintf(m.checkFailure, result.err))
		return
	}
	latest := strings.TrimPrefix(strings.TrimSpace(result.release.TagName), "v")
	if latest == "" {
		showError(m.invalidRelease)
		return
	}
	runningPath := currentExecutableDisplayPath()
	if compareVersions(latest, appVersion) <= 0 {
		messageBox(mainWindow, fmt.Sprintf(m.latest, appVersion, runningPath), m.title, MB_OK|MB_TOPMOST|MB_SETFOREGROUND)
		return
	}
	question := fmt.Sprintf(m.available, appVersion, latest, runningPath)
	if messageBoxResult(mainWindow, question, m.title, MB_YESNO|MB_ICONQUESTION|MB_TOPMOST|MB_SETFOREGROUND) != IDYES {
		return
	}
	messageBox(mainWindow, m.downloading, m.title, MB_OK|MB_TOPMOST|MB_SETFOREGROUND)
	startUpdateDownload(result.release)
}

func currentExecutableDisplayPath() string {
	path, err := os.Executable()
	if err != nil {
		if language == "en" {
			return "(unknown)"
		}
		return "(알 수 없음)"
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return filepath.Clean(path)
}

func startUpdateDownload(release githubRelease) {
	updateMu.Lock()
	if updateDownloadBusy {
		updateMu.Unlock()
		return
	}
	updateDownloadBusy = true
	updateMu.Unlock()
	go func() {
		result := prepareUpdate(release)
		updateMu.Lock()
		pendingUpdateDownload = result
		updateMu.Unlock()
		procPostMessageW.Call(mainWindow, WM_UPDATE_DOWNLOAD_RESULT, 0, 0)
	}()
}

func handleUpdateDownloadResult() {
	updateMu.Lock()
	result := pendingUpdateDownload
	pendingUpdateDownload = updateDownloadResult{}
	updateDownloadBusy = false
	updateMu.Unlock()
	if result.err != nil {
		showError(fmt.Sprintf(currentUpdaterMessages().downloadFailure, result.err))
		return
	}
	cmd := exec.Command(result.updaterPath, "--apply-update", result.targetPath, result.newExePath)
	if err := cmd.Start(); err != nil {
		showError(fmt.Sprintf(currentUpdaterMessages().applyFailure, err))
		return
	}
	procDestroyWindow.Call(mainWindow)
}

func fetchLatestRelease() (githubRelease, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, latestReleaseAPI, nil)
	if err != nil {
		return githubRelease{}, err
	}
	req.Header.Set("User-Agent", "Quick-Anydesk-Connect/"+appVersion)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return githubRelease{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return githubRelease{}, fmt.Errorf("GitHub HTTP %d", resp.StatusCode)
	}
	var release githubRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&release); err != nil {
		return githubRelease{}, err
	}
	return release, nil
}

func prepareUpdate(release githubRelease) updateDownloadResult {
	var exeURL, hashURL string
	for _, asset := range release.Assets {
		switch asset.Name {
		case updateExeAssetName:
			exeURL = asset.BrowserDownloadURL
		case updateHashAssetName:
			hashURL = asset.BrowserDownloadURL
		}
	}
	if exeURL == "" || hashURL == "" {
		return updateDownloadResult{err: errors.New(currentUpdaterMessages().assetMissing)}
	}
	target, err := os.Executable()
	if err != nil {
		return updateDownloadResult{err: err}
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return updateDownloadResult{err: err}
	}
	tempDir, err := os.MkdirTemp("", "QuickAnydeskConnect-update-*")
	if err != nil {
		return updateDownloadResult{err: err}
	}
	newExe := filepath.Join(tempDir, "QuickAnydeskConnect.new.exe")
	hashFile := filepath.Join(tempDir, updateHashAssetName)
	updater := filepath.Join(tempDir, "QuickAnydeskConnect.updater.exe")
	if err := downloadFile(exeURL, newExe); err != nil {
		return updateDownloadResult{err: err}
	}
	if err := downloadFile(hashURL, hashFile); err != nil {
		return updateDownloadResult{err: err}
	}
	if err := verifySHA256File(newExe, hashFile); err != nil {
		return updateDownloadResult{err: err}
	}
	if err := copyFile(target, updater); err != nil {
		return updateDownloadResult{err: err}
	}
	return updateDownloadResult{updaterPath: updater, newExePath: newExe, targetPath: target}
}

func downloadFile(url, path string) error {
	client := &http.Client{Timeout: 90 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Quick-Anydesk-Connect/"+appVersion)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download HTTP %d", resp.StatusCode)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, io.LimitReader(resp.Body, 100<<20))
	return err
}

func verifySHA256File(exePath, hashPath string) error {
	f, err := os.Open(hashPath)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	var expected string
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 0 {
			expected = strings.ToLower(strings.TrimSpace(fields[0]))
		}
	}
	f.Close()
	if err := scanner.Err(); err != nil {
		return err
	}
	if len(expected) != 64 {
		return errors.New(currentUpdaterMessages().hashInvalid)
	}
	if _, err := hex.DecodeString(expected); err != nil {
		return errors.New(currentUpdaterMessages().hashInvalid)
	}
	exe, err := os.Open(exePath)
	if err != nil {
		return err
	}
	h := sha256.New()
	_, err = io.Copy(h, exe)
	exe.Close()
	if err != nil {
		return err
	}
	if hex.EncodeToString(h.Sum(nil)) != expected {
		return errors.New(currentUpdaterMessages().hashMismatch)
	}
	return nil
}

func runApplyUpdate(targetPath, newExePath string) error {
	time.Sleep(700 * time.Millisecond)
	backup := targetPath + ".update-backup"
	_ = os.Remove(backup)
	var lastErr error
	for i := 0; i < 80; i++ {
		if err := os.Rename(targetPath, backup); err != nil {
			lastErr = err
			time.Sleep(250 * time.Millisecond)
			continue
		}
		if err := copyFile(newExePath, targetPath); err != nil {
			_ = os.Rename(backup, targetPath)
			return err
		}
		_ = os.Remove(backup)
		if err := exec.Command(targetPath).Start(); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("could not replace executable: %w", lastErr)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func compareVersions(a, b string) int {
	pa, pb := parseVersion(a), parseVersion(b)
	for i := 0; i < 3; i++ {
		if pa[i] < pb[i] {
			return -1
		}
		if pa[i] > pb[i] {
			return 1
		}
	}
	return 0
}

func parseVersion(v string) [3]int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	v = strings.SplitN(v, "-", 2)[0]
	parts := strings.Split(v, ".")
	var out [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		out[i], _ = strconv.Atoi(parts[i])
	}
	return out
}
