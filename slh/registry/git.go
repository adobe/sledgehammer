/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package registry

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/adobe/sledgehammer/slh/kit"
	"github.com/adobe/sledgehammer/slh/out"

	"github.com/adobe/sledgehammer/utils/contracts"

	"github.com/adobe/sledgehammer/utils"

	"github.com/sirupsen/logrus"

	"github.com/adobe/sledgehammer/slh/tool"
	"gopkg.in/src-d/go-git.v4"
)

var (
	// ErrorNoValidRepositoryGiven will be thrown if the given repository is not valid
	ErrorNoValidRepositoryGiven = errors.New("Given repository is not valid")
	// ErrorNoIndexInGitRepository will be thrown when the given git repository has no index.json file
	ErrorNoIndexInGitRepository = errors.New("The given git repository has no index.json file")
)

// RegTypeGit is the type of the repository this file handles
const RegTypeGit = "git"

// GitRegistry is the base json structure for all git repositories
type GitRegistry struct {
	Repository string `json:"repository"`
	Core       Data   `json:"core"`
}

// GitFactory is the factory for the GitRegistry
type GitFactory struct{}

// Raw will return a raw GitRepository struct for populating from the db
func (g *GitFactory) Raw() Registry {
	return &GitRegistry{}
}

// Create will take data and return a GitRegistry from the given arguments
func (g *GitFactory) Create(d Data, args []string) (Registry, error) {
	if len(args) != 1 {
		return nil, ErrorNoValidRepositoryGiven
	}
	d.Type = RegTypeGit
	if len(d.Name) == 0 {
		d.Name = strings.Replace(filepath.Base(args[0]), filepath.Ext(args[0]), "", 1)
	}
	reg := &GitRegistry{
		Core:       d,
		Repository: args[0],
	}

	return reg, nil
}

// Data return the data of the registry
func (r *GitRegistry) Data() *Data {
	return &r.Core
}

// Remove will be called when the registry should be removed.
// It will try to delete the path of the cloned registry
func (r *GitRegistry) Remove() error {
	// delete path on system
	logrus.WithField("path", r.Data().Path).Info("Removing at path")
	err := os.RemoveAll(r.Data().Path)
	if err != nil {
		return err
	}
	return nil
}

// Update will do a git pull on the repository to update the internal state
func (r *GitRegistry) Update() error {
	logrus.WithField("path", r.Core.Path).WithField("repository", r.Repository).Info("Pulling repository")
	repo, err := git.PlainOpen(r.Core.Path)
	if err != nil {
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = worktree.Pull(&git.PullOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	return nil
}

// Initialize will do the initial clone for the repository and will verify that there is a index.json
func (r *GitRegistry) Initialize() error {

	absPath, err := filepath.Abs(r.Repository)

	if err != nil {
		return r.initializeRemote()
	}
	st, err := os.Stat(absPath)
	if st == nil || os.IsNotExist(err) {
		return r.initializeRemote()
	}
	if !st.IsDir() {
		return ErrorNoValidRepositoryGiven
	}
	r.Repository = absPath
	return r.initializeLocal()
}

func (r *GitRegistry) initializeLocal() error {
	// git directory is already there locally, use that
	r.Core.Path = r.Repository
	return r.postInitialize()
}

func (r *GitRegistry) initializeRemote() error {
	logrus.WithField("path", r.Core.Path).WithField("repository", r.Repository).Info("Cloning repository")
	_, err := git.PlainClone(r.Core.Path, false, &git.CloneOptions{
		URL: r.Repository,
		// We only need the current status and pull from there on, history is not important
		Depth: 1,
	})
	if err != nil {
		return err
	}
	return r.postInitialize()
}

func (r *GitRegistry) postInitialize() error {
	exists, err := utils.Exists(filepath.Join(r.Core.Path, "index.json"))
	if err != nil {
		return nil
	}

	jsonReg, err := r.readRegistry()
	if err != nil {
		return err
	}
	r.Core.Description = jsonReg.Description
	r.Core.Maintainer = jsonReg.Maintainer

	if !exists {
		return ErrorNoIndexInGitRepository
	}
	return err
}

// Info will show some detailed information about the registry
func (r *GitRegistry) Info(ct *out.Container) {
	ct.Add(out.NewValue("Repository", r.Repository))
}

// Kits will return the tools currently in this registry
func (r *GitRegistry) Kits() ([]kit.Kit, error) {
	// read json
	kits := []kit.Kit{}
	// get content

	registry, err := r.readRegistry()
	if err != nil {
		return nil, err
	}

	for _, v := range registry.Kits {
		df := kit.Kit{
			Description: v.Description,
			Name:        v.Name,
			Tools:       []kit.Tool{},
		}
		for _, t := range v.Tools {
			kt := kit.Tool{
				Alias:   t.Alias,
				Name:    t.Name,
				Version: t.Version,
			}
			df.Tools = append(df.Tools, kt)
		}

		kits = append(kits, df)
	}
	return kits, nil
}

// Tools will return the tools currently in this registry
func (r *GitRegistry) Tools() ([]tool.Tool, error) {
	// read folders
	tools := []tool.Tool{}

	files, err := ioutil.ReadDir(filepath.Join(r.Core.Path, "tools"))
	if err != nil {
		return tools, err
	}

	for _, f := range files {
		if f.IsDir() {
			content, err := ioutil.ReadFile(filepath.Join(r.Core.Path, "tools", f.Name(), "tool.json"))
			if err != nil {
				return tools, err
			}
			v := contracts.Tool{}
			err = json.Unmarshal(content, &v)
			if err != nil {
				return tools, err
			}
			fun, found := tool.Types[v.Type]
			if found {
				df := tool.Data{
					Type:          v.Type,
					Description:   v.Description,
					Image:         v.Image,
					Name:          f.Name(),
					Registry:      r.Core.Name,
					ImageRegistry: v.Registry,
					Entry:         v.Entry,
				}
				if v.Daemon != nil {
					df.Daemon = &tool.Daemon{
						Entry: v.Daemon.Entry,
					}
				}
				tool := fun.Create(df)
				logrus.WithField("tool", tool.Data().Name).Debug("Found tool")
				tools = append(tools, tool)
			}
		}
	}
	return tools, nil
}

func (r *GitRegistry) readRegistry() (*contracts.Registry, error) {
	registry := &contracts.Registry{}

	content, err := ioutil.ReadFile(filepath.Join(r.Core.Path, "index.json"))
	if err != nil {
		return nil, ErrorNoValidPathGiven
	}

	// validate it is a json file
	// validate it can be parsed
	err = json.Unmarshal(content, &registry)
	return registry, err
}
