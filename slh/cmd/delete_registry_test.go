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

func TestDeleteRegistry(t *testing.T) {
	pathToCreate := test.NewTmpDir(t)
	test.PrepareLocalRegistries(pathToCreate)
	defer test.DeleteTmpDir(pathToCreate, t)

	cases := []*test.TestCase{
		{
			Name: "Delete default registry",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("delete registry default"),
					Not: []string{"file", "foo", "plaschke@adobe.com"},
				},
			},
		},
		{
			Name: "Delete valid registries",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "foo.json")),
					Has: []string{"file", "foo", "plaschke@adobe.com"},
				},
				{
					Cmd: fmt.Sprintf("delete registry foo"),
					Not: []string{"file", "foo"},
				},
				{
					Cmd: fmt.Sprintf("delete registry default"),
					Not: []string{"file", "foo", "plaschke@adobe.com"},
				},
			},
		},
		{
			Name: "Delete invalid registry",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "foo.json")),
					Has: []string{"file", "foo", "plaschke@adobe.com"},
				},
				{
					Cmd: fmt.Sprintf("delete registry foobar"),
					Has: []string{registry.ErrorRegistryNotFound.Error()},
					Not: []string{"Usage"},
				},
			},
		},
	}
	test.DoTest(t, cases)
}
