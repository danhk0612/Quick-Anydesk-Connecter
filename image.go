//go:build windows

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"syscall"
	"time"
	"unsafe"
)

const (
	BI_RGB       = 0
	BI_BITFIELDS = 3
)

func clipboardHasImage() bool {
	for _, format := range []uintptr{registeredClipboardFormat("PNG"), registeredClipboardFormat("image/png"), CF_DIBV5, CF_DIB, CF_BITMAP} {
		if format == 0 {
			continue
		}
		if r, _, _ := procIsClipboardFormatAvailable.Call(format); r != 0 {
			return true
		}
	}
	return false
}

func registeredClipboardFormat(name string) uintptr {
	p, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return 0
	}
	r, _, _ := procRegisterClipboardFormatW.Call(uintptr(unsafe.Pointer(p)))
	return r
}

func openClipboardWithRetry() bool {
	// WM_CLIPBOARDUPDATE can arrive while the source application still owns
	// the clipboard for a few milliseconds. Retry briefly instead of treating
	// that normal race as a read failure.
	for i := 0; i < 10; i++ {
		if r, _, _ := procOpenClipboard.Call(mainWindow); r != 0 {
			return true
		}
		if i < 9 {
			time.Sleep(15 * time.Millisecond)
		}
	}
	return false
}

func readClipboardGlobalBytes(format uintptr) ([]byte, bool) {
	if format == 0 || !clipboardFormatAvailable(format) {
		return nil, false
	}
	handle, _, _ := procGetClipboardData.Call(format)
	if handle == 0 {
		return nil, false
	}
	ptr, _, _ := procGlobalLock.Call(handle)
	if ptr == 0 {
		return nil, false
	}
	defer procGlobalUnlock.Call(handle)
	size, _, _ := procGlobalSize.Call(handle)
	if size == 0 {
		return nil, false
	}
	return append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(ptr)), int(size))...), true
}

func readClipboardImage() (image.Image, []byte, [32]byte, error) {
	if !openClipboardWithRetry() {
		return nil, nil, [32]byte{}, fmt.Errorf("%s", currentMessages().errClipboardImageRead)
	}
	defer procCloseClipboard.Call()

	var img image.Image
	var err error

	// Many browsers and screenshot tools expose a registered PNG clipboard
	// format. Prefer it because it preserves the image without DIB quirks.
	for _, format := range []uintptr{registeredClipboardFormat("PNG"), registeredClipboardFormat("image/png")} {
		if raw, ok := readClipboardGlobalBytes(format); ok {
			decoded, _, decodeErr := image.Decode(bytes.NewReader(raw))
			if decodeErr == nil {
				img = decoded
				break
			}
			err = decodeErr
		}
	}

	// Standard Windows bitmap clipboard formats.
	if img == nil {
		for _, format := range []uintptr{CF_DIBV5, CF_DIB} {
			raw, ok := readClipboardGlobalBytes(format)
			if !ok {
				continue
			}
			img, err = decodeDIB(raw)
			if err == nil {
				break
			}
		}
	}

	if img == nil && clipboardFormatAvailable(CF_BITMAP) {
		handle, _, _ := procGetClipboardData.Call(CF_BITMAP)
		if handle != 0 {
			img, err = imageFromHBITMAP(handle)
		}
	}
	if img == nil {
		return nil, nil, [32]byte{}, fmt.Errorf("%s", currentMessages().errClipboardImageRead)
	}

	analysisImage := resizeImageToMax(img, 1600)
	var buf bytes.Buffer
	if err := png.Encode(&buf, analysisImage); err != nil {
		return nil, nil, [32]byte{}, fmt.Errorf(currentMessages().errImageEncode, err)
	}
	data := buf.Bytes()
	return img, data, sha256.Sum256(data), nil
}

func clipboardFormatAvailable(format uintptr) bool {
	r, _, _ := procIsClipboardFormatAvailable.Call(format)
	return r != 0
}

func decodeDIB(data []byte) (image.Image, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("invalid DIB header")
	}
	headerSize := int(binary.LittleEndian.Uint32(data[0:4]))
	if headerSize < 40 || headerSize > len(data) {
		return nil, fmt.Errorf("unsupported DIB header")
	}
	width := int(int32(binary.LittleEndian.Uint32(data[4:8])))
	rawHeight := int(int32(binary.LittleEndian.Uint32(data[8:12])))
	bitCount := int(binary.LittleEndian.Uint16(data[14:16]))
	compression := binary.LittleEndian.Uint32(data[16:20])
	if width <= 0 || rawHeight == 0 || (bitCount != 24 && bitCount != 32) {
		return nil, fmt.Errorf("unsupported DIB format")
	}
	if compression != BI_RGB && compression != BI_BITFIELDS {
		return nil, fmt.Errorf("unsupported DIB compression")
	}
	topDown := rawHeight < 0
	height := rawHeight
	if height < 0 {
		height = -height
	}

	offset := headerSize
	if headerSize == 40 && compression == BI_BITFIELDS {
		offset += 12
	}
	stride := ((width*bitCount + 31) / 32) * 4
	needed := offset + stride*height
	if offset < 0 || needed > len(data) {
		return nil, fmt.Errorf("truncated DIB")
	}

	out := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		srcY := y
		if !topDown {
			srcY = height - 1 - y
		}
		row := data[offset+srcY*stride:]
		for x := 0; x < width; x++ {
			pos := x * (bitCount / 8)
			b, g, r := row[pos], row[pos+1], row[pos+2]
			a := byte(255)
			if bitCount == 32 && row[pos+3] != 0 {
				a = row[pos+3]
			}
			out.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
	return out, nil
}

