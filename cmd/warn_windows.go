//go:build windows
// +build windows

package main

import (
	"syscall"
	"unsafe"
)

func boxW(hwnd uintptr, caption, title string, flags uint) int {
	captionPtr, _ := syscall.UTF16PtrFromString(caption)
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	ret, _, _ := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").Call(
		hwnd,
		uintptr(unsafe.Pointer(captionPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(flags),
	)
	return int(ret)
}

func warn(content string) {
	_ = boxW(0, content, "警告", 0x00000030|0x00000000)
}
