//go:build windows

package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"sync"
	"syscall"
	"unsafe"
)

const (
	NIM_MODIFY = 0x00000001
)

type trayVisualState int

const (
	trayVisualNormal trayVisualState = iota
	trayVisualImageAnalysis
	trayVisualBusy
)

var (
	visualIconOnce sync.Once
	windowAppIcon  uintptr
	trayIcons      [3]uintptr
	trayVisualNow  trayVisualState = -1
)

func initVisualIcons() {
	windowAppIcon = loadAppIcon()
	trayIcons[trayVisualNormal] = windowAppIcon
	trayIcons[trayVisualImageAnalysis] = createBadgedIcon(trayVisualImageAnalysis)
	trayIcons[trayVisualBusy] = createBadgedIcon(trayVisualBusy)
	for i := range trayIcons {
		if trayIcons[i] == 0 {
			trayIcons[i] = windowAppIcon
		}
	}
}

func appWindowIcon() uintptr {
	visualIconOnce.Do(initVisualIcons)
	return windowAppIcon
}

func registerWindowIconClasses() {
	icon := appWindowIcon()
	if icon == 0 {
		return
	}
	hInstance, _, _ := procGetModuleHandleW.Call(0)
	classes := []struct {
		name string
		proc uintptr
	}{
		{"QuickAnydeskConnectInputDialog", syscall.NewCallback(dialogWindowProc)},
		{"QuickAnydeskConnectOpenRouterDialog", syscall.NewCallback(dialogWindowProc)},
		{"QuickAnydeskConnectImagePreview", syscall.NewCallback(previewWindowProc)},
	}
	for _, class := range classes {
		name, _ := syscall.UTF16PtrFromString(class.name)
		wc := wndClassEx{
			cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
			lpfnWndProc:   class.proc,
			hInstance:     hInstance,
			hIcon:         icon,
			hIconSm:       icon,
			hbrBackground: dialogBgBrush,
			lpszClassName: name,
		}
		procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	}
}

func currentTrayVisualState() trayVisualState {
	if clipboardImageBusy || imageAnalysisBusy || updateCheckBusy || updateDownloadBusy {
		return trayVisualBusy
	}
	if imageAnalysisEnabled {
		return trayVisualImageAnalysis
	}
	return trayVisualNormal
}

func refreshTrayAppearance() {
	if trayIcon.hWnd == 0 {
		return
	}
	visualIconOnce.Do(initVisualIcons)
	state := currentTrayVisualState()
	icon := trayIcons[state]
	if icon == 0 {
		icon = windowAppIcon
	}
	trayIcon.hIcon = icon
	trayIcon.uFlags = NIF_ICON | NIF_TIP
	for i := range trayIcon.szTip {
		trayIcon.szTip[i] = 0
	}
	copy(trayIcon.szTip[:], syscall.StringToUTF16(trayTooltip(state)))
	procShellNotifyIconW.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&trayIcon)))
	trayVisualNow = state
}

func trayTooltip(state trayVisualState) string {
	if language == "en" {
		switch state {
		case trayVisualBusy:
			return "Quick Anydesk Connect - Working..."
		case trayVisualImageAnalysis:
			return "Quick Anydesk Connect - Image Analysis ON"
		default:
			return "Quick Anydesk Connect"
		}
	}
	switch state {
	case trayVisualBusy:
		return "Quick Anydesk Connect - 작업 중..."
	case trayVisualImageAnalysis:
		return "Quick Anydesk Connect - 이미지 자동 분석 ON"
	default:
		return "Quick Anydesk Connect"
	}
}

func createBadgedIcon(state trayVisualState) uintptr {
	base, ok := decodeEmbeddedIconImage()
	if !ok {
		return 0
	}
	base = resizeImageHighQuality(base, 32, 32)
	canvas := image.NewRGBA(image.Rect(0, 0, 32, 32))
	draw.Draw(canvas, canvas.Bounds(), base, base.Bounds().Min, draw.Src)

	badge := color.RGBA{R: 49, G: 181, B: 121, A: 255}
	if state == trayVisualBusy {
		badge = color.RGBA{R: 235, G: 157, B: 45, A: 255}
	}
	drawCircle(canvas, 24, 24, 8, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	drawCircle(canvas, 24, 24, 7, badge)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if state == trayVisualImageAnalysis {
		for d := -4; d <= 4; d++ {
			canvas.Set(24+d, 24, white)
			canvas.Set(24, 24+d, white)
		}
		canvas.Set(22, 22, white)
		canvas.Set(26, 22, white)
		canvas.Set(22, 26, white)
		canvas.Set(26, 26, white)
	} else {
		drawCircle(canvas, 20, 24, 1, white)
		drawCircle(canvas, 24, 24, 1, white)
		drawCircle(canvas, 28, 24, 1, white)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return 0
	}
	data := buf.Bytes()
	if len(data) == 0 {
		return 0
	}
	icon, _, _ := procCreateIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)), 1, 0x00030000, 32, 32, LR_DEFAULTCOLOR,
	)
	return icon
}

func decodeEmbeddedIconImage() (image.Image, bool) {
	if len(embeddedIcon) < 6 {
		return nil, false
	}
	count := int(uint16(embeddedIcon[4]) | uint16(embeddedIcon[5])<<8)
	bestScore := 1 << 30
	var best []byte
	for i := 0; i < count; i++ {
		base := 6 + i*16
		if base+16 > len(embeddedIcon) {
			break
		}
		w, h := int(embeddedIcon[base]), int(embeddedIcon[base+1])
		if w == 0 {
			w = 256
		}
		if h == 0 {
			h = 256
		}
		size := int(uint32(embeddedIcon[base+8]) | uint32(embeddedIcon[base+9])<<8 | uint32(embeddedIcon[base+10])<<16 | uint32(embeddedIcon[base+11])<<24)
		offset := int(uint32(embeddedIcon[base+12]) | uint32(embeddedIcon[base+13])<<8 | uint32(embeddedIcon[base+14])<<16 | uint32(embeddedIcon[base+15])<<24)
		if size <= 8 || offset < 0 || offset+size > len(embeddedIcon) {
			continue
		}
		data := embeddedIcon[offset : offset+size]
		if !bytes.Equal(data[:8], []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
			continue
		}
		score := absInt(w-32) + absInt(h-32)
		if score < bestScore {
			bestScore = score
			best = data
		}
	}
	if len(best) == 0 {
		return nil, false
	}
	img, err := png.Decode(bytes.NewReader(best))
	return img, err == nil
}

func drawCircle(dst *image.RGBA, cx, cy, r int, c color.RGBA) {
	r2 := r * r
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= r2 && image.Pt(x, y).In(dst.Bounds()) {
				dst.SetRGBA(x, y, c)
			}
		}
	}
}
