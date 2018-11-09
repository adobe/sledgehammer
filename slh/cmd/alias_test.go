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

	"github.com/adobe/sledgehammer/slh/cmd"
	"github.com/adobe/sledgehammer/utils/test"
)

func TestAlias(t *testing.T) {
	pathToCreate := test.NewTmpDir(t)
	os.MkdirAll(filepath.Join(pathToCreate, "tmp"), 0777)
	defer test.DeleteTmpDir(pathToCreate, t)

	cases := []*test.TestCase{
		{
			Name: "Valid path",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create mount %s/tmp", pathToCreate),
					Has: []string{fmt.Sprintf("%s/tmp", pathToCreate)},
				},
			},
		},
		{
			Name: "Invalid path",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create mount /foobar"),
					Has: []string{cmd.ErrorInvalidPath.Error()},
				},
			},
		},
		{
			Name: "Valid path, replace",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create mount %s/tmp", pathToCreate),
					Has: []string{fmt.Sprintf("%s/tmp", pathToCreate)},
				},
				{
					Cmd: fmt.Sprintf("create mount %s", pathToCreate),
					Has: []string{fmt.Sprintf("%s", pathToCreate)},
					Not: []string{fmt.Sprintf("%s/tmp", pathToCreate)},
				},
			},
		},
		{
			Name: "Valid path, already exists",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create mount %s/tmp", pathToCreate),
					Has: []string{fmt.Sprintf("%s/tmp", pathToCreate)},
				},
				{
					Cmd: fmt.Sprintf("create mount %s/tmp", pathToCreate),
					Has: []string{fmt.Sprintf("%s/tmp", pathToCreate)},
				},
			},
		},
		{
			Name: "Valid path, parent already exists",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create mount %s", pathToCreate),
					Has: []string{fmt.Sprintf("%s", pathToCreate)},
				},
				{
					Cmd: fmt.Sprintf("create mount %s/tmp", pathToCreate),
					Has: []string{fmt.Sprintf("%s", pathToCreate)},
					Not: []string{fmt.Sprintf("%s/tmp", pathToCreate)},
				},
			},
		},
	}
	test.DoTest(t, cases)
}
