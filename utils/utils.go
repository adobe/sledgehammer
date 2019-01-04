/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package utils

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	docker "github.com/fsouza/go-dockerclient"
)

// GetRegistryAndTool will return the registry and the tool name in separat variables if possible
func GetRegistryAndTool(tool string) (string, string) {
	toRun := strings.Split(tool, "/")
	if len(toRun) == 1 {
		return "", tool
	}
	return strings.Join(toRun[:len(toRun)-1], "/"), toRun[len(toRun)-1]
}

// Exists will check if the given path exists
func Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	// fmt.Println(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// ExecutablePath will get the current path of the executable.
// Useful for creating symlinks, as symlinks are only created in the same directory as the slh executable
func ExecutablePath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(ex), nil
}

// ExecutableName will get the current name of the executable.
func ExecutableName() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Base(ex), nil
}

// RandomString will generate a random string of a given length
func RandomString(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%X", b)
}

// IsPipe will determine if sledgehammer was called as part of a pipe or standalone
func IsPipe() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		logrus.Warnln(err.Error())
		return false
	}

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		logrus.Debugln("Determined that the input is a pipe")
		return true
	}
	logrus.Debugln("Determined that the input is NOT a pipe")
	return false
}

// PrepareEnvironment will prepare the local environment variables for the use in the container
func PrepareEnvironment(envs []string) []string {
	noEnv := []string{"PATH", "USER", "_", "TMP", "PWD", "SHELL"}

	// get environment
	var dockerEnvs []string

OUTER:
	for _, env := range envs {
		for _, noEnv := range noEnv {
			if strings.HasPrefix(env, noEnv) {
				dockerEnvs = append(dockerEnvs, "SLH_HOST_"+env)
				continue OUTER
			}
		}
		dockerEnvs = append(dockerEnvs, env)
	}

	return dockerEnvs
}

// PrepareMounts will prepare all mounts that should be used in the container
func PrepareMounts(mos []string) []docker.HostMount {
	var mounts []docker.HostMount

	for _, m := range mos {
		mounts = append(mounts, docker.HostMount{
			Source: m,
			Target: ContainerPath(m),
			Type:   "bind",
			// MacOS only
			// Consistency: "delegated",
		})
	}
	return mounts
}

func ImportPath(path string) string {
	// clean first
	path = filepath.Clean(path)
	path, _ = filepath.Abs(path)
	return path
}

func WorkingDirectory() (string, error) {
	workspace, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return ContainerPath(filepath.ToSlash(workspace)), nil
}
