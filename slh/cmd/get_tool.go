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
	"github.com/adobe/sledgehammer/slh/alias"
	"github.com/adobe/sledgehammer/slh/out"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/tool"
	"github.com/spf13/cobra"
)

func GetToolCommand(cfg *config.Config) *cobra.Command {
	toolCommand := &cobra.Command{
		Use:     "tool",
		Short:   "Get all tools",
		Long:    "Will get all tools registered on this system for Sledgehammer",
		Aliases: []string{"tools", "to"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return GetTools(cfg, args[0])
			}
			return GetTools(cfg, "")
		},
	}

	toolCommand.AddCommand(DescribeToolCommand(cfg))

	return toolCommand
}

// GetTools will get all tools that are registered with Sledgehammer on this system
func GetTools(cfg *config.Config, search string) error {
	var err error

	database, err := cfg.OpenDatabase()
	if database != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}

	m := tool.New(config.Database{DB: database})
	a := alias.New(config.Database{DB: database})

	var sortedTools []string
	var toolsMap map[string][]tool.Tool

	if len(search) > 0 {
		sortedTools, toolsMap, err = m.Search(search)
	} else {
		sortedTools, toolsMap, err = m.List()
	}
	if err != nil {
		return err
	}

	// tool registry default
	table := out.NewTable("Tools", "Name", "Registry", "Default", "Installed", "Image")
	table.MergeCells("...")

	for _, mo := range sortedTools {
		for _, sto := range toolsMap[mo] {
			aliases, err := a.From(sto.Data().Name, sto.Data().Registry)
			if err != nil {
				return err
			}
			table.Add(sto.Data().Name, sto.Data().Registry+"/"+sto.Data().Name, sto.Data().Default, len(aliases) > 0, tool.FullImage(sto, ""))
		}
	}
	cfg.Output.Set(table)

	return nil
}
