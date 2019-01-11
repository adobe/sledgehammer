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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/adobe/sledgehammer/utils"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// CreateRootCommand will create the main command that should be used to run the application
func CreateRootCommand(cfg *config.Config) *cobra.Command {

	rootCommand := &cobra.Command{
		Use:   "slh",
		Short: "Sledgehammer is a toolbox for your dockerized tools",
		Long: `Sledgehammer - Dependency isolated executions.
Use dockerized tools like they are installed on the  system.`,
	}

	var logLevel string
	rootCommand.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := validateOutputType(cfg.OutputType); err != nil {
			return err
		}

		initializeLogger(cfg, logLevel, cfg.OutputType)

		if err := configureOutput(cfg); err != nil {
			return err
		}
		if err := setConfigDir(cfg); err != nil {
			return err
		}
		shouldInit, err := utils.ShouldInitialize(cfg)
		if err != nil {
			return err
		}
		if shouldInit {
			if err := addDefaultRegistry(cfg); err != nil {
				return err
			}
			if err := addDefaultMount(cfg); err != nil {
				return err
			}
		}
		return nil
	}

	rootCommand.PersistentFlags().StringVarP(&cfg.OutputType, "output", "o", "text", "Define the output, currently supported is none|text|json")
	rootCommand.PersistentFlags().StringVar(&logLevel, "log-level", "none", "Set the log level (debug|info|warning|error|fatal|panic)")
	rootCommand.PersistentFlags().StringVar(&cfg.ConfigDir, "confdir", cfg.ConfigDir, "Location of configuration directory. Default is bindir/.slh")

	// add all other commands
	rootCommand.AddCommand(GetCommand(cfg))
	rootCommand.AddCommand(CreateCommand(cfg))
	rootCommand.AddCommand(DeleteCommand(cfg))
	rootCommand.AddCommand(InstallCommand(cfg))
	rootCommand.AddCommand(ResetCommand(cfg))
	rootCommand.AddCommand(DescribeCommand(cfg))
	rootCommand.AddCommand(RunCommand(cfg))
	rootCommand.AddCommand(RunAliasCommand(cfg))
	rootCommand.AddCommand(UpdateCommand(cfg))

	rootCommand.SetOutput(cfg.IO.Out)

	rootCommand.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		return output(cfg)
	}

	rootCommand.Version = version.Version
	rootCommand.SetVersionTemplate(`{{printf "%s" .Version}}
`)
	return rootCommand
}

// Execute will execute the root command
func Execute(cfg *config.Config) {
	cmd := CreateRootCommand(cfg)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	if cfg.Output != nil {
		os.Exit(cfg.Output.ExitCode)
	}
}

func configureOutput(cfg *config.Config) error {
	if cfg.Output == nil {
		cfg.Output = config.NewOutput(cfg)
	}
	return nil
}

func validateOutputType(output string) error {
	validOutputs := []string{"none", "text", "json"}
	match := false
	for _, v := range validOutputs {
		if v == output {
			match = true
		}
	}
	if !match {
		return fmt.Errorf("The output type %s does not match %v", output, validOutputs)
	}
	return nil
}

func initializeLogger(cfg *config.Config, logLevel string, output string) {
	//  init logrus
	logrus.SetOutput(cfg.IO.Out)

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.SetLevel(logrus.PanicLevel)
	} else {
		logrus.SetLevel(level)
	}
	if output == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	if output == "none" {
		logrus.SetOutput(ioutil.Discard)
	}
}

func setConfigDir(cfg *config.Config) error {
	if len(cfg.ConfigDir) == 0 {
		path, _ := utils.ExecutablePath()
		cfg.ConfigDir = filepath.Join(path, ".slh")
	}

	cfg.ConfigDir = utils.ImportPath(cfg.ConfigDir)
	logrus.WithField("confdir", cfg.ConfigDir).Info("Using configuration directory")
	return nil
}

func output(cfg *config.Config) error {
	// render the output now
	if cfg.Output != nil && cfg.OutputType != "none" {
		if cfg.Output.Element != nil {
			cfg.Output.Render()
		} else {
			logrus.Warn("No output defined!")
		}
	}
	return nil
}
