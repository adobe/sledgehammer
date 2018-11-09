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
	fmt.Fprintln(writer, color.GreenString("INFO: "), "Installed 'slh' to the local directory, call it with 'slh'")
	fmt.Fprintln(writer, color.GreenString("INFO: "))
	fmt.Fprintln(writer, color.GreenString("INFO: "), "Sledgehammer installed, have fun!")
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
	fmt.Fprintln(writer, color.GreenString("INFO: "), "We could not detect a location to save the executable to.")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "Mount a local directory in on you computer to ':/data'")
	fmt.Fprintln(writer, color.GreenString("INFO: "))
	fmt.Fprintln(writer, color.GreenString("INFO: "), "Simplest syntax:")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "	docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "		-v <YOUR_LOCAL_PATH>:/data")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "		adobe/slh")
}

func noSystemSelected(writer io.Writer, localPath string, systems []string) {
	fmt.Fprintln(writer, color.GreenString("INFO: "), "If you want to manually select a system, choose from the following:")
	fmt.Fprintln(writer, color.GreenString("INFO: "))
	for _, data := range systems {
		fmt.Fprintln(writer, color.GreenString("INFO: "), "	- "+data)
	}
	fmt.Fprintln(writer, color.GreenString("INFO: "))
	fmt.Fprintln(writer, color.GreenString("INFO: "), "with:")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "	docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "		-v "+localPath+":/data")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "		-e SYSTEM=<YOUR_SYSTEM>")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "		adobe/slh")

}

func dockerMountMissing(writer io.Writer) {
	fmt.Fprintln(writer, color.GreenString("INFO: "), "You are missing a mandatory parameter:")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "	1. Mount 'docker.sock' for accessing Docker with unix sockets.")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "	2. Or, set DOCKER_HOST to Docker's location (unix or tcp).")
	fmt.Fprintln(writer, color.GreenString("INFO: "))
	fmt.Fprintln(writer, color.GreenString("INFO: "), "Mount Syntax:")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "	Start with 'docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock ...")
	fmt.Fprintln(writer, color.GreenString("INFO: "))
	fmt.Fprintln(writer, color.GreenString("INFO: "), "DOCKER_HOST Syntax:")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "	Start with 'docker run -it --rm -e DOCKER_HOST=<daemon-location> ...")
	fmt.Fprintln(writer, color.GreenString("INFO: "))
	fmt.Fprintln(writer, color.GreenString("INFO: "), "Possible root causes:")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "	1. Your admin has not granted permissions to /var/run/docker.sock.")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "	2. You passed '--user uid:gid' with bad values.")
	fmt.Fprintln(writer, color.GreenString("INFO: "), "	3. Your firewall is blocking TCP ports for accessing Docker daemon.")
}
