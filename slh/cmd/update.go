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
	"sync"

	"github.com/adobe/sledgehammer/slh/cache"
	"github.com/adobe/sledgehammer/slh/out"
	"github.com/adobe/sledgehammer/slh/registry"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/tool"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// UpdateCommand is the command to update all registries and tools
type updateCommand struct {
	Force bool
}

func UpdateCommand(cfg *config.Config) *cobra.Command {
	updateCmd := updateCommand{}
	updateCommand := &cobra.Command{
		Use:   "update",
		Short: "Update all registries",
		Long:  "Will loop over each registries and updates their tools",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateCmd.Execute(cfg)
		},
	}

	updateCommand.Flags().BoolVar(&updateCmd.Force, "force", false, "Will force the update of all entities, regardless of TTL")

	return updateCommand
}

// Execute will update all registries with the current tools
func (u *updateCommand) Execute(cfg *config.Config) error {

	db, err := cfg.OpenDatabase()
	if db != nil {
		defer cfg.CloseDatabase()
	}
	if err != nil {
		return err
	}

	logrus.Debug("Updating all registries")

	tools := tool.New(config.Database{DB: db})
	r := registry.New(config.Database{DB: db})
	registries, err := r.List()
	caches := cache.New(config.Database{DB: db})

	if err != nil {
		return err
	}

	// clear caches
	// caches.Container.ClearAll()
	// caches.Registry.ClearAll()
	// caches.Versions.ClearAll()

	// Custom wait group for all registries
	var wg sync.WaitGroup

	for _, reg := range registries {
		wg.Add(1)
		go func(reg registry.Registry) {
			defer wg.Done()
			currentTools, err := tools.From(reg.Data().Name)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"registry": reg.Data().Name,
					"error":    err.Error(),
				}).Warn("Error occured during fetching all tools for a registry")
				return
			}
			if u.Force {
				caches.Registry.Clear(reg)
			}
			_, err = caches.Registry.LastUpdate(reg)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"registry": reg.Data().Name,
					"error":    err.Error(),
				}).Warn("Error occured while updating the registry")
				return
			}
			newTools, err := reg.Tools()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"registry": reg.Data().Name,
					"error":    err.Error(),
				}).Warn("Error occured during fetching all tools from a registry")
				return
			}

			for _, to := range newTools {
				logrus.WithField("tool", to.Data().Name).WithField("registry", reg.Data().Name).Debug("Found a new tool")
				tools.Add(to)
			}

			for _, to := range currentTools {
				if !tool.Exists(newTools, to) {
					logrus.WithField("tool", to.Data().Name).WithField("registry", reg.Data().Name).Debug("Removing an old tool")
					tools.Remove(to.Data().Registry, to.Data().Name)
				}
			}
			logrus.WithField("registry", reg.Data().Name).Debug("Updated registry")
		}(reg)
	}
	// wait for all registries to be updated
	wg.Wait()
	cfg.Output.Set(out.NewSuccess())
	return nil
}
