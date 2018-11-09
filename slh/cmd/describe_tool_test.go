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

	"github.com/adobe/sledgehammer/slh/tool"
	"github.com/adobe/sledgehammer/utils/test"
)

func TestDescribeTool(t *testing.T) {
	pathToCreate := test.NewTmpDir(t)
	test.PrepareLocalRegistries(pathToCreate)
	defer test.DeleteTmpDir(pathToCreate, t)

	cases := []*test.TestCase{
		{
			Name: "Describe valid tool",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "foo.json")),
					Has: []string{"local", "foo", "plaschke@adobe.com"},
				},
				{
					Cmd: fmt.Sprintf("describe tool foo"),
					Has: []string{"Image", "foo-tools/foo", "IsDefault", "true"},
				},
			},
		},
		{
			Name: "Describe invalid tool",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry local %s", filepath.Join(pathToCreate, "foo.json")),
					Has: []string{"local", "foo", "plaschke@adobe.com"},
				},
				{
					Cmd: fmt.Sprintf("describe tool foobar"),
					Has: []string{tool.ErrorToolNotFound.Error()},
					Not: []string{"Usage"},
				},
			},
		},
	}
	test.DoTest(t, cases)
}
