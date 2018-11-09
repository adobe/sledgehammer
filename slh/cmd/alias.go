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
	"github.com/adobe/sledgehammer/slh/config"
	"github.com/spf13/cobra"
)

func RunAliasCommand(cfg *config.Config) *cobra.Command {
	runAliasCommand := &cobra.Command{
		Use:                "alias",
		Short:              "List run an aliased tool",
		Long:               "Will run the aliased tool and passes all parameters to it",
		Aliases:            []string{"al"},
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := RunAlias(cfg, args[0], args[1:])
			if err != nil {
				cmd.SilenceUsage = true
			}
			return err
		},
	}
	runAliasCommand.Hidden = true
	return runAliasCommand
}

// RunAlias will run the given alias and passes all arguments to it
func RunAlias(cfg *config.Config, toolAlias string, arguments []string) error {
	database, err := cfg.OpenDatabase()
	if database != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}

	aliases := alias.New(config.Database{DB: database})

	al, err := aliases.Get(toolAlias)
	if err != nil {
		return err
	}

	cfg.CloseDatabase()

	runCommand := RunCmd{
		arguments: arguments,
		registry:  al.Registry,
		tool:      al.Tool,
		version:   al.Version,
	}

	return runCommand.Execute(cfg)
}
