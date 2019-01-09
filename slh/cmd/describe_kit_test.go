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
	"github.com/adobe/sledgehammer/slh/kit"
	"github.com/adobe/sledgehammer/utils/test"
)

func TestDescribeKit(t *testing.T) {
	pathToCreate := test.NewTmpDir(t)
	test.PrepareLocalRegistries(pathToCreate)
	defer test.DeleteTmpDir(pathToCreate, t)

	cases := []*test.TestCase{
		{
			Name: "Describe valid kit",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "foo.json")),
					Has: []string{"file", "foo", "plaschke@adobe.com"},
				},
				{
					Cmd: "describe kit foo/foo-kit",
					Has: []string{"foo", "Version", "Alias", "foo-kit"},
				},
			},
		},
		{
			Name: "Describe invalid kit",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "foo.json")),
					Has: []string{"file", "foo", "plaschke@adobe.com"},
				},
				{
					Cmd: fmt.Sprintf("describe kit foobar"),
					Has: []string{cmd.ErrorNoRegistryGiven.Error()},
					Not: []string{"Usage"},
				},
			},
		},
		{
			Name: "Describe kit with invalid registry",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "foo.json")),
					Has: []string{"file", "foo", "plaschke@adobe.com"},
				},
				{
					Cmd: fmt.Sprintf("describe kit foo/bar"),
					Has: []string{kit.ErrorKitNotFound.Error()},
					Not: []string{"Usage"},
				},
			},
		},
	}
	test.DoTest(t, cases)
}
