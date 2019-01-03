/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adobe/sledgehammer/installer/env"
	docker "github.com/adobe/sledgehammer/utils/docker"

	bolt "github.com/coreos/bbolt"
	"github.com/fatih/color"
	client "github.com/fsouza/go-dockerclient"
)

var (
	// Will be injected with all possible flavour as string separated by ;
	flavours string
	// ErrorServerVersionNotFound will be thrown if the version of the docker daemon cannot be determined
	ErrorServerVersionNotFound = errors.New("Server version not fetchable")
	// ErrorClientVersionNotFound will be thrown if the version of the client cannot be determined
	ErrorClientVersionNotFound = errors.New("Client info not fetchable")
	// ErrorDockerMountMissing will be thrown if the docker socket is missing
	ErrorDockerMountMissing = errors.New("The docker mount is missing")
	// ErrorHostNameNotDetected will be thrown if the hostname of the container cannot be determined
	ErrorHostNameNotDetected = errors.New("Hostname not detected")
	// ErrorContainerNotFound will be thrown if the container running this installer cannot be found
	ErrorContainerNotFound = errors.New("Container not found")
	// ErrorNoVolumeMounted will be thrown if no volume could be detected that the executable should be installed to
	ErrorNoVolumeMounted = errors.New("No Volume mounted")
	// ErrorSystemNotDetected will be thrown if the flavour of the executable could not be determined
	ErrorSystemNotDetected = errors.New("Could not autodetect system")
	// ErrorSystemNotAvailable will be thrown if the selected flavour is not available for install
	ErrorSystemNotAvailable = errors.New("The selected system is not available")
	// ErrorCopyFailed will be thrown if the executable could not be copied to the local file system
	ErrorCopyFailed = errors.New("Could not copy the binary")
)

// Config The config for the main routine
type Config struct {
	Writer      io.Writer
	Docker      docker.Client
	Env         env.ENV
	Systems     []string
	DB          *bolt.DB
	WorkingPath string
	InstallPath string
}

