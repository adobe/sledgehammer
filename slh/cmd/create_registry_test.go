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
	"github.com/adobe/sledgehammer/slh/registry"
	"github.com/adobe/sledgehammer/utils/test"
)

func TestCreateRegistry(t *testing.T) {

	pathToCreate := test.NewTmpDir(t)
	defer test.DeleteTmpDir(pathToCreate, t)
	test.PrepareLocalRegistries(pathToCreate)

	cases := []*test.TestCase{
		{
			Name: "Valid registry foo",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "foo.json")),
					Has: []string{"foo", "local", "plaschke@adobe.com"},
				},
			},
		},
		{
			Name: "Valid registry bar",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "bar.json")),
					Has: []string{"bar", "local", "plaschke@adobe.com"},
				},
			},
		},
		{
			Name: "Invalid registry foobar",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "foobar.json")),
					Has: []string{registry.ErrorNoValidPathGiven.Error()},
				},
			},
		},
		{
			Name: "Invalid registry with same name",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "bar.json")),
					Has: []string{"bar", "local", "plaschke@adobe.com"},
				},
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "bar.json")),
					Has: []string{registry.ErrorAlreadyExists.Error()},
				},
			},
		},
		{
			Name: "Valid registry with same name -- overwrite",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "bar.json")),
					Has: []string{"bar", "local", "plaschke@adobe.com"},
				},
				{
					Cmd: fmt.Sprintf("create registry local %s --name bar2", filepath.Join(pathToCreate, "bar.json")),
					Has: []string{"bar", "bar2", "...", "local", "plaschke@adobe.com"},
				},
			},
		},
		{
			Name: "Invalid registry type",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry foobar %s", filepath.Join(pathToCreate, "bar.json")),
					Has: []string{cmd.ErrorRegistryTypeNotSupported.Error()},
				},
			},
		},
		{
			Name: "Force creation",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "bar.json")),
					Has: []string{"bar", "local", "plaschke@adobe.com"},
				},
				{
					Cmd: fmt.Sprintf("create registry local %s --force", filepath.Join(pathToCreate, "bar.json")),
					Has: []string{"bar", "local", "local", "plaschke@adobe.com"},
					Not: []string{"..."},
				},
			},
		},
	}
	test.DoTest(t, cases)
}
