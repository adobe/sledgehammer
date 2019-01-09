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
					Cmd: fmt.Sprintf("cr reg file %s", filepath.Join(pathToCreate, "foo.json")),
					Has: []string{"Name", "Type", "Maintainer", "foo", "file"},
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
					Cmd: fmt.Sprintf("cr reg file %s", filepath.Join(pathToCreate, "baz.json")),
					Has: []string{"Name", "Type", "Maintainer", "baz", "file"},
				},
				{
					Cmd: fmt.Sprintf("get to ba"),
					Has: []string{"Name", "Registry", "Installed", "baz"},
					Not: []string{"foo"},
				},
			},
		},
		{
			Name: "Multiple registries",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("cr reg file %s --name baz1", filepath.Join(pathToCreate, "baz.json")),
					Has: []string{"Name", "Type", "Maintainer", "baz1", "file"},
				},
				{
					Cmd: fmt.Sprintf("cr reg file %s --name baz2", filepath.Join(pathToCreate, "baz.json")),
					Has: []string{"Name", "Type", "Maintainer", "baz2", "file"},
				},
				{
					Cmd: fmt.Sprintf("cr reg file %s --name baz3", filepath.Join(pathToCreate, "baz.json")),
					Has: []string{"Name", "Type", "Maintainer", "baz3", "file"},
				},
				{
					Cmd: fmt.Sprintf("get to"),
					Has: []string{"Name", "Registry", "Installed", "baz1", "baz2", "baz3", "foo"},
				},
			},
		},
	}
	test.DoTest(t, cases)
}
