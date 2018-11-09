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
	"github.com/adobe/sledgehammer/slh/cache"
	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/out"
	"github.com/fsouza/go-dockerclient"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type resetCmd struct {
	alias    string
	resetAll bool
}

func ResetCommand(cfg *config.Config) *cobra.Command {
	resetCmd := resetCmd{}
	resetCommand := &cobra.Command{
		Use:     "reset <alias>",
		Short:   "Reset an alias",
		Long:    "Will remove the given alias from the system. Will also reset the daemon status if possible",
		Aliases: []string{"rst"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				resetCmd.alias = args[0]
			}
			err := resetCmd.Reset(cfg)
			if err != nil {
				cmd.SilenceUsage = true
			}
			return err
		},
	}

	resetCommand.Flags().BoolVar(&resetCmd.resetAll, "all", false, "Will reset all aliases and daemons")

	return resetCommand
}

func (cmd *resetCmd) ResetAll(cfg *config.Config) error {
	database, err := cfg.OpenDatabase()
	if database != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}

	// if no name is given, we assume a general reset which means to stop all daemons and remove all aliases

	c := cache.New(config.Database{DB: database})
	a := alias.New(config.Database{DB: database})

	daemons, err := c.Container.CurrentDaemons()
	if err != nil {
		return err
	}

	aliases, err := a.List()
	if err != nil {
		return err
	}

	cfgWithDatabase := cfg.WithDatabase(database)
	for _, al := range aliases {
		cmd := resetCmd{
			alias: al.Name,
		}
		err = cmd.ResetAlias(cfgWithDatabase)
		if err != nil {
			return err
		}
	}

	for tool, container := range daemons {
		logrus.WithField("id", container).WithField("image", tool).Debugln("Removing container for image")
		cfg.Docker.Docker.RemoveContainer(docker.RemoveContainerOptions{
			Force: true,
			ID:    container,
		})
		if err != nil {
			return err
		}
	}
	cfg.Output.Set(out.NewSuccess())
	return nil
}

func (cmd *resetCmd) ResetAlias(cfg *config.Config) error {
	database, err := cfg.OpenDatabase()
	if database != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}

	// delete symlink
	err = alias.RemoveSymlink(cmd.alias)
	if err != nil {
		return err
	}

	aliases := alias.New(config.Database{DB: database})

	// delete alias
	err = aliases.Remove(cmd.alias)
	if err != nil {
		return err
	}

	cfg.Output.Set(out.NewSuccess())

	return nil
}

// Reset will reset the given tool to the original state
func (cmd *resetCmd) Reset(cfg *config.Config) error {
	if len(cmd.alias) > 0 {
		return cmd.ResetAlias(cfg)
	}
	return cmd.ResetAll(cfg)
}
