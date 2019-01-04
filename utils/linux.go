// +build !windows

package utils

func DecorateExecutable(name string) string {
	return name
}

func ContainerPath(path string) string {
	return path
}
