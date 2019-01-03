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

	"github.com/coreos/bbolt"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/tool"
	"github.com/sirupsen/logrus"
)

var (
	// LocalVersionBucket is the name of the bucket where local versions are cached
	LocalVersionBucket = "localVersionCache"
	// RemoteVersionBucket is the name of the bucket where remote versions are cached
	RemoteVersionBucket = "remoteVersionCache"
	// RemoteTimeToLive is the duration that remote versions are valid.
	// Effectively this means we will check after the duration if a new version is available
	RemoteTimeToLive = 12 * time.Hour
	// LocalTimeToLive is the duration that local versions are valid.
	// Effectively this means we will check after the duration if a new version is available
	LocalTimeToLive = 1 * time.Minute
)

// Version is the struct to work with cached versions
type Version struct {
	db *bolt.DB
}

func newVersionCache(db *bolt.DB) Version {
	return Version{db: db}
}

func (c *Version) ClearAll() error {
	return c.db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(LocalVersionBucket))
		if err != nil {
			return err
		}
		return tx.DeleteBucket([]byte(RemoteVersionBucket))
	})
}

// Clear will clear the cached local and remote versions of this tool
func (v *Version) Clear(to tool.Tool) error {
	err := clear(v.db, LocalVersionBucket, getLocalVersionCacheEntry(to))
	if err != nil {
		return err
	}
	return clear(v.db, RemoteVersionBucket, getRemoteVersionCacheEntry(to))
}

// Local will return the remote versions of a tool from the cache if possible
func (v *Version) Local(to tool.Tool, client config.Docker) ([]string, error) {
	logrus.WithField("entry", getLocalVersionCacheEntry(to)).Debugln("Checking local versions")
	versions, err := resolve(
		v.db,
		LocalVersionBucket,
		getLocalVersionCacheEntry(to),
		func(oldValue json.RawMessage) (json.RawMessage, time.Duration, error) {
			versions, err := tool.Versions(client, to)
			if err != nil {
				return nil, LocalTimeToLive, err
			}
			b, err := json.Marshal(versions)
			return b, LocalTimeToLive, err
		},
	)
	if versions != nil {
		sVersions := []string{}
		err := json.Unmarshal(versions, &sVersions)
		return sVersions, err
	}
	return nil, err
}

// Remote will return the remote versions of a tool from the cache if possible
func (v *Version) Remote(to tool.Tool) ([]string, error) {
	logrus.WithField("entry", getRemoteVersionCacheEntry(to)).Debugln("Checking remote versions")
	versions, err := resolve(
		v.db,
		RemoteVersionBucket,
		getRemoteVersionCacheEntry(to),
		func(oldValue json.RawMessage) (json.RawMessage, time.Duration, error) {
			versions, err := to.Versions()
			if err != nil {
				return nil, RemoteTimeToLive, err
			}
			b, err := json.Marshal(versions)
			return b, RemoteTimeToLive, err
		},
	)
	if versions != nil {
		sVersions := []string{}
		err := json.Unmarshal(versions, &sVersions)
		return sVersions, err
	}
	return nil, err
}

func getLocalVersionCacheEntry(to tool.Tool) string {
	return "local/" + to.Data().Registry + "/" + to.Data().Name
}

func getRemoteVersionCacheEntry(to tool.Tool) string {
	return "remote/" + to.Data().Registry + "/" + to.Data().Name
}
