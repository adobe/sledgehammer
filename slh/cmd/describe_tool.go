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

	"github.com/adobe/sledgehammer/slh/out"

	"github.com/adobe/sledgehammer/slh/cache"

	"github.com/adobe/sledgehammer/slh/version"
	"github.com/adobe/sledgehammer/utils"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/tool"
	"github.com/spf13/cobra"
)

func DescribeToolCommand(cfg *config.Config) *cobra.Command {
	describeToolCommand := &cobra.Command{
		Use:     "tool <name>",
		Short:   "Describe details about a tool",
		Long:    "Will describe detailed information about a given tool",
		Aliases: []string{"tools", "to"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := DescribeTool(cfg, args[0])
			if err != nil {
				cmd.SilenceUsage = true
			}
			return err
		},
	}
	return describeToolCommand
}

// DescribeTool will describe detailed information about a tool
func DescribeTool(cfg *config.Config, name string) error {
	database, err := cfg.OpenDatabase()
	if database != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}

	registryName, toolName := utils.GetRegistryAndTool(name)

	tools := tool.New(config.Database{DB: database})
	caches := cache.New(config.Database{DB: database})

	to, err := tools.Get(registryName, toolName)
	if err != nil || to == nil {
		return err
	}

	localVersions, err := caches.Versions.Local(to, cfg.Docker)
	if err != nil {
		localVersions = []string{err.Error()}
	}
	remoteVersions, err := caches.Versions.Remote(to)
	if err != nil {
		remoteVersions = []string{err.Error()}
	}

	ct := out.NewContainer(to.Data().Name)
	ct.Add(out.NewValue("Name", to.Data().Name))
	ct.Add(out.NewValue("Registry", to.Data().Registry))
	ct.Add(out.NewValue("Type", to.Data().Type))
	ct.Add(out.NewValue("Image", filepath.Join(to.Data().ImageRegistry, to.Data().Image)))
	ct.Add(out.NewValue("Description", to.Data().Description))
	ct.Add(out.NewValue("Added", to.Data().Added))
	ct.Add(out.NewValue("IsDefault", to.Data().Default))
	ct.Add(out.NewValue("Entry", to.Data().Entry))
	ct.Add(out.NewValue("Daemon", to.Data().Daemon != nil))
	if to.Data().Daemon != nil {
		ct.Add(out.NewValue("Daemon Entry", to.Data().Daemon.Entry))
	}
	ct.Add(out.NewEmpty())

	mergedVersions := version.Merge(localVersions, remoteVersions)
	if len(mergedVersions) > 0 {
		table := out.NewTable("Versions", "Version", "Local")
		table.MergeCells("...")
		for i := len(mergedVersions) - 1; i >= 0; i-- {
			ver := mergedVersions[i]
			table.Add(ver, version.Has(localVersions, ver))
		}
		ct.Add(table)
	}

	cfg.Output.Set(ct)
	return nil
}
