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

func TestReset(t *testing.T) {
	pathToCreate := test.NewTmpDir(t)
	test.PrepareLocalRegistries(pathToCreate)
	defer test.DeleteTmpDir(pathToCreate, t)

	cases := []*test.TestCase{
		{
			Name: "Reset a valid tool",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "local"},
				},
				{
					Cmd: fmt.Sprintf("install real -o json"),
					Has: []string{"success"},
				},
				{
					Cmd: fmt.Sprintf("reset real -o json"),
					Has: []string{"success"},
				},
			},
		},
		{
			Name: "Reset an invalid tool",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("reset real -o json"),
					Has: []string{"success"},
				},
			},
		},
		// {
		// 	Name: "Reset daemon tools",
		// 	Steps: []*test.Step{
		// 		{
		// 			Cmd: fmt.Sprintf("reset real -o json"),
		// 			Has: []string{"success"},
		// 		},
		// 	},
		// },
	}
	test.DoTest(t, cases)
}
