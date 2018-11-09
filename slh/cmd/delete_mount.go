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
	"path/filepath"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/mount"
	"github.com/spf13/cobra"
)

func DeleteMountCommand(cfg *config.Config) *cobra.Command {
	deleteMountCommand := &cobra.Command{
		Use:     "mount <path>",
		Short:   "Deletes a mount",
		Long:    "Will delete the given mount <path> from Sledgehammer",
		Aliases: []string{"mo", "mounts"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := DeleteMount(cfg, args[0])
			if err != nil {
				cmd.SilenceUsage = true
			}
			return err
		},
	}
	return deleteMountCommand
}

// DeleteMount will list all mounts that are registered with Sledgehammer on this system
func DeleteMount(cfg *config.Config, path string) error {
	database, err := cfg.OpenDatabase()
	if database != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	m := mount.New(config.Database{DB: database})

	err = m.Remove(absPath)
	if err != nil {
		return err
	}
	return GetMounts(cfg.WithDatabase(database))
}
