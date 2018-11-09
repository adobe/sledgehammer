/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package alias

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/adobe/sledgehammer/utils"

	"github.com/sirupsen/logrus"

	"github.com/adobe/sledgehammer/slh/config"
	bolt "github.com/coreos/bbolt"
)

var (
	// BucketKey is the name of the bucket where aliases are stored
	BucketKey = "alias"
	// ErrorNotFound will be thrown if the alias can not be found
	ErrorNotFound = errors.New("Alias not found")
	// ErrorNoName will be thrown if the alias has no name
	ErrorNoName = errors.New("An alias has no name")
	// ErrorDuplicateAlias will be thrown if there is already an alias with that name
	ErrorDuplicateAlias = errors.New("The alias already exist")
	// ErrorFileAlreadyPresent will be thrown if the symlink cannot be created due to an existing file with the same name
	ErrorFileAlreadyPresent = errors.New("The symlink could not be created because there is already a file with that name")
	// ErrorNaughtyBoy will be thrown it the user tries to install a tool with the slh alias
	ErrorNaughtyBoy = errors.New("It is not recommended to alias any tool with 'slh'...")
)

// Alias is a struct that will be used when the user installs a tool. That will create an alias that can be used as a shortcut.
type Alias struct {
	Name     string `json:"name"`
	Registry string `json:"registry"`
	Tool     string `json:"tool"`
	Version  string `json:"version"`
}

// Aliases is the main access point for adding/removing/editing aliases
type Aliases struct {
	config.Database
}

// New will create a new Aliases struct based on the given bolt database.
// It offers methods to add/remove and list all aliases registered with Sledgehammer
func New(db config.Database) *Aliases {
	return &Aliases{
		Database: db,
	}
}

// List will return the current aliases
func (m *Aliases) List() ([]Alias, error) {
	aliases := []Alias{}
	err := m.DB.View(func(tx *bolt.Tx) error {
		// Get the alias bucket
		bucket := tx.Bucket([]byte(BucketKey))
		if bucket != nil {
			err := bucket.ForEach(func(key []byte, value []byte) error {
				dbAlias := Alias{}
				err := json.Unmarshal(value, &dbAlias)
				if err != nil {
					return err
				}
				logrus.WithField("alias", dbAlias.Name).Debug("Found alias")
				aliases = append(aliases, dbAlias)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return aliases, err
}

// Add will add the given alias to sledgehammer.
// It will overwrite any entry with the same name.
func (m *Aliases) Add(alias Alias) error {
	if len(alias.Name) == 0 {
		return ErrorNoName
	}
	if alias.Name == "slh" {
		return ErrorNaughtyBoy
	}
	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BucketKey))
		if err != nil {
			return err
		}
		if bucket != nil {
			b, err := json.Marshal(alias)
			if err != nil {
				return err
			}
			logrus.WithField("alias", alias.Name).Debug("Added alias")
			return bucket.Put([]byte(alias.Name), b)
		}
		return nil
	})
	return err
}

// Get will get a single alias if available
func (m *Aliases) Get(name string) (*Alias, error) {
	logrus.WithField("alias", name).Debug("Getting alias")
	alias := Alias{}
	err := m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketKey))
		if bucket != nil {
			dbAlias := bucket.Get([]byte(name))
			if dbAlias != nil {
				return json.Unmarshal(dbAlias, &alias)
			}
		}
		return ErrorNotFound
	})
	return &alias, err
}

// Get will get all alias for a given tool
func (m *Aliases) From(name string, registry string) ([]Alias, error) {
	logrus.WithField("alias", name).WithField("registry", registry).Debug("Getting alias for a given tool")
	matchingAliases := []Alias{}
	aliases, err := m.List()
	if err != nil {
		return aliases, err
	}
	for _, al := range aliases {
		logrus.WithField("registry", al.Registry).WithField("tool", al.Tool).Debug("Checking tool")
		if al.Registry == registry && al.Tool == name {
			matchingAliases = append(matchingAliases, al)
		}
	}
	return matchingAliases, nil
}

// Has will check if a single alias is present in the database
func (m *Aliases) Has(name string) (bool, error) {
	logrus.WithField("alias", name).Debug("Checking alias")
	found := false
	err := m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketKey))
		if bucket != nil {
			dbAlias := bucket.Get([]byte(name))
			if dbAlias != nil {
				found = true
				return nil
			}
		}
		return nil
	})
	return found, err
}

// Remove will try to remove the given alias and returns an error if any problem occurs.
func (m *Aliases) Remove(name string) error {
	logrus.WithField("alias", name).Debug("Removing alias")
	return m.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketKey))
		if bucket != nil {
			return bucket.Delete([]byte(name))
		}
		return nil
	})
}

// HasSymlink will check if the given symlink is already present
func HasSymlink(name string) (bool, error) {
	path, err := utils.ExecutablePath()
	if err != nil {
		return false, err
	}
	absPath := filepath.Join(path, name)
	// fmt.Println(absPath)
	return utils.Exists(absPath)
}

// CreateSymlink will create the given symlink if possible
func CreateSymlink(name string) error {
	path, err := utils.ExecutablePath()
	if err != nil {
		return err
	}
	executableName, err := utils.ExecutableName()
	if err != nil {
		return err
	}
	absPath := filepath.Join(path, name)
	// fmt.Println(absPath)
	return os.Symlink(executableName, absPath)
}

// RemoveSymlink will remove the given symlink if possible
func RemoveSymlink(name string) error {
	path, err := utils.ExecutablePath()
	if err != nil {
		return err
	}
	absPath := filepath.Join(path, name)
	return os.RemoveAll(absPath)
}
