package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type pathCandidate struct {
	path    string
	modTime int64
}

func findAnyDesk() (string, error) {
	patterns := []string{
		`C:\Program Files (x86)\AnyDesk-*\AnyDesk-*.exe`,
		`C:\Program Files\AnyDesk-*\AnyDesk-*.exe`,
		`C:\Program Files (x86)\AnyDesk\AnyDesk.exe`,
		`C:\Program Files\AnyDesk\AnyDesk.exe`,
	}

	var candidates []pathCandidate

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}
			candidates = append(candidates, pathCandidate{
				path:    match,
				modTime: info.ModTime().UnixNano(),
			})
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("%s", currentMessages().errAnyDeskMissing)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].modTime > candidates[j].modTime
	})

	return candidates[0].path, nil
}

func normalizeID(s string) string {
	s = strings.TrimSpace(s)

	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		switch r {
		case ' ', '\t', '\r', '\n', '-':
			continue
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}

func isAnyDeskID(s string) bool {
	return idPattern.MatchString(s)
}

func ensureAnyDeskRunning(anyDeskPath string) error {
	running, err := isAnyDeskRunning(anyDeskPath)
	if err == nil && running {
		return nil
	}

	// AnyDesk가 완전히 종료된 경우 먼저 정상 초기화를 시킨다.
	cmd := exec.Command(anyDeskPath, "--start")
	if err := cmd.Start(); err != nil {
		return err
	}

	// --start 직후 곧바로 --with-password 세션을 붙이면
	// 일부 환경에서 검은 화면/응답 없음이 발생할 수 있어 초기화 시간을 둔다.
	time.Sleep(1500 * time.Millisecond)

	return nil
}

func isAnyDeskRunning(anyDeskPath string) (bool, error) {
	processName := filepath.Base(anyDeskPath)

	cmd := exec.Command(
		"tasklist",
		"/FI", "IMAGENAME eq "+processName,
		"/NH",
		"/FO", "CSV",
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	text := strings.ToLower(string(output))
	return strings.Contains(text, strings.ToLower(processName)), nil
}

func connect(anyDeskPath, target, p string) error {
	cmd := exec.Command(anyDeskPath, target, "--with-password")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := io.WriteString(stdin, p+"\n"); err != nil {
		_ = stdin.Close()
		return err
	}

	return stdin.Close()
}

func readClipboardText() (string, error) {
	if !openClipboardWithRetry() {
		return "", fmt.Errorf("clipboard unavailable")
	}
	defer procCloseClipboard.Call()

	h, _, _ := procGetClipboardData.Call(CF_UNICODETEXT)
	if h == 0 {
		return "", fmt.Errorf("clipboard has no unicode text")
	}

	p, _, _ := procGlobalLock.Call(h)
	if p == 0 {
		return "", fmt.Errorf("clipboard lock failed")
	}
	defer procGlobalUnlock.Call(h)

	u16 := make([]uint16, 0, 128)
	for i := uintptr(0); i < 1<<20; i++ {
		ch := *(*uint16)(unsafe.Pointer(p + i*2))
		if ch == 0 {
			break
		}
		u16 = append(u16, ch)
	}

	return syscall.UTF16ToString(u16), nil
}