func main() {
	// create a new client
	cl, err := client.NewClientFromEnv()
	if err != nil {
		dockerMountMissing(os.Stdout)
		os.Exit(1)
	}

	config := Config{
		Writer:      os.Stdout,
		Docker:      cl,
		Env:         &env.OSENV{},
		Systems:     strings.Split(flavours, ";"),
		WorkingPath: "/",
		InstallPath: "/data",
	}

	// call main install routine
	err = InstallSledgehammer(config)
	if err != nil {
		// fmt.Fprintln(os.Stdout, color.RedString("ERROR:"), err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

// InstallSledgehammer is the main function to check for all requirements and install the executable
func InstallSledgehammer(config Config) error {

	printHeader(config.Writer)

	err := checkImageList(config.Writer, config.Docker)
	if err != nil {
		return err
	}

	localPath, err := checkMounts(config.Writer, config.Docker)
	if err != nil {
		return err
	}

	selectedSystem, err := checkSystem(config, localPath)
	if err != nil {
		return err
	}

	err = copyBinary(config.WorkingPath, config.InstallPath, selectedSystem, config.Writer)
	if err != nil {
		return err
	}

	printFooter(config.Writer)
	return nil
}

func containsSystem(e string, systems []string) bool {
	for _, a := range systems {
		if a == e {
			return true
		}
	}
	return false
}

func detectSystem(writer io.Writer, client docker.Client) (string, error) {
	system, err := client.Version()
	if err != nil {
		fmt.Fprintln(writer, color.RedString("ERROR:"), "Could not fetch the server version")
		return "", ErrorServerVersionNotFound
	}
	info, err := client.Info()
	if err != nil {
		fmt.Fprintln(writer, color.RedString("ERROR:"), "Could not fetch the client info")
		return "", ErrorClientVersionNotFound
	}

	// check os and arch
	if system.Get("Os") == "linux" && system.Get("Arch") == "amd64" {
		if strings.Contains(strings.ToLower(info.OperatingSystem), "docker for windows") {
			return "windows-" + system.Get("Arch"), nil
		}
		if strings.Contains(strings.ToLower(info.OperatingSystem), "docker for mac") {
			return "darwin-" + system.Get("Arch"), nil
		}
	}
	return system.Get("Os") + "-" + system.Get("Arch"), nil
}

func checkImageList(writer io.Writer, dock docker.Client) error {

	// check if we can list the images
	_, err := dock.ListImages(client.ListImagesOptions{})
	if err != nil {
		dockerMountMissing(writer)
		return ErrorDockerMountMissing
	}
	return nil
}

func checkHostName(writer io.Writer) (string, error) {
	// get the hostname
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Fprintln(writer, color.RedString("ERROR:"), "Could not determine hostname")
		return "", ErrorHostNameNotDetected
	}
	return hostname, nil
}

func checkMounts(writer io.Writer, client docker.Client) (string, error) {
	var localPath string
	hostname, err := checkHostName(writer)
	if err != nil {
		return "", err
	}

	// inspect this container so we can check the mounts
	list, err := client.InspectContainer(hostname)
	if err != nil {
		fmt.Fprintln(writer, color.RedString("ERROR:"), "Could not inspect the running container, are you running me in docker?")
		return "", ErrorContainerNotFound
	}

	// loop over every mount and check if :/data is mounted
	dataMounted := false
	for _, volume := range list.Mounts {
		if volume.Destination == "/data" {
			dataMounted = true
			localPath = volume.Source
		}
	}

	// if not mounted
	if !dataMounted {
		noVolumeMounted(writer)
		return "", ErrorNoVolumeMounted
	}
	return localPath, nil
}

func checkSystem(config Config, localPath string) (string, error) {
	// check if a system is selected
	system, _ := config.Env.GetSystem()

	if len(system) == 0 {
		// try to detect the system automatically
		system, err := detectSystem(config.Writer, config.Docker)
		if err != nil {
			return "", err
		}
		if !containsSystem(system, config.Systems) {
			fmt.Fprintln(config.Writer, color.GreenString("INFO: "), "We could not autodetect the system to install.")
			return "", ErrorSystemNotDetected
		}

		fmt.Fprintln(config.Writer, color.GreenString("INFO: "), "We autodetected the system to be:")
		fmt.Fprintln(config.Writer, color.GreenString("INFO: "))
		fmt.Fprintln(config.Writer, color.GreenString("INFO: "), "	"+system)
		fmt.Fprintln(config.Writer, color.GreenString("INFO: "))
		noSystemSelected(config.Writer, localPath, config.Systems)
		// needed as system is locally scoped...
		return system, nil
	} else if !containsSystem(system, config.Systems) {
		fmt.Fprintln(config.Writer, color.RedString("ERROR: "), "The selected system is not available.")
		noSystemSelected(config.Writer, localPath, config.Systems)
		return "", ErrorSystemNotAvailable
	}
	return system, nil
}

func copyBinary(fromPath string, toPath string, system string, writer io.Writer) error {
	// system selected, move binary
	srcFile, err := os.Open(filepath.Join(fromPath, "slh/slh-"+system))
	if err != nil {
		fmt.Fprintln(writer, color.RedString("ERROR: "), "Could not open the binary '"+filepath.Join(fromPath, "/slh/slh-"+system)+"'")
		return ErrorCopyFailed
	}
	defer srcFile.Close()

	targetFile := "slh"
	if strings.Contains(system, "windows-") {
		// is windows, so add the well known .exe suffix
		targetFile += ".exe"
	}

	destFile, err := os.Create(filepath.Join(toPath, targetFile))
	if err != nil {
		fmt.Fprintln(writer, color.RedString("ERROR: "), "Could not create the binary '"+filepath.Join(toPath, targetFile)+"'")
		return ErrorCopyFailed
	}
	err = os.Chmod(filepath.Join(toPath, targetFile), 0777)
	if err != nil {
		fmt.Fprintln(writer, color.RedString("ERROR: "), "Could not make the binary '"+filepath.Join(toPath, targetFile)+"' executable")
		return ErrorCopyFailed
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		fmt.Fprintln(writer, color.RedString("ERROR: "), "Could not copy the binary '"+filepath.Join(fromPath, "/slh/slh-"+system)+"' to '"+filepath.Join(toPath, targetFile)+"'")
		return ErrorCopyFailed
	}

	err = destFile.Sync()
	if err != nil {
		fmt.Fprintln(writer, color.RedString("ERROR: "), "Could not sync final bytes from '"+filepath.Join(fromPath, "/slh/slh-"+system)+"' to '"+filepath.Join(toPath, targetFile)+"'")
		return ErrorCopyFailed
	}

	return nil
}
