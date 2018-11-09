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

	"github.com/adobe/sledgehammer/utils"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/kit"
	"github.com/adobe/sledgehammer/slh/out"
	"github.com/adobe/sledgehammer/slh/registry"
	"github.com/spf13/cobra"
)

var (
	// ErrorNoRegistryGiven will be thrown if no registry is passed
	ErrorNoRegistryGiven = errors.New("No registry passed to the command")
)

func DescribeKitCommand(cfg *config.Config) *cobra.Command {
	describeKitCommand := &cobra.Command{
		Use:     "kit <name>",
		Short:   "Describe details about a kit",
		Long:    "Will describe detailed information about a given kit",
		Aliases: []string{"kits", "ki"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, to := utils.GetRegistryAndTool(args[0])
			err := DescribeKit(cfg, reg, to)
			if err != nil {
				cmd.SilenceUsage = true
			}
			return err
		},
	}
	return describeKitCommand
}

// DescribeKit will describe detailed information about a kit
func DescribeKit(cfg *config.Config, registryName string, kitName string) error {

	if len(registryName) == 0 {
		return ErrorNoRegistryGiven
	}

	database, err := cfg.OpenDatabase()
	if database != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}

	r := registry.New(config.Database{DB: database})

	reg, err := r.Get(registryName)
	if err != nil {
		return err
	}

	kits, err := reg.Kits()
	if err != nil {
		return err
	}

	for _, kit := range kits {
		if kit.Name == kitName {
			ct := out.NewContainer(kit.Name)

			ct.Add(out.NewValue("Name", kit.Name))
			ct.Add(out.NewValue("Description", kit.Description))

			table := out.NewTable("Tools", "Tool", "Alias", "Version")

			for _, t := range kit.Tools {
				alias := t.Alias
				if len(t.Alias) == 0 {
					alias = t.Name
				}
				table.Add(t.Name, alias, t.Version)
			}
			ct.Add(out.NewEmpty())
			ct.Add(table)

			cfg.Output.Set(ct)
			return nil
		}
	}
	return kit.ErrorKitNotFound
}