func imageFromHBITMAP(hBitmap uintptr) (image.Image, error) {
	var bm bitmap
	if r, _, _ := procGetObjectW.Call(hBitmap, unsafe.Sizeof(bm), uintptr(unsafe.Pointer(&bm))); r == 0 {
		return nil, fmt.Errorf("GetObject failed")
	}
	width, height := int(bm.bmWidth), int(bm.bmHeight)
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid bitmap size")
	}

	header := bitmapInfoHeader{biSize: uint32(unsafe.Sizeof(bitmapInfoHeader{})), biWidth: int32(width), biHeight: -int32(height), biPlanes: 1, biBitCount: 32, biCompression: BI_RGB}
	info := bitmapInfo{bmiHeader: header}
	pixels := make([]byte, width*height*4)
	hdc, _, _ := procGetDC.Call(0)
	if hdc == 0 {
		return nil, fmt.Errorf("GetDC failed")
	}
	defer procReleaseDC.Call(0, hdc)
	if r, _, _ := procGetDIBits.Call(hdc, hBitmap, 0, uintptr(height), uintptr(unsafe.Pointer(&pixels[0])), uintptr(unsafe.Pointer(&info)), DIB_RGB_COLORS); r == 0 {
		return nil, fmt.Errorf("GetDIBits failed")
	}
	out := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := (y*width + x) * 4
			out.SetRGBA(x, y, color.RGBA{R: pixels[pos+2], G: pixels[pos+1], B: pixels[pos], A: 255})
		}
	}
	return out, nil
}

func resizeImageToMax(src image.Image, maxSide int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxSide && h <= maxSide {
		return src
	}
	nw, nh := w, h
	if w >= h {
		nw = maxSide
		nh = maxInt(1, h*maxSide/w)
	} else {
		nh = maxSide
		nw = maxInt(1, w*maxSide/h)
	}
	return resizeImageHighQuality(src, nw, nh)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func resizeImageToFit(src image.Image, maxWidth, maxHeight int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxWidth && h <= maxHeight {
		return src
	}
	scaleW := float64(maxWidth) / float64(w)
	scaleH := float64(maxHeight) / float64(h)
	scale := scaleW
	if scaleH < scale {
		scale = scaleH
	}
	nw := maxInt(1, int(float64(w)*scale))
	nh := maxInt(1, int(float64(h)*scale))
	return resizeImageHighQuality(src, nw, nh)
}

func resizeImageHighQuality(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if width <= 0 || height <= 0 || sw <= 0 || sh <= 0 {
		return dst
	}

	// Bilinear interpolation avoids the severe aliasing/moiré produced by the
	// previous nearest-neighbor sampling when downscaling photos of displays.
	for y := 0; y < height; y++ {
		sy := (float64(y)+0.5)*float64(sh)/float64(height) - 0.5
		y0 := int(sy)
		fy := sy - float64(y0)
		if y0 < 0 {
			y0, fy = 0, 0
		}
		y1 := y0 + 1
		if y1 >= sh {
			y1 = sh - 1
		}
		for x := 0; x < width; x++ {
			sx := (float64(x)+0.5)*float64(sw)/float64(width) - 0.5
			x0 := int(sx)
			fx := sx - float64(x0)
			if x0 < 0 {
				x0, fx = 0, 0
			}
			x1 := x0 + 1
			if x1 >= sw {
				x1 = sw - 1
			}

			r00, g00, b00, a00 := src.At(b.Min.X+x0, b.Min.Y+y0).RGBA()
			r10, g10, b10, a10 := src.At(b.Min.X+x1, b.Min.Y+y0).RGBA()
			r01, g01, b01, a01 := src.At(b.Min.X+x0, b.Min.Y+y1).RGBA()
			r11, g11, b11, a11 := src.At(b.Min.X+x1, b.Min.Y+y1).RGBA()

			interp := func(v00, v10, v01, v11 uint32) uint8 {
				top := float64(v00)*(1-fx) + float64(v10)*fx
				bottom := float64(v01)*(1-fx) + float64(v11)*fx
				v := top*(1-fy) + bottom*fy
				return uint8(v/257.0 + 0.5)
			}
			dst.SetRGBA(x, y, color.RGBA{
				R: interp(r00, r10, r01, r11),
				G: interp(g00, g10, g01, g11),
				B: interp(b00, b10, b01, b11),
				A: interp(a00, a10, a01, a11),
			})
		}
	}
	return dst
}

func createHBITMAPFromImage(src image.Image) (uintptr, int, int, error) {
	preview := resizeImageToFit(src, 700, 460)
	b := preview.Bounds()
	w, h := b.Dx(), b.Dy()
	header := bitmapInfoHeader{biSize: uint32(unsafe.Sizeof(bitmapInfoHeader{})), biWidth: int32(w), biHeight: -int32(h), biPlanes: 1, biBitCount: 32, biCompression: BI_RGB}
	info := bitmapInfo{bmiHeader: header}
	var bits uintptr
	hbmp, _, _ := procCreateDIBSection.Call(0, uintptr(unsafe.Pointer(&info)), DIB_RGB_COLORS, uintptr(unsafe.Pointer(&bits)), 0, 0)
	if hbmp == 0 || bits == 0 {
		return 0, 0, 0, fmt.Errorf("CreateDIBSection failed")
	}
	pixels := unsafe.Slice((*byte)(unsafe.Pointer(bits)), w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, bv, _ := preview.At(b.Min.X+x, b.Min.Y+y).RGBA()
			pos := (y*w + x) * 4
			pixels[pos] = byte(bv >> 8)
			pixels[pos+1] = byte(g >> 8)
			pixels[pos+2] = byte(r >> 8)
			pixels[pos+3] = 255
		}
	}
	return hbmp, w, h, nil
}

func hashEqual(a, b [32]byte) bool { return a == b }

func utf16Ptr(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }
