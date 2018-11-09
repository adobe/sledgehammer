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
	"os"
	"path/filepath"

	"github.com/adobe/sledgehammer/slh/cmd"
	"github.com/adobe/sledgehammer/slh/config"
	"github.com/fsouza/go-dockerclient"
)

func main() {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		panic(err)
	}
	cfg := &config.Config{
		IO: &config.IO{
			Out: os.Stdout,
			In:  os.Stdin,
			Err: os.Stderr,
		},
		ConfigDir: "",
		Docker: config.Docker{
			Docker: client,
		},
	}
	// TODO: Check if the executable is called slh or different.
	executableName := filepath.Base(os.Args[0])
	if executableName != "slh" && executableName != "slh.exe" {
		// git rev-parse HEAD
		// -> slh run --alias git -a rev-parse -a HEAD --version <from-alias>
		os.Args = append([]string{"slh", "alias", executableName}, os.Args[1:]...)
	}
	cmd.Execute(cfg)
}
