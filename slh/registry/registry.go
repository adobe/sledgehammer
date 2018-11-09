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
	"time"

	"github.com/sirupsen/logrus"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/kit"
	"github.com/adobe/sledgehammer/slh/out"
	"github.com/adobe/sledgehammer/slh/tool"
	bolt "github.com/coreos/bbolt"
)

var (
	// BucketKey is the key under which all registries are stored in the database
	BucketKey = "registries"
	// Types is a factory function to bootstrap different registries
	Types = map[string]Factory{
		RegTypeGit:   &GitFactory{},
		RegTypeLocal: &LocalFactory{},
		RegTypeURL:   &URLFactory{},
	}
	// ErrorNoName will be thrown if a registry has no name, which is required
	ErrorNoName = errors.New("A registry has no name")
	// ErrorAlreadyExists will be thrown if the registry already exists
	ErrorAlreadyExists = errors.New("Registry already exists")
	// ErrorRegistryNotFound will be thrown if the given registry cannot be found in the database
	ErrorRegistryNotFound = errors.New("Registry not found")
	// RegistryUpdateInterval is the interval after that a registry will automatically update their tools. This can be forced
	RegistryUpdateInterval = 24 * time.Hour
	// DefaultRegistryURL is the URL of the default regsistry. It will be added on the first run only.
	DefaultRegistryURL = "https://github.com/adobe/sledgehammer-registry.git"
)

// Factory is a factory to create registries from the db and from user input
type Factory interface {
	Raw() Registry
	Create(d Data, args []string) (Registry, error)
}

// Data contains the base data for a registry
type Data struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Path        string `json:"path"`
	Maintainer  string `json:"maintainer"`
	Description string `json:"description"`
}

// Registry is the base interface of a registry, can be extended
type Registry interface {
	Data() *Data
	// Update will be called when the registry should update their local files.
	Update() error
	// Initialize will be called when the registry should be initialized the first time.
	Initialize() error
	// Remove will be called when the registry should be removed. All cleanups should happen here.
	Remove() error
	Tools() ([]tool.Tool, error)
	Kits() ([]kit.Kit, error)
	// Info requires the registry to add their own information to the given container
	Info(ct *out.Container)
}

// JSON is the base json structure that will be stored in the database.
// We need to append the type of the registry so that we can parse them back correctly from the database.
type JSON struct {
	Type     string          `json:"type"`
	Registry json.RawMessage `json:"registry"`
}

// Registries is the struct that can be used to access the registries registered with Sledgehammer
type Registries struct {
	config.Database
}

// New will create a new Registries struct to access the repositories registered with Sledgehammer
func New(db config.Database) *Registries {
	return &Registries{
		Database: db,
	}
}

// Exists will check if a given registry already exists
func (r *Registries) Exists(name string) (bool, error) {
	registries, err := r.List()
	if err != nil {
		return false, err
	}
	for _, reg := range registries {
		if reg.Data().Name == name {
			logrus.WithField("registry", name).Debug("Registry exists")
			return true, nil
		}
	}
	logrus.WithField("registry", name).Debug("Registry does not exists")
	return false, nil
}

// Remove will remove the given registry and all containing tools from the database
func (r *Registries) Remove(name string) error {
	logrus.WithField("registry", name).Debug("Removing registry")
	if len(name) == 0 {
		return ErrorNoName
	}
	tools := tool.New(r.Database)
	reg, err := r.Get(name)
	if err != nil {
		return err
	}
	toolsList, err := tools.From(name)
	if err != nil {
		return err
	}

	for _, to := range toolsList {
		tools.Remove(to.Data().Registry, to.Data().Name)
	}

	err = reg.Remove()
	if err != nil {
		return err
	}

	err = r.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BucketKey))
		if err != nil {
			return err
		}
		if bucket != nil {
			bucket.Delete([]byte(name))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Get will return a single registry given by the name
func (r *Registries) Get(name string) (Registry, error) {
	logrus.WithField("registry", name).Debug("Getting single registry")
	var reg Registry
	err := r.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BucketKey))
		if err != nil {
			return err
		}
		if bucket != nil {
			jsonReg := bucket.Get([]byte(name))
			if jsonReg != nil {
				var m JSON
				err := json.Unmarshal(jsonReg, &m)
				if err != nil {
					return err
				}
				fun, found := Types[m.Type]
				if found {
					reg = fun.Raw()
					if err != nil {
						return err
					}
					err := json.Unmarshal(m.Registry, &reg)
					if err != nil {
						return err
					}
					logrus.WithField("registry", reg.Data().Name).Debug("Found registry")
					return nil
				}
			}
		}
		return ErrorRegistryNotFound
	})
	return reg, err
}

// List will return all registries that are registered with Sledgehammer
func (r *Registries) List() ([]Registry, error) {
	logrus.Debug("Listing all registries")
	registries := []Registry{}
	err := r.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketKey))
		if bucket != nil {
			bucket.ForEach(func(_, v []byte) error {
				var m JSON
				err := json.Unmarshal(v, &m)
				if err != nil {
					return err
				}
				fun, found := Types[m.Type]
				if found {
					p := fun.Raw()
					if err != nil {
						return err
					}
					err := json.Unmarshal(m.Registry, &p)
					if err != nil {
						return err
					}
					// After creating our struct, we should save it
					registries = append(registries, p)
					logrus.WithField("registry", p.Data().Name).Debug("Found registry")
				}
				if !found {
					return errors.New("Registry type not found: '" + m.Type + "'")
				}
				return nil
			})
		}
		return nil
	})
	logrus.WithField("registries", registries).Debug("Found registries")
	return registries, err
}

// Add adds the given registry to the registry files
func (r *Registries) Add(registry Registry) error {
	if len(registry.Data().Name) == 0 {
		return ErrorNoName
	}
	exists, err := r.Exists(registry.Data().Name)
	if err != nil {
		return err
	}
	if exists {
		return ErrorAlreadyExists
	}
	logrus.WithField("registry", registry.Data().Name).Debug("Adding registry")

	err = r.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BucketKey))
		if err != nil {
			return err
		}
		if bucket != nil {
			jsonReg, err := json.Marshal(registry)
			if err != nil {
				return err
			}
			jsonData, err := json.Marshal(JSON{
				Registry: jsonReg,
				Type:     registry.Data().Type,
			})
			if err != nil {
				return err
			}
			return bucket.Put([]byte(registry.Data().Name), jsonData)
		}
		return err
	})
	if err != nil {
		return err
	}

	tools := tool.New(r.Database)
	toolsList, err := registry.Tools()
	if err != nil {
		return err
	}
	// add the tools
	return tools.Add(toolsList...)
}
