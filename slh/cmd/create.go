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

func CreateCommand(cfg *config.Config) *cobra.Command {

	createCommand := &cobra.Command{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "Create a ressources",
		Long:    "Will create a ressource in Sledgehammer",
	}

	createCommand.AddCommand(CreateMountCommand(cfg))
	createCommand.AddCommand(CreateRegistryCommand(cfg))

	return createCommand
}
