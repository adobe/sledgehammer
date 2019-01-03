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

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/tool"
	"github.com/coreos/bbolt"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/sirupsen/logrus"
)

var (
	// DaemonContainerBucket is the name of the bucket where the ids of daemon containers are cached
	DaemonContainerBucket = "DaemonContainerCache"
	// ContainerIDTTL represents the time after that a daemon container will be renewed
	ContainerIDTTL = 10 * time.Minute
)

// Container represents the container cache. If the tool is daemonized then it can be that the daemon is already running.
// Normally it is then cached in the database.
// This provides a fast way to detect running daemons to a certain degree
type Container struct {
	db *bolt.DB
}

func newContainerCache(db *bolt.DB) Container {
	return Container{db: db}
}

func (c *Container) ClearAll() error {
	return c.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(DaemonContainerBucket))
	})
}

// Clear will clear the entry of the given tool.
// This is useful if you want to clear things
func (c *Container) Clear(to tool.Tool, tag string) error {
	return clear(c.db, DaemonContainerBucket, getContainerCacheEntry(to, tag))
}

// CurrentDaemons will return all currently cached daemon tools.
// Useful to kill all currently running daemons.
func (c *Container) CurrentDaemons() (map[string]string, error) {
	logrus.Debugln("Getting all currently running daemons")
	containerIDs := map[string]string{}

	err := c.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(DaemonContainerBucket))
		if err != nil {
			return err
		}
		if bucket != nil {
			err = bucket.ForEach(func(key []byte, value []byte) error {
				var item cacheItem
				err := json.Unmarshal(value, &item)
				if err != nil {
					return err
				}
				id := ""
				err = json.Unmarshal(item.Item, &id)
				if err != nil {
					return err
				}
				containerIDs[string(key)] = id
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return containerIDs, err
	}
	return containerIDs, nil
}

// Get will return the id of the running container if possible
func (c *Container) Get(to tool.Tool, tag string, cfg *config.Config, mos []string) (string, error) {
	logrus.WithField("entry", getContainerCacheEntry(to, tag)).Debugln("Check container")
	id, err := resolve(
		c.db,
		DaemonContainerBucket,
		getContainerCacheEntry(to, tag),
		func(oldValue json.RawMessage) (json.RawMessage, time.Duration, error) {
			if oldValue != nil {
				// shutdown container if possible
				var id string
				err := json.Unmarshal(oldValue, &id)
				if err == nil {
					cfg.Docker.Docker.RemoveContainer(docker.RemoveContainerOptions{
						Force: true,
						ID:    id,
					})
				}
			}
			id, err := tool.StartIfDaemon(&tool.ExecutionOptions{
				Tool:    to,
				Mounts:  mos,
				IO:      cfg.IO,
				Docker:  &cfg.Docker,
				Version: tag,
			})
			if err != nil {
				return json.RawMessage{}, ContainerIDTTL, err
			}
			b, err := json.Marshal(id)
			return b, ContainerIDTTL, err
		},
	)
	if err != nil {
		return "", err
	}

	if id != nil {
		pID := ""
		err := json.Unmarshal(id, &pID)
		return pID, err
	}
	return "", err
}

// GetContainerCacheEntry will return the name of the cache entry for the given tool and version
func getContainerCacheEntry(to tool.Tool, version string) string {
	return "container/" + to.Data().Registry + "/" + to.Data().Name + "/" + to.Data().Image + ":" + version
}
