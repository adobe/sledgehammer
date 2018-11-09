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
	"sort"
	"strings"

	"github.com/adobe/sledgehammer/slh/out"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/registry"
	"github.com/spf13/cobra"
)

func GetRegistryCommand(cfg *config.Config) *cobra.Command {
	getRegistryCommand := &cobra.Command{
		Use:     "registry",
		Short:   "Get all registries",
		Long:    "Will get all registries registered on this system for Sledgehammer",
		Aliases: []string{"registries", "reg"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetRegistries(cfg)
		},
	}
	return getRegistryCommand
}

// GetRegistries will get all registries that are registered with Sledgehammer on this system
func GetRegistries(cfg *config.Config) error {
	database, err := cfg.OpenDatabase()
	if database != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}

	m := registry.New(config.Database{DB: database})

	registries, err := m.List()
	if err != nil {
		return err
	}
	sort.Slice(registries, func(i, j int) bool {
		return strings.Compare(registries[i].Data().Name, registries[j].Data().Name) < 0
	})

	table := out.NewTable("Registries", "Name", "Type", "Maintainer")
	table.MergeCells("...")
	for _, r := range registries {
		table.Add(r.Data().Name, r.Data().Type, r.Data().Maintainer)
	}

	cfg.Output.Set(table)
	return nil
}
