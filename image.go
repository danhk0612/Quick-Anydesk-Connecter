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
	"unsafe"
)

const (
	BI_RGB       = 0
	BI_BITFIELDS = 3
)

func clipboardHasImage() bool {
	for _, format := range []uintptr{CF_DIBV5, CF_DIB, CF_BITMAP} {
		if r, _, _ := procIsClipboardFormatAvailable.Call(format); r != 0 {
			return true
		}
	}
	return false
}

func readClipboardImage() (image.Image, []byte, [32]byte, error) {
	if r, _, _ := procOpenClipboard.Call(mainWindow); r == 0 {
		return nil, nil, [32]byte{}, fmt.Errorf("%s", currentMessages().errClipboardImageRead)
	}
	defer procCloseClipboard.Call()

	var img image.Image
	var err error
	for _, format := range []uintptr{CF_DIBV5, CF_DIB} {
		if r, _, _ := procIsClipboardFormatAvailable.Call(format); r == 0 {
			continue
		}
		handle, _, _ := procGetClipboardData.Call(format)
		if handle == 0 {
			continue
		}
		ptr, _, _ := procGlobalLock.Call(handle)
		if ptr == 0 {
			continue
		}
		size, _, _ := procGlobalSize.Call(handle)
		if size == 0 {
			procGlobalUnlock.Call(handle)
			continue
		}
		raw := append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(ptr)), int(size))...)
		procGlobalUnlock.Call(handle)
		img, err = decodeDIB(raw)
		if err == nil {
			break
		}
	}

	if img == nil && clipboardFormatAvailable(CF_BITMAP) {
		handle, _, _ := procGetClipboardData.Call(CF_BITMAP)
		if handle != 0 {
			img, err = imageFromHBITMAP(handle)
		}
	}
	if img == nil {
		if err == nil {
			err = fmt.Errorf("%s", currentMessages().errClipboardImageRead)
		}
		return nil, nil, [32]byte{}, err
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
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < nh; y++ {
		sy := b.Min.Y + y*h/nh
		for x := 0; x < nw; x++ {
			sx := b.Min.X + x*w/nw
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
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
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < nh; y++ {
		sy := b.Min.Y + y*h/nh
		for x := 0; x < nw; x++ {
			sx := b.Min.X + x*w/nw
			dst.Set(x, y, src.At(sx, sy))
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
