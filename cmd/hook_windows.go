//go:build windows
// +build windows

package main

import (
	"syscall"
	"time"
)

const (
	CTRL_C_EVENT        = uint32(0)
	CTRL_BREAK_EVENT    = uint32(1)
	CTRL_CLOSE_EVENT    = uint32(2)
	CTRL_LOGOFF_EVENT   = uint32(5)
	CTRL_SHUTDOWN_EVENT = uint32(6)
)

var (
	kernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleCtrlHandler = kernel32.NewProc("SetConsoleCtrlHandler")
)

func exitHook(f func()) error {
	n, _, err := procSetConsoleCtrlHandler.Call(
		syscall.NewCallback(func(controlType uint32) uint {
			f()
			time.Sleep(time.Second * 1)
			switch controlType {
			case CTRL_CLOSE_EVENT:
				return 1
			default:
				return 0
			}
		}),
		1)

	if n == 0 {
		return err
	}
	return nil
}
