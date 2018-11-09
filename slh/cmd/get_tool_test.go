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

	"github.com/adobe/sledgehammer/utils/test"
)

func TestGetTool(t *testing.T) {
	pathToCreate := test.NewTmpDir(t)
	test.PrepareLocalRegistries(pathToCreate)
	defer test.DeleteTmpDir(pathToCreate, t)

	cases := []*test.TestCase{
		{
			Name: "No Registries",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("get tools"),
					Not: []string{"foo"},
					Has: []string{"Name", "Registry", "Installed"},
				},
			},
		},
		{
			Name: "Existing registries",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("cr reg local %s", filepath.Join(pathToCreate, "foo.json")),
					Has: []string{"Name", "Type", "Maintainer", "foo", "local"},
				},
				{
					Cmd: fmt.Sprintf("get to"),
					Has: []string{"Name", "Registry", "Installed", "foo-tools/foo", "foo/foo"},
				},
			},
		},
		{
			Name: "Search",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("cr reg local %s", filepath.Join(pathToCreate, "baz.json")),
					Has: []string{"Name", "Type", "Maintainer", "baz", "local"},
				},
				{
					Cmd: fmt.Sprintf("get to ba"),
					Has: []string{"Name", "Registry", "Installed", "baz"},
					Not: []string{"foo"},
				},
			},
		},
	}
	test.DoTest(t, cases)
}
