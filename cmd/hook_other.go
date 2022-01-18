//go:build !windows
// +build !windows

package main

func exitHook(func()) error {
	return nil
}
