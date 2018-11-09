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

	"github.com/adobe/sledgehammer/slh/registry"
	"github.com/adobe/sledgehammer/utils/test"
)

func TestDescribeRegistry(t *testing.T) {
	pathToCreate := test.NewTmpDir(t)
	test.PrepareLocalRegistries(pathToCreate)
	defer test.DeleteTmpDir(pathToCreate, t)

	cases := []*test.TestCase{
		{
			Name: "Describe valid registry",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "foo.json")),
					Has: []string{"local", "foo", "plaschke@adobe.com"},
				},
				{
					Cmd: fmt.Sprintf("describe registry foo"),
					Has: []string{"local", "foo", "plaschke@adobe.com", "Tools", "Last_Update"},
				},
			},
		},
		{
			Name: "Describe invalid registry",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "foo.json")),
					Has: []string{"local", "foo", "plaschke@adobe.com"},
				},
				{
					Cmd: fmt.Sprintf("describe registry foobar"),
					Has: []string{registry.ErrorRegistryNotFound.Error()},
					Not: []string{"Usage"},
				},
			},
		},
	}
	test.DoTest(t, cases)
}
