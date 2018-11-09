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
	"github.com/adobe/sledgehammer/slh/registry"
	"github.com/spf13/cobra"
)

func DeleteRegistryCommand(cfg *config.Config) *cobra.Command {
	deleteRegistryCommand := &cobra.Command{
		Use:     "registry <name>",
		Short:   "Deletes a registry",
		Long:    "Will delete the given registry <name> from Sledgehammer",
		Aliases: []string{"reg", "registries"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := DeleteRegistry(cfg, args[0])
			if err != nil {
				cmd.SilenceUsage = true
			}
			return err
		},
	}
	return deleteRegistryCommand
}

// DeleteRegistry will delete the given registry from Sledgehammer
func DeleteRegistry(cfg *config.Config, name string) error {
	database, err := cfg.OpenDatabase()
	if database != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}

	registries := registry.New(config.Database{DB: database})

	err = registries.Remove(name)
	if err != nil {
		return err
	}
	return GetRegistries(cfg.WithDatabase(database))
}
