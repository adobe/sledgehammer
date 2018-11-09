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

func DescribeCommand(cfg *config.Config) *cobra.Command {
	describeCommand := &cobra.Command{
		Use:     "describe",
		Aliases: []string{"info", "in", "sh", "show"},
		Short:   "Describe detailed information about ressources",
		Long:    "Will describe all available information about a resource",
	}

	// add list commands
	describeCommand.AddCommand(DescribeRegistryCommand(cfg))
	describeCommand.AddCommand(DescribeToolCommand(cfg))
	describeCommand.AddCommand(DescribeKitCommand(cfg))

	return describeCommand
}
