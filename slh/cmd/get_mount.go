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
	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/mount"
	"github.com/adobe/sledgehammer/slh/out"
	"github.com/spf13/cobra"
)

func GetMountCommand(cfg *config.Config) *cobra.Command {
	getMountCommand := &cobra.Command{
		Use:     "mount",
		Short:   "Get all mounts",
		Long:    "Will get all mounts (paths) registered on this system for Sledgehammer",
		Aliases: []string{"mounts", "mo"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetMounts(cfg)
		},
	}
	return getMountCommand
}

// GetMounts will get all mounts that are registered with Sledgehammer on this system
func GetMounts(cfg *config.Config) error {
	database, err := cfg.OpenDatabase()
	if database != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}

	m := mount.New(config.Database{DB: database})

	mounts, err := m.List()
	if err != nil {
		return err
	}
	table := out.NewTable("Mounts", "Mounts")
	for _, m := range mounts {
		table.Add(m)
	}

	cfg.Output.Set(table)
	return nil
}
