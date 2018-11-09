/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package cache

import (
	"encoding/json"
	"time"

	"github.com/adobe/sledgehammer/slh/registry"
	bolt "github.com/coreos/bbolt"
)

var (
	// RegistryBucket is the name of the bucket where local versions are cached
	RegistryBucket = "registryUpdateCache"
	// RegistryTTL represents the ttl after that a registry will be seen as outdated and will be updated
	RegistryTTL = 3 * time.Hour
)

// Registry will store the time when the registry was updated the last time
type Registry struct {
	db *bolt.DB
}

func newRegistryCache(db *bolt.DB) Registry {
	return Registry{db: db}
}

// Clear will clear the lastUpdate timestamp of the given registry
func (r *Registry) Clear(reg registry.Registry) error {
	return clear(r.db, RegistryBucket, getRegistryCacheEntry(reg))
}

// LastUpdate will return the last update of the registry.
// If the last update is older than the TTL it will update the registry
func (r *Registry) LastUpdate(reg registry.Registry) (time.Time, error) {
	lastUpdate, err := resolve(
		r.db,
		RegistryBucket,
		getRegistryCacheEntry(reg),
		func(oldValue json.RawMessage) (json.RawMessage, time.Duration, error) {
			lastUpdate := time.Now()
			err := reg.Update()
			if err != nil {
				return nil, RegistryTTL, err
			}
			b, err := json.Marshal(lastUpdate)
			return b, RegistryTTL, err
		},
	)
	if lastUpdate != nil {
		lu := time.Now()
		err := json.Unmarshal(lastUpdate, &lu)
		return lu, err
	}
	return time.Now(), err
}

func getRegistryCacheEntry(reg registry.Registry) string {
	return "registry/" + reg.Data().Name
}
