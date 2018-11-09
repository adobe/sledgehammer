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

	"github.com/adobe/sledgehammer/utils"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/version"

	"github.com/adobe/sledgehammer/utils/test"
)

func TestRoot(t *testing.T) {
	var ver string
	cases := []*test.TestCase{
		{
			Name: "Get version",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("--version"),
					DoBefore: func(cfg *config.Config) {
						ver = utils.RandomString(8)
						version.Version = ver
					},
					Has: []string{ver},
				},
			},
		},
	}
	test.DoTest(t, cases)
}
