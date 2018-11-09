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
	"testing"

	"github.com/mitchellh/go-homedir"

	"github.com/adobe/sledgehammer/utils/test"
)

func TestGetMount(t *testing.T) {
	pathToCreate := test.NewTmpDir(t)
	home, err := homedir.Dir()
	if err != nil {
		t.Fatal(err)
	}
	defer test.DeleteTmpDir(pathToCreate, t)

	cases := []*test.TestCase{
		{
			Name: "No Mounts",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("get mounts"),
					Not: []string{fmt.Sprintf("%s", pathToCreate)},
					Has: []string{"Mounts"},
				},
			},
		},
		{
			Name: "Existing mounts",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create mount %s", pathToCreate),
					Has: []string{"Mounts", pathToCreate},
				},
				{
					Cmd: fmt.Sprintf("get mounts"),
					Has: []string{"Mounts", pathToCreate},
				},
			},
		},
		{
			Name: "Default homedir",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("get mounts"),
					Has: []string{"Mounts", home},
				},
			},
		},
	}
	test.DoTest(t, cases)
}
