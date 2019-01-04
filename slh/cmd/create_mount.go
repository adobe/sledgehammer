/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/adobe/sledgehammer/utils"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/mount"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var (
	// ErrorInvalidPath will be thrown if the given path is invalid
	ErrorInvalidPath = errors.New("Could not create given path, please make sure it exists and is a directory")
)

func CreateMountCommand(cfg *config.Config) *cobra.Command {
	createMountCommand := &cobra.Command{
		Use:     "mount <path>",
		Short:   "Create a mount",
		Long:    "Will create a given mount (paths) to Sledgehammer",
		Aliases: []string{"mo", "mounts"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := CreateMount(cfg, args[0])
			if err != nil {
				cmd.SilenceUsage = true
			}
			return err
		},
	}
	return createMountCommand
}

// CreateMount will create a new mount if possible
func CreateMount(cfg *config.Config, path string) error {
	database, err := cfg.OpenDatabase()
	if database != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}
	path = utils.ImportPath(path)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if _, err := os.Stat(absPath); err != nil {
		return ErrorInvalidPath
	}

	m := mount.New(config.Database{DB: database})

	err = m.Add(path)
	if err != nil {
		return err
	}
	return GetMounts(cfg.WithDatabase(database))
}

func addDefaultMount(cfg *config.Config) error {
	home, err := homedir.Dir()
	home = utils.ImportPath(home)
	if err != nil {
		return err
	}
	return CreateMount(cfg, home)

}
