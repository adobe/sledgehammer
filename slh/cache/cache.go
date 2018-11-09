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
	"errors"
	"time"

	"github.com/adobe/sledgehammer/slh/config"
	bolt "github.com/coreos/bbolt"
	"github.com/sirupsen/logrus"
)

var (
	// ErrorNameRequired will be thrown, if Cache Entry key "Name" not set
	ErrorNameRequired = errors.New("Cache Entry Name not set")
)

// Cache is a struct that offers access to any cacheable item.
type Cache struct {
	Versions  Version
	Container Container
	Registry  Registry
}

// CachedVersion is the data type that will be stored in the database
type cacheItem struct {
	ValidUntil int64           `json:"validUntil"`
	Item       json.RawMessage `json:"item"`
}

// New will return a new cache instance
func New(db config.Database) *Cache {
	return &Cache{
		Versions:  newVersionCache(db.DB),
		Container: newContainerCache(db.DB),
		Registry:  newRegistryCache(db.DB),
	}
}

// Add is the function to add any data to the chache
func add(db *bolt.DB, bucket string, name string, item json.RawMessage, ttl time.Duration) error {
	if name == "" {
		return ErrorNameRequired
	}
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		b, err := json.Marshal(cacheItem{
			ValidUntil: time.Now().Add(ttl).Unix(),
			Item:       item,
		})
		logrus.WithFields(logrus.Fields{
			"tool":  name,
			"until": time.Now().Add(ttl).Unix(),
			"item":  string(item)}).Infoln("Cached remote versions")

		return bucket.Put([]byte(name), b)
	})
}

// Clear will clear the given cache entry in the given bucket
func clear(db *bolt.DB, bucketName string, entry string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		return bucket.Delete([]byte(entry))
	})
}

// Resolve will return the string of the given name from the cache if possible.
// If the strings are not available, the callback will be called and added to the cache
// The fallback takes no arguments and requires the strings and a duration as return values.
func resolve(db *bolt.DB, bucket string, name string, fallback func(oldValue json.RawMessage) (json.RawMessage, time.Duration, error)) (json.RawMessage, error) {
	cachedItem := &cacheItem{}
	// logrus.WithField("tool", to.Data().Name).Infoln("Checking db for remote versions")
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		jsonCache := bucket.Get([]byte(name))
		if jsonCache != nil {
			err = json.Unmarshal(jsonCache, &cachedItem)
			if err != nil {
				return err
			}
			logrus.WithFields(logrus.Fields{
				"name":  name,
				"until": cachedItem.ValidUntil,
				"item":  string(cachedItem.Item)}).Infoln("Found cached item")
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if cachedItem == nil || time.Unix(cachedItem.ValidUntil, 0).Before(time.Now()) {
		var item json.RawMessage
		if cachedItem != nil && time.Unix(cachedItem.ValidUntil, 0).Before(time.Now()) {
			logrus.WithField("name", name).Infoln("Cached versions expired")
			item = cachedItem.Item
		}
		logrus.WithField("name", name).Infoln("Remote versions not cached yet")
		remoteVersions, ttl, err := fallback(item)
		if err != nil {
			return nil, err
		}
		return remoteVersions, add(db, bucket, name, remoteVersions, ttl)

	}
	return cachedItem.Item, err
}
