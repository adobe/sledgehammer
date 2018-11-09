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
	"github.com/adobe/sledgehammer/slh/cache"
	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/out"
	"github.com/adobe/sledgehammer/slh/registry"
	"github.com/spf13/cobra"
)

func DescribeRegistryCommand(cfg *config.Config) *cobra.Command {
	describeRegistryCommand := &cobra.Command{
		Use:     "registry <name>",
		Short:   "Describe a registry",
		Long:    "Will describe detailed information about a given registry",
		Aliases: []string{"registries", "reg"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := DescribeRegistry(cfg, args[0])
			if err != nil {
				cmd.SilenceUsage = true
			}
			return err
		},
	}
	return describeRegistryCommand
}

// DescribeRegistry will describe detailed information about a registry
func DescribeRegistry(cfg *config.Config, name string) error {
	database, err := cfg.OpenDatabase()
	if database != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}

	m := registry.New(config.Database{DB: database})
	c := cache.New(config.Database{DB: database})

	reg, err := m.Get(name)
	if err != nil {
		return err
	}

	lastUpdated, err := c.Registry.LastUpdate(reg)
	if err != nil {
		return err
	}
	tools, err := reg.Tools()
	if err != nil {
		return err
	}

	// kits, err := reg.Kits()
	// if err != nil {
	// 	return err
	// }
	ct := out.NewContainer(reg.Data().Name)

	ct.Add(out.NewValue("Name", reg.Data().Name))
	ct.Add(out.NewValue("Type", reg.Data().Type))
	ct.Add(out.NewValue("Maintainer", reg.Data().Maintainer))
	ct.Add(out.NewValue("Description", reg.Data().Description))
	ct.Add(out.NewValue("Last_Update", lastUpdated.String()))

	toolList := out.NewList("Tools")
	for _, t := range tools {
		toolList.Add(t.Data().Name)
	}

	reg.Info(ct)

	ct.Add(toolList)
	cfg.Output.Set(ct)
	return nil
}
