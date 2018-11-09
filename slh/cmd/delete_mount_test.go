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
	"os"
	"path/filepath"
	"testing"

	"github.com/adobe/sledgehammer/utils/test"
)

func TestDeleteMount(t *testing.T) {
	pathToCreate := test.NewTmpDir(t)
	os.MkdirAll(filepath.Join(pathToCreate, "tmp"), 0777)
	defer test.DeleteTmpDir(pathToCreate, t)

	cases := []*test.TestCase{
		{
			Name: "Delete valid mount",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create mount %s", pathToCreate),
					Has: []string{fmt.Sprintf("%s", pathToCreate)},
				},
				{
					Cmd: fmt.Sprintf("delete mount %s", pathToCreate),
					Not: []string{fmt.Sprintf("%s", pathToCreate)},
				},
			},
		},
		{
			Name: "Delete invalid mount",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create mount %s", pathToCreate),
					Has: []string{fmt.Sprintf("%s", pathToCreate)},
				},
				{
					Cmd: fmt.Sprintf("delete mount %s/foobar", pathToCreate),
					Has: []string{fmt.Sprintf("%s", pathToCreate)},
				},
			},
		},
	}
	test.DoTest(t, cases)
}
