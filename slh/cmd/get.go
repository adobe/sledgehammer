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
	"github.com/adobe/sledgehammer/slh/config"
	"github.com/spf13/cobra"
)

func GetCommand(cfg *config.Config) *cobra.Command {
	getCommand := &cobra.Command{
		Use:     "get",
		Aliases: []string{"ls"},
		Short:   "Get ressources",
		Long:    "Will get any ressource registered with Sledgehammer",
	}

	// add get commands
	getCommand.AddCommand(GetMountCommand(cfg))
	getCommand.AddCommand(GetRegistryCommand(cfg))
	getCommand.AddCommand(GetToolCommand(cfg))
	getCommand.AddCommand(GetKitCommand(cfg))

	return getCommand
}
