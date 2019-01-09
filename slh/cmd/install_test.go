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

	"github.com/stretchr/testify/assert"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/tool"

	"github.com/adobe/sledgehammer/slh/alias"

	"github.com/adobe/sledgehammer/utils/test"
)

func TestInstallTool(t *testing.T) {
	pathToCreate := test.NewTmpDir(t)
	test.PrepareLocalRegistries(pathToCreate)
	defer test.DeleteTmpDir(pathToCreate, t)

	cases := []*test.TestCase{
		{
			Name: "Tool does not exist",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("install foobar"),
					Has: []string{tool.ErrorToolNotFound.Error()},
					Not: []string{"Usage"},
				},
			},
		},
		{
			Name: "Alias does already exist",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "file"},
				},
				{
					Cmd: fmt.Sprintf("install real -o json"),
					Has: []string{"success"},
				},
				{
					Cmd: fmt.Sprintf("install real"),
					Has: []string{alias.ErrorDuplicateAlias.Error()},
					Not: []string{"Usage"},
				},
				{
					Cmd: fmt.Sprintf("reset real -o json"),
					Has: []string{"success"},
				},
			},
		},
		{
			Name: "Symlink does already exist",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "file"},
				},
				{
					Cmd: fmt.Sprintf("install real"),
					Has: []string{alias.ErrorFileAlreadyPresent.Error()},
					Not: []string{"Usage"},
					DoBefore: func(cfg *config.Config) {
						alias.CreateSymlink("real")
					},
					DoAfter: func(cfg *config.Config) {
						alias.RemoveSymlink("real")
					},
				},
				{
					Cmd: fmt.Sprintf("reset real -o json"),
					Has: []string{"success"},
				},
			},
		},
		{
			Name: "Success with tool name only",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "file"},
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
			Name: "Success with registry and tool name",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "file"},
				},
				{
					Cmd: fmt.Sprintf("install real/real -o json"),
					Has: []string{"success"},
				},
				{
					Cmd: fmt.Sprintf("reset real -o json"),
					Has: []string{"success"},
				},
			},
		},
		{
			Name: "Alias does already exist, force overwrite",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "file"},
				},
				{
					Cmd: fmt.Sprintf("install real -o json"),
					Has: []string{"success"},
				},
				{
					Cmd: fmt.Sprintf("install real -o json --force"),
					Has: []string{"success"},
				},
				{
					Cmd: fmt.Sprintf("reset real -o json"),
					Has: []string{"success"},
				},
			},
		},
		{
			Name: "Symlink does already exist, force overwrite",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "file"},
				},
				{
					Cmd: fmt.Sprintf("install real -o json --force"),
					Has: []string{"success"},
					DoBefore: func(cfg *config.Config) {
						alias.CreateSymlink("real")
					},
				},
				{
					Cmd: fmt.Sprintf("reset real -o json"),
					Has: []string{"success"},
				},
			},
		},
		{
			Name: "Success with different name",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "file"},
				},
				{
					Cmd: fmt.Sprintf("install real --alias foobar -o json"),
					Has: []string{"success"},
				},
				{
					Cmd: fmt.Sprintf("reset foobar -o json"),
					Has: []string{"success"},
				},
			},
		},
		{
			Name: "Success with given version",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "file"},
				},
				{
					Cmd: fmt.Sprintf("install real -o json --version ~1.0"),
					Has: []string{"success"},
					DoAfter: func(cfg *config.Config) {
						db, err := cfg.OpenDatabase()
						if err != nil {
							t.Fatal(err)
						}
						if db != nil {
							defer cfg.CloseDatabase()
						}

						a := alias.New(config.Database{DB: db})
						ali, err := a.Get("real")
						if err != nil {
							t.Fatal(err)
						}
						assert.Equal(t, "~1.0", ali.Version)
					},
				},
				{
					Cmd: fmt.Sprintf("reset real -o json"),
					Has: []string{"success"},
				},
			},
		},
		{
			Name: "Install kits with registry name",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "foo.json")),
					Has: []string{"Name", "Type", "Maintainer", "foo", "file"},
				},
				{
					Cmd: fmt.Sprintf("install foo/foo-kit --kit -o json"),
					Has: []string{"success"},
				},
				{
					Cmd: fmt.Sprintf("reset foo -o json"),
					Has: []string{"success"},
				},
			},
		},
		{
			Name: "Install kits without registry name",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "foo.json")),
					Has: []string{"Name", "Type", "Maintainer", "foo", "file"},
				},
				{
					Cmd: fmt.Sprintf("install foo-kit --kit -o json"),
					Has: []string{"success"},
				},
				{
					Cmd: fmt.Sprintf("reset foo -o json"),
					Has: []string{"success"},
				},
			},
		},
		{
			Name: "Install with slh as alias",
			Steps: []*test.Step{
				{
					Cmd: fmt.Sprintf("create registry file %s", filepath.Join(pathToCreate, "real.json")),
					Has: []string{"Name", "Type", "Maintainer", "real", "file"},
				},
				{
					Cmd: fmt.Sprintf("install real --alias slh"),
					Has: []string{alias.ErrorNaughtyBoy.Error()},
				},
				{
					Cmd: fmt.Sprintf("reset real -o json"),
					Has: []string{"success"},
				},
			},
		},
	}
	test.DoTest(t, cases)
}
