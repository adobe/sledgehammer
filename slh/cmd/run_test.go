/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package cmd_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/adobe/sledgehammer/slh/cmd"
	"github.com/adobe/sledgehammer/utils/test"
)

func TestRun(t *testing.T) {
	pathToCreate := test.NewTmpDir(t)
	test.PrepareLocalRegistries(pathToCreate)
	defer test.DeleteTmpDir(pathToCreate, t)

	cases := []*test.TestCase{
		{
			Name: "Tool not available",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("run foobar"),
					Has: []string{cmd.ErrorToolNotFound.Error()},
				},
			},
		},
		{
			Name: "Success",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "local"},
				},
				{
					Cmd: fmt.Sprintf("run real"),
					Has: []string{"Hello from the real tool"},
				},
			},
		},
		{
			Name: "Success with a version",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "local"},
				},
				{
					Cmd: fmt.Sprintf("run alpine-test-version --version 3.7"),
					Has: []string{"3.7."},
				},
			},
		},
		{
			Name: "Success with update",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "local"},
				},
				{
					Cmd: fmt.Sprintf("run real --update"),
					Has: []string{"Hello from the real tool"},
				},
			},
		},
		{
			Name: "Success with arguments",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "local"},
				},
				{
					Cmd: fmt.Sprintf("run argument-test -a FOOBAR!!!"),
					Has: []string{"Foo: FOOBAR!!!"},
				},
			},
		},
		{
			Name: "Success with daemon tool",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "local"},
				},
				{
					Cmd: fmt.Sprintf("run real-daemon"),
					Has: []string{"Hello from the daemon tool"},
				},
				{
					Cmd: fmt.Sprintf("reset -o json"),
					Has: []string{"success"},
				},
			},
		},
	}
	test.DoTest(t, cases)
}
