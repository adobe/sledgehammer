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
	"fmt"
	"io"

	"github.com/adobe/sledgehammer/slh/version"

	"github.com/fatih/color"
)

func printFooter(writer io.Writer) {
	fmt.Fprintln(writer, color.GreenString("INFO: "))
	fmt.Fprintln(writer, color.GreenString("INFO: "), "Sledgehammer installed, call it with 'slh'")
}

func printHeader(writer io.Writer) {
	fmt.Fprintln(writer, color.GreenString("INFO: "), "Welcome to Sledgehammer", version.Version)
	fmt.Fprintln(writer, color.GreenString("INFO: "))
}

func dockerNotInstalled(writer io.Writer) {
	fmt.Fprintln(writer, color.GreenString("INFO: "), "The docker executable could not be found.")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "Without it the installer can not perform all neccessary steps.")
}

func noVolumeMounted(writer io.Writer) {
	fmt.Fprintln(writer, color.GreenString("INFO: "), "STEP 2 of 2:")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "Please provide a local directory for the `slh` executable")
	fmt.Fprintln(writer, color.GreenString("INFO: "))
	fmt.Fprintln(writer, color.GreenString("INFO: "), "Mount a local directory in on you computer to ':/bin'")
	fmt.Fprintln(writer, color.GreenString("INFO: "))
	fmt.Fprintln(writer, color.GreenString("INFO: "), "Simplest syntax:")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "	docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "		-v <YOUR_LOCAL_PATH>:/bin")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "		adobe/slh")
}

func noSystemSelected(writer io.Writer, localPath string, systems []string) {
	fmt.Fprintln(writer, color.GreenString("INFO: "), "Available systems:")
	fmt.Fprintln(writer, color.GreenString("INFO: "))
	for _, bin := range systems {
		fmt.Fprintln(writer, color.GreenString("INFO: "), "	- "+bin)
	}
}

func dockerMountMissing(writer io.Writer) {
	fmt.Fprintln(writer, color.GreenString("INFO: "), "STEP 1 of 2:")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "Please provide the docker socket")
	fmt.Fprintln(writer, color.GreenString("INFO: "))
	fmt.Fprintln(writer, color.GreenString("INFO: "), "For most cases you just need to add the /var/run/docker.sock location:")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "  'docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock adobe/slh'")
	fmt.Fprintln(writer, color.GreenString("INFO: "))
	fmt.Fprintln(writer, color.GreenString("INFO: "), "If you know what you are doing, you can also add the DOCKER_HOST:")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "	 'docker run -it --rm -e DOCKER_HOST=<daemon-location> adobe/slh")
}
