//go:build !windows
// +build !windows

package warn

import "fmt"

func Warn(content string) {
	fmt.Println(content)
}
