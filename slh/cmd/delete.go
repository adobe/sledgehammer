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

func DeleteCommand(cfg *config.Config) *cobra.Command {
	deleteCommand := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"rm", "remove", "del"},
		Short:   "Delete a resource",
		Long:    "Will delete a given ressource registered with Sledgehammer",
	}

	// add list commands
	deleteCommand.AddCommand(DeleteMountCommand(cfg))
	deleteCommand.AddCommand(DeleteRegistryCommand(cfg))

	return deleteCommand
}
