// +build windows

package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

func DecorateExecutable(name string) string {
	if !strings.HasSuffix(name, ".exe") {
		return fmt.Sprintf("%s.exe", name)
	}
	return name
}

func ContainerPath(path string) string {
	// C:/Users/labuser -> /slh/mnt/C/Users/labuser
	vol := filepath.VolumeName(path)
	if len(vol) == 2 {
		path = strings.Replace(path, vol, fmt.Sprintf("/slh/mnt/%v", []rune(vol)[0]), 1)
	}
	return path
}
