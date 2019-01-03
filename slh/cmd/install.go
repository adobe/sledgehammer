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
	"fmt"

	"github.com/adobe/sledgehammer/slh/alias"
	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/out"
	"github.com/adobe/sledgehammer/slh/registry"
	"github.com/adobe/sledgehammer/slh/tool"
	"github.com/adobe/sledgehammer/utils"
	"github.com/spf13/cobra"
)

var (
	// ErrorToolNotFound will be thrown if a tool cannot be found
	ErrorToolNotFound = errors.New("Tool not found")
)

type installCommand struct {
	registry string
	tool     string
	alias    string
	version  string
	force    bool
	isKit    bool
}

func InstallCommand(cfg *config.Config) *cobra.Command {
	installCmd := installCommand{}
	installCommand := &cobra.Command{
		Use:   "install <tool>",
		Short: "Install a tool on the system",
		Long:  "Will install the given tool on the system and symlinks it to the Sledgehammer executable",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			installCmd.registry, installCmd.tool = utils.GetRegistryAndTool(args[0])
			err := installCmd.Execute(cfg)
			if err != nil {
				cmd.SilenceUsage = true
			}
			return err
		},
	}

	installCommand.Flags().StringVar(&installCmd.alias, "alias", "", "The alias which should be used. It then can be called by this alias (e.g. py2)")
	installCommand.Flags().StringVar(&installCmd.version, "version", "latest", "The version constraint that should be used (e.g. '^2' to stay on major version 2).")
	installCommand.Flags().BoolVar(&installCmd.force, "force", false, "True if the installation should be forced. Will overwrite previous installed tools.")
	installCommand.Flags().BoolVar(&installCmd.isKit, "kit", false, "True if the type is a kit that should be installed")

	return installCommand
}

func (cmd *installCommand) Execute(cfg *config.Config) error {
	if cmd.isKit {
		return cmd.InstallKit(cfg)
	}
	return cmd.InstallTool(cfg)
}

func (cmd *installCommand) InstallKit(cfg *config.Config) error {
	database, err := cfg.OpenDatabase()
	if err != nil {
		return err
	}
	if database != nil {
		defer cfg.CloseDatabase()
	}

	r := registry.New(config.Database{DB: database})

	regs, err := r.List()
	if err != nil {
		return err
	}

	// get kit
	for _, reg := range regs {
		kits, err := reg.Kits()
		if err != nil {
			return err
		}
		for _, kit := range kits {
			if kit.Name == cmd.tool && (len(cmd.registry) == 0 || cmd.registry == reg.Data().Name) {
				subCfg := cfg.WithDatabase(database)
				// get tool
				for _, t := range kit.Tools {
					// install each tool
					alias := t.Alias
					if len(t.Alias) == 0 {
						alias = t.Name
					}
					alias = fmt.Sprintf("%s%s", cmd.alias, alias)
					c := installCommand{
						alias:    alias,
						force:    cmd.force,
						registry: cmd.registry,
						tool:     t.Name,
						version:  t.Version,
					}
					err = c.InstallTool(subCfg)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	cfg.Output.Set(out.NewSuccess())
	return nil
}

func (cmd *installCommand) InstallTool(cfg *config.Config) error {

	if len(cmd.alias) == 0 {
		cmd.alias = utils.DecorateExecutable(cmd.tool)
	}

	database, err := cfg.OpenDatabase()
	if err != nil {
		return err
	}
	if database != nil {
		defer cfg.CloseDatabase()
	}

	tools := tool.New(config.Database{DB: database})
	aliases := alias.New(config.Database{DB: database})

	to, err := tools.Get(cmd.registry, cmd.tool)
	if err != nil {
		return err
	}
	hasAlias, err := aliases.Has(cmd.alias)
	if err != nil {
		return err
	}

	if hasAlias && !cmd.force {
		return alias.ErrorDuplicateAlias
	}

	hasSymlink, err := alias.HasSymlink(cmd.alias)
	if err != nil {
		return err
	}

	if hasSymlink && !cmd.force {
		return alias.ErrorFileAlreadyPresent
	} else if hasSymlink && cmd.force {
		err = alias.RemoveSymlink(cmd.alias)
		if err != nil {
			return err
		}
	}

	err = aliases.Add(alias.Alias{
		Name:     cmd.alias,
		Registry: to.Data().Registry,
		Tool:     to.Data().Name,
		Version:  cmd.version,
	})
	if err != nil {
		return err
	}
	// close as early as possible
	cfg.CloseDatabase()

	cfg.Output.Set(out.NewSuccess())
	return alias.CreateSymlink(cmd.alias)
}
