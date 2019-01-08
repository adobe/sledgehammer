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
	"time"

	"github.com/fsouza/go-dockerclient"

	"github.com/adobe/sledgehammer/slh/cache"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/mount"
	"github.com/adobe/sledgehammer/slh/tool"
	"github.com/adobe/sledgehammer/slh/version"
	"github.com/adobe/sledgehammer/utils"
	bolt "github.com/coreos/bbolt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// ErrorNoVersionFound will be thrown if there is no version that can be run
	ErrorNoVersionFound = errors.New("Could not find a version to run")
	// DefaultTimeout is the timeout used to pull an image
	DefaultTimeout = 300 * time.Second
)

type RunCmd struct {
	registry  string
	tool      string
	version   string
	arguments []string
	update    bool
}

func RunCommand(cfg *config.Config) *cobra.Command {
	runCmd := RunCmd{}
	runCommand := &cobra.Command{
		Use:   "run <tool>",
		Short: "Run an installed tool",
		Long:  "Will run the given tool with the provided arguments",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runCmd.registry, runCmd.tool = utils.GetRegistryAndTool(args[0])
			err := runCmd.Execute(cfg)
			if err != nil {
				cmd.SilenceUsage = true
			}
			return err
		},
	}

	runCommand.Flags().BoolVar(&runCmd.update, "update", false, "Will force the download of a newer remote image if available before running the tool")
	runCommand.Flags().StringVar(&runCmd.version, "version", "", "The version constraint that the tool should fulfill")
	runCommand.Flags().StringSliceVarP(&runCmd.arguments, "arguments", "a", []string{}, "The arguments to pass to the tool")

	runCommand.Hidden = true

	return runCommand
}

