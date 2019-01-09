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

	"github.com/adobe/sledgehammer/slh/config"

	"github.com/adobe/sledgehammer/utils/test"
)

func TestUpdate(t *testing.T) {
	pathToCreate := test.NewTmpDir(t)
	test.PrepareLocalRegistries(pathToCreate)
	defer test.DeleteTmpDir(pathToCreate, t)
	cases := []*test.TestCase{
		{
			Name: "Update a tool",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("cr reg file %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "file"},
				},
				{
					Cmd: fmt.Sprintf("describe to real"),
					Has: []string{"Name", "Registry", "sh -c echo Hello from the real tool"},
					Not: []string{"now updated"},
				},
				{
					Cmd: fmt.Sprintf("update -o json"),
					Has: []string{"success"},
					DoBefore: func(cfg *config.Config) {
						test.PrepareChangedRegistries(pathToCreate)
					},
				},
				{
					Cmd: fmt.Sprintf("describe to real"),
					Has: []string{"Name", "Registry", "sh -c echo Hello from the real tool, now updated"},
				},
			},
		},
	}
	test.DoTest(t, cases)
}
