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
	"strings"

	"github.com/adobe/sledgehammer/slh/out"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/registry"
	"github.com/spf13/cobra"
)

func GetKitCommand(cfg *config.Config) *cobra.Command {
	getKitCommand := &cobra.Command{
		Use:     "kit",
		Short:   "Get all kits",
		Long:    "Will get all kits registered on this system for Sledgehammer",
		Aliases: []string{"kits", "ki"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetKits(cfg)
		},
	}
	return getKitCommand
}

// GetKits will get all kits that are registered with Sledgehammer on this system
func GetKits(cfg *config.Config) error {
	database, err := cfg.OpenDatabase()
	if database != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}

	r := registry.New(config.Database{DB: database})

	regs, err := r.List()
	if err != nil {
		return err
	}

	// tool registry default
	table := out.NewTable("Kits", "Name", "Registry", "Description", "Tools")
	// table.MergeCells("...")

	for _, re := range regs {
		kits, err := re.Kits()
		if err != nil {
			return err
		}
		for _, k := range kits {
			toolStr := []string{}
			for _, t := range k.Tools {
				toolStr = append(toolStr, t.Name)
			}
			table.Add(k.Name, re.Data().Name, k.Description, strings.Join(toolStr, ", "))
		}
	}
	cfg.Output.Set(table)

	return nil
}