// Execute will run the given tool
func (r *RunCmd) Execute(cfg *config.Config) error {

	var exitCode int

	database, err := cfg.OpenDatabase()
	if err != nil {
		return err
	}

	// get tool
	tools := tool.New(config.Database{DB: database})
	mounts := mount.New(config.Database{DB: database})
	caches := cache.New(config.Database{DB: database})

	to, err := tools.Get(r.registry, r.tool)
	if err != nil {
		return err
	}
	mos, err := mounts.List()
	if err != nil {
		return err
	}

	pullDone := make(chan error, 1)
	closeDB := make(chan bool, 1)

	version, err := r.selectVersion(cfg.Docker, database, to, pullDone, closeDB)
	if err != nil {
		return err
	}

	containerID, err := caches.Container.Get(to, version, cfg, mos)
	if err != nil {
		return err
	}
	<-closeDB
	logrus.Info("Closing database")
	// close db
	cfg.CloseDatabase()

	if len(version) == 0 {
		return ErrorNoVersionFound
	}

	arguments := r.arguments
	if len(arguments) == 1 && len(arguments[0]) == 0 {
		arguments = []string{}
	}

	executionOptions := &tool.ExecutionOptions{
		IO:        cfg.IO,
		Docker:    &cfg.Docker,
		Tool:      to,
		Version:   version,
		Arguments: r.arguments,
		Mounts:    mos,
	}

	if containerID == "" {
		logrus.Info("Starting and executing tool")
		exitCode, err = tool.StartAndExecute(executionOptions)
		if err != nil {
			return err
		}
	} else {
		logrus.Debug("Tool has been started as daemon, executing")
		exitCode, err = tool.Execute(containerID, executionOptions)
		logrus.Debugln("executing done...")
		if err != nil {
			if _, ok := err.(*docker.NoSuchContainer); ok {
				// container not found, clear cache for this entry and start daemon
				// open a new database connection because nothing else will help...
				database, err := cfg.OpenDatabase()
				if err != nil {
					return err
				}
				defer cfg.CloseDatabase()
				caches := cache.New(config.Database{DB: database})
				caches.Container.Clear(to, version)
				containerID, err := caches.Container.Get(to, version, cfg, mos)
				if err != nil {
					return err
				}
				database.Close()
				exitCode, err = tool.Execute(containerID, executionOptions)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	// Wait until potential pulls are done, otherwise the pull will stop midexecution -> race condition
	err = <-pullDone
	logrus.Info("Potential pull done")

	cfg.Output.ExitCode = exitCode
	return err
}

// selectVersion will select the version that should be used to run the tool.
// It will take the constraint into consideration and will pull the image if needed.
func (r *RunCmd) selectVersion(client config.Docker, db *bolt.DB, to tool.Tool, doneChan chan error, closeDBChan chan bool) (string, error) {

	// create channels to fetch the image locally and from the registry
	localVersionChannel := make(chan func() (string, error))
	repositoryVersionChannel := make(chan func() (string, error))

	caches := cache.New(config.Database{DB: db})

	// Fetch local versions
	go r.fetchLocalVersions(localVersionChannel, db, client, to)
	// Fetch repository versions
	go r.FetchRepositoryVersions(repositoryVersionChannel, db, to)

	if r.update {
		localVersion, err := (<-localVersionChannel)()
		if err != nil {
			doneChan <- err
			closeDBChan <- true
			return "", err
		}
		repositoryVersion, err := (<-repositoryVersionChannel)()
		if err != nil {
			doneChan <- err
			closeDBChan <- true
			return "", err
		}
		if version.ShouldPull(localVersion, repositoryVersion) {
			caches.Versions.Clear(to)
			closeDBChan <- true
			err = tool.Pull(client, to, repositoryVersion, DefaultTimeout)
			if err != nil {
				doneChan <- err
				closeDBChan <- true
				return "", err
			}
			doneChan <- nil
			return repositoryVersion, nil
		}
		closeDBChan <- true
		doneChan <- nil
		return localVersion, nil
	}
	localVersion, err := (<-localVersionChannel)()
	if err != nil {
		doneChan <- err
		closeDBChan <- true
		return "", err
	}
	pull := func(doneChan chan error) (string, error) {
		logrus.Infoln("Fetching remote versions")
		repositoryVersion, err := (<-repositoryVersionChannel)()
		dbClosed := false
		if err == nil && version.ShouldPull(localVersion, repositoryVersion) {
			logrus.WithField("remoteVersion", repositoryVersion).Infoln("Should pull remote version")
			caches.Versions.Clear(to)
			closeDBChan <- true
			dbClosed = true
			err = tool.Pull(client, to, repositoryVersion, DefaultTimeout)
			if err != nil {
				doneChan <- err
				return "", err
			}
		}
		if !dbClosed {
			closeDBChan <- true
		}
		doneChan <- err
		return repositoryVersion, err
	}
	if len(localVersion) > 0 {
		logrus.WithField("version", localVersion).Infoln("Found local image")
		go pull(doneChan)
		return localVersion, nil
	}
	return pull(doneChan)
}

// fetchLocalVersions will fetch the local version asynchronously
func (r *RunCmd) fetchLocalVersions(c chan func() (string, error), db *bolt.DB, client config.Docker, to tool.Tool) {
	c <- (func() (string, error) {
		// get local versions
		caches := cache.New(config.Database{DB: db})
		versions, err := caches.Versions.Local(to, client)
		logrus.WithField("localversions", versions).Infoln("Found local versions")
		if err != nil {
			return "", err
		}
		return version.Select(versions, r.version), nil
	})
}

// FetchRepositoryVersions will fetch the remote versions for the tool
func (r *RunCmd) FetchRepositoryVersions(c chan func() (string, error), db *bolt.DB, to tool.Tool) {
	c <- (func() (string, error) {
		caches := cache.New(config.Database{DB: db})
		versions, err := caches.Versions.Remote(to)
		logrus.WithField("versions", versions).Infoln("Found remote versions")
		if err != nil {
			return "", err
		}
		return version.Select(versions, r.version), nil
	})
}
