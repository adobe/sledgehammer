// +build windows

package utils

import "fmt"

func DecorateExecutable(name string) string {
	return fmt.Sprintf("%s.exe", name)
}
