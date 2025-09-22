//go:build windows
// +build windows

package clipboard

import (
	"syscall"
	"unsafe"
)

var (
	user32                     = syscall.NewLazyDLL("user32.dll")
	openClipboard              = user32.NewProc("OpenClipboard")
	closeClipboard             = user32.NewProc("CloseClipboard")
	emptyClipboard             = user32.NewProc("EmptyClipboard")
	getClipboardData           = user32.NewProc("GetClipboardData")
	setClipboardData           = user32.NewProc("SetClipboardData")
	isClipboardFormatAvailable = user32.NewProc("IsClipboardFormatAvailable")

	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	globalAlloc  = kernel32.NewProc("GlobalAlloc")
	globalFree   = kernel32.NewProc("GlobalFree")
	globalLock   = kernel32.NewProc("GlobalLock")
	globalUnlock = kernel32.NewProc("GlobalUnlock")
	lstrcpy      = kernel32.NewProc("lstrcpyW")
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

func GetClipboard() (string, error) {
	r, _, _ := openClipboard.Call(0)
	if r == 0 {
		return "", syscall.GetLastError()
	}
	defer closeClipboard.Call()

	formatAvailable, _, _ := isClipboardFormatAvailable.Call(cfUnicodeText)
	if formatAvailable == 0 {
		return "", nil
	}

	h, _, _ := getClipboardData.Call(cfUnicodeText)
	if h == 0 {
		return "", syscall.GetLastError()
	}

	l, _, _ := globalLock.Call(h)
	if l == 0 {
		return "", syscall.GetLastError()
	}
	defer globalUnlock.Call(h)

	text := syscall.UTF16ToString((*[1 << 20]uint16)(unsafe.Pointer(l))[:])
	return text, nil
}

func SetClipboard(text string) error {
	r, _, _ := openClipboard.Call(0)
	if r == 0 {
		return syscall.GetLastError()
	}
	defer closeClipboard.Call()

	emptyClipboard.Call()

	// Конвертируем в UTF-16
	text16, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}

	// Выделяем память
	size := len(text16)*2 + 2
	h, _, _ := globalAlloc.Call(gmemMoveable, uintptr(size))
	if h == 0 {
		return syscall.GetLastError()
	}

	l, _, _ := globalLock.Call(h)
	if l == 0 {
		globalFree.Call(h)
		return syscall.GetLastError()
	}

	// Копируем данные
	lstrcpy.Call(l, uintptr(unsafe.Pointer(&text16[0])))
	globalUnlock.Call(h)

	setClipboardData.Call(cfUnicodeText, h)
	return nil
}
