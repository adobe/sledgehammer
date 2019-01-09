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
	"path/filepath"
	"strings"

	"github.com/adobe/sledgehammer/slh/out"

	"github.com/adobe/sledgehammer/slh/kit"

	"github.com/adobe/sledgehammer/slh/tool"
	"github.com/adobe/sledgehammer/utils/contracts"
	"github.com/sirupsen/logrus"
)

var (
	// ErrorEmptyPathGiven will be thrown if the given path is not valid
	ErrorEmptyPathGiven = errors.New("No path given... call it with 'file <path>'")
	// ErrorNoValidPathGiven will be thrown if the given path is not valid
	ErrorNoValidPathGiven = errors.New("The given path is not valid")
)

// RegTypeLocal is the type of this registry
const RegTypeLocal = "local"

// RegTypeFile is the type of this registry
const RegTypeFile = "file"

// FileRegistry represents a registry where the tools are stored on a local filesystem.
// This is mostly useful for development of tools, registries and sledgehammer itself.
type FileRegistry struct {
	Location string `json:"location"`
	Core     Data   `json:"core"`
}

// FileFactory is the factory for the FileRegistry
type FileFactory struct{}

// Raw will return a raw GitRepository struct for populating from the db
func (g *FileFactory) Raw() Registry {
	return &FileRegistry{}
}

// Remove will be called when the registry should be removed.
// It will actually do nothing as we do not touch the original file
func (r *FileRegistry) Remove() error {
	// do nothing
	return nil
}

// Create will take data and return a FileRegistry from the given arguments
func (g *FileFactory) Create(d Data, args []string) (Registry, error) {
	if len(args) == 0 {
		return nil, ErrorEmptyPathGiven
	}
	if len(args) != 1 {
		return nil, ErrorNoValidPathGiven
	}
	absPath, err := filepath.Abs(args[0])
	if err != nil {
		return nil, ErrorNoValidPathGiven
	}

	if len(d.Name) == 0 {
		d.Name = strings.Replace(filepath.Base(args[0]), filepath.Ext(args[0]), "", 1)
	}
	d.Type = RegTypeFile
	reg := &FileRegistry{
		Core:     d,
		Location: absPath,
	}

	return reg, nil
}

// Tools will return all tools that are stored in the json file that this registry points to
func (r FileRegistry) Tools() ([]tool.Tool, error) {
	tools := []tool.Tool{}

	registry, err := r.readRegistry()
	if err != nil {
		return tools, err
	}

	for _, v := range registry.Tools {
		fun, found := tool.Types[v.Type]
		if found {
			df := tool.Data{
				Type:          v.Type,
				Description:   v.Description,
				Image:         v.Image,
				Name:          v.Name,
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
	return tools, nil
}

// Kits will return all kits that the are stated in the registry.
func (r FileRegistry) Kits() ([]kit.Kit, error) {
	kits := []kit.Kit{}

	// reg, err := r.readRegistry()
	reg, err := r.readRegistry()
	if err != nil {
		return kits, err
	}

	for _, v := range reg.Kits {
		df := kit.Kit{
			Description: v.Description,
			Name:        v.Name,
			Tools:       []kit.Tool{},
		}

		for _, tk := range v.Tools {
			t := kit.Tool{
				Alias:   tk.Alias,
				Name:    tk.Name,
				Version: tk.Version,
			}
			df.Tools = append(df.Tools, t)
		}
		logrus.WithField("kit", df.Name).WithField("tools", df.Tools).Debug("Found kit")
		kits = append(kits, df)
	}
	return kits, nil
}

// Data return the data of the registry
func (r *FileRegistry) Data() *Data {
	return &r.Core
}

// Info will show some detailed information about the registry
func (r *FileRegistry) Info(ct *out.Container) {
	ct.Add(out.NewValue("Location", r.Location))
}

// Update will do nothing as the file is already on the disk, we just need to read it again
func (r *FileRegistry) Update() error {
	return nil
}

// Initialize will do nothing for a file registry
func (r *FileRegistry) Initialize() error {
	jsonReg, err := r.readRegistry()
	if err != nil {
		return err
	}
	r.Core.Description = jsonReg.Description
	r.Core.Maintainer = jsonReg.Maintainer
	return nil
}

func (r *FileRegistry) readRegistry() (*contracts.Registry, error) {
	registry := &contracts.Registry{}

	if len(r.Location) == 0 {
		return nil, ErrorNoValidPathGiven
	}
	logrus.WithField("location", r.Location).Debug("Retreiving tools from file registry")

	// read json
	// get content
	content, err := ioutil.ReadFile(r.Location)
	if err != nil {
		return nil, ErrorNoValidPathGiven
	}

	// validate it is a json file
	// validate it can be parsed
	err = json.Unmarshal(content, registry)
	return registry, err
}
