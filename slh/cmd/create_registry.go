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
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/registry"
	"github.com/spf13/cobra"
)

type createRegistryCommand struct {
	Name  string
	Type  string
	Force bool
}

var (
	// ErrorRegistryTypeNotSupported will be thrown if the registry type is not supported
	ErrorRegistryTypeNotSupported = errors.New("The type of the registry is not supported")
)

func CreateRegistryCommand(cfg *config.Config) *cobra.Command {
	createRegistryCmd := createRegistryCommand{}
	createRegistryCommand := &cobra.Command{
		Use:   "registry <type>",
		Short: "Create a registry",
		Long: `Will create a given registry to Sledgehammer. 
Currently supported registries are local|git|url.

Local: registry local <path> 
Git: registry git <repository>
URL: registry url <url>`,
		Aliases: []string{"reg", "registries"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			createRegistryCmd.Type = args[0]
			err := createRegistryCmd.CreateRegistry(cfg, args[1:])
			if err != nil {
				cmd.SilenceUsage = true
			}
			return err
		},
	}
	createRegistryCommand.Flags().StringVar(&createRegistryCmd.Name, "name", "", "Set the name of the registry. If not set will be determined by the registry itself.")
	createRegistryCommand.Flags().BoolVar(&createRegistryCmd.Force, "force", false, "Will overwrite existing registries")

	return createRegistryCommand
}

// CreateRegistry will create the given registry to Sledgehammer
func (cmd *createRegistryCommand) CreateRegistry(cfg *config.Config, arguments []string) error {
	database, err := cfg.OpenDatabase()
	if database != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}

	// detect type
	factory, found := registry.Types[cmd.Type]
	if !found {
		return ErrorRegistryTypeNotSupported
	}

	reg, err := factory.Create(registry.Data{
		Name: cmd.Name,
	}, arguments)

	if err != nil {
		return err
	}

	// create path for registry
	path, err := filepath.Abs(filepath.Join(cfg.ConfigDir, "registries", reg.Data().Name))
	if err != nil {
		return err
	}
	reg.Data().Path = path

	registries := registry.New(config.Database{DB: database})

	exists, err := registries.Exists(reg.Data().Name)
	if err != nil {
		return err
	}

	if exists && cmd.Force {
		registries.Remove(reg.Data().Name)
	}

	err = reg.Initialize()
	if err != nil {
		// remove the content there
		os.RemoveAll(path)
		logrus.Warningln("Could not initialize registry")
		return err
	}

	err = registries.Add(reg)
	if err != nil {
		return err
	}
	return GetRegistries(cfg.WithDatabase(database))
}

func addDefaultRegistry(cfg *config.Config) error {
	createRegistryCmd := createRegistryCommand{
		Type: "git",
		Name: "default",
	}
	err := createRegistryCmd.CreateRegistry(cfg, []string{registry.DefaultRegistryURL})
	return err
}
