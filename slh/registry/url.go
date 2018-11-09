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
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/adobe/sledgehammer/slh/kit"
	"github.com/adobe/sledgehammer/slh/out"
	"github.com/adobe/sledgehammer/slh/tool"
	"github.com/adobe/sledgehammer/utils/contracts"
	"github.com/sirupsen/logrus"
)

var (
	// ErrorNoValidURLGiven will be thrown if the given URL is not valid
	ErrorNoValidURLGiven = errors.New("No valid URL given")
)

// RegTypeURL is the type of the repository of this file
const RegTypeURL = "url"

// URLRegistry is the JSON representation of the registry of this file
type URLRegistry struct {
	URL  string `json:"url"`
	Core Data   `json:"core"`
}

// URLFactory is the factory for the URLRegistry
type URLFactory struct{}

// Raw will return a raw GitRepository struct for populating from the db
func (g *URLFactory) Raw() Registry {
	return &URLRegistry{}
}

// Create will take data and return a URLRegistry from the given arguments
func (g *URLFactory) Create(d Data, args []string) (Registry, error) {
	logrus.WithField("args", args).Debugln("Creating URL registry")
	if len(args) != 1 {
		return nil, ErrorNoValidURLGiven
	}
	d.Type = RegTypeURL
	if len(d.Name) == 0 {
		ur, err := url.Parse(args[0])
		if err != nil {
			return nil, ErrorNoValidURLGiven
		}
		d.Name = strings.Replace(filepath.Base(ur.Host+"/"+ur.Path), filepath.Ext(ur.Host+"/"+ur.Path), "", 1)
	}
	return &URLRegistry{
		Core: d,
		URL:  args[0],
	}, nil
}

// Remove will be called when the registry should be removed.
// It will do nothing as of now...
func (r *URLRegistry) Remove() error {
	// delete path on system
	logrus.WithField("path", r.Data().Path).Info("Removing at path")
	err := os.RemoveAll(r.Data().Path)
	if err != nil {
		return err
	}
	return nil
}

// Data return the data of the registry
func (r *URLRegistry) Data() *Data {
	return &r.Core
}

// Update will do a git pull on the repository to update the internal state
func (r *URLRegistry) Update() error {
	// same as initialize
	logrus.WithField("path", filepath.Join(r.Core.Path, "index.json")).Debugln("Updating URL registry")
	return r.downloadFile(filepath.Join(r.Core.Path, "index.json"), r.URL)
}

// Initialize will download the registry from the URL and places it on the local file system
func (r *URLRegistry) Initialize() error {
	logrus.WithField("path", filepath.Join(r.Core.Path, "index.json")).Debugln("Initializing URL registry")
	err := os.MkdirAll(r.Core.Path, 0777)
	if err != nil {
		return err
	}
	// Download and store in path as index.json
	err = r.downloadFile(filepath.Join(r.Core.Path, "index.json"), r.URL)
	if err != nil {
		return nil
	}
	jsonReg, err := r.readRegistry()
	if err != nil {
		return err
	}
	r.Core.Description = jsonReg.Description
	r.Core.Maintainer = jsonReg.Maintainer
	return nil
}

// Info will show some detailed information about the registry
func (r *URLRegistry) Info(ct *out.Container) {
	ct.Add(out.NewValue("URL", r.URL))
}

// Tools will return the tools currently in this registry
func (r *URLRegistry) Tools() ([]tool.Tool, error) {
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

// Kits will return the tools currently in this registry
func (r *URLRegistry) Kits() ([]kit.Kit, error) {
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

func (r *URLRegistry) readRegistry() (*contracts.Registry, error) {
	registry := &contracts.Registry{}
	logrus.WithField("path", filepath.Join(r.Core.Path, "index.json")).Debugln("Reading URL registry path")
	content, err := ioutil.ReadFile(filepath.Join(r.Core.Path, "index.json"))
	if err != nil {
		return nil, ErrorNoValidPathGiven
	}

	// validate it is a json file
	// validate it can be parsed
	err = json.Unmarshal(content, &registry)
	return registry, err
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func (u *URLRegistry) downloadFile(filepath string, url string) error {

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(filepath, body, 0777)
	}
	return nil
}
