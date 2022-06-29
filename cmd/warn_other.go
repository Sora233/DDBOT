//go:build !windows
// +build !windows

package main

import "fmt"

func warn(content string) {
	fmt.Println(content)
}
