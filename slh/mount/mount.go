/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package mount

import (
	"encoding/json"
	"strings"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/coreos/bbolt"
	"github.com/sirupsen/logrus"
)

var (
	// BucketKey is the name of the bucket where mounts are stored
	BucketKey = "mounts"
	// EntryKey is the name of the entry where the list of mounts are stored
	EntryKey = "mounts"
)

// Mounts is a struct that can be used to access the mounts on this system
type Mounts struct {
	config.Database
}

// New will create a new Mounts struct based on the given bolt database.
// It offers methods to add/remove and list all mounts registered with Sledgehammer
func New(db config.Database) Mounts {
	initDB(db.DB)
	return Mounts{
		Database: db,
	}
}

// initDB will Create the mounts bucket in the database if it does not exist.
func initDB(db *bolt.DB) {
	db.Update(func(tx *bolt.Tx) error {
		logrus.WithField("bucket", BucketKey).Debug("Trying to create the mount bucket")
		tx.CreateBucketIfNotExists([]byte(BucketKey))
		return nil
	})
}

// List will return the current mounts.
func (m *Mounts) List() ([]string, error) {
	logrus.Debug("Listing all mounts")
	mounts := []string{}

	err := m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketKey))
		if bucket != nil {
			byteMounts := bucket.Get([]byte(EntryKey))
			if byteMounts != nil {
				err := json.Unmarshal(byteMounts, &mounts)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	logrus.WithField("mounts", mounts).Debug("Found mounts")
	return mounts, err
}

// Add will add the given path as a mount to Sledgehammer.
// Tools will then be able to access files under this path while they are running
func (m *Mounts) Add(mount ...string) error {
	logrus.WithField("mount", mount).Info("Adding new mount")
	mounts := []string{}
	err := m.DB.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists([]byte(BucketKey))
		if err != nil {
			return err
		}
		if bucket != nil {
			byteMounts := bucket.Get([]byte(EntryKey))
			if byteMounts != nil {
				err := json.Unmarshal(byteMounts, &mounts)
				if err != nil {
					return err
				}
			}
			for _, mo := range mount {
				if !hasMount(mounts, mo) {
					mounts = append(mounts, mo)

					//  now remove all mounts that are included in the new mount
					for i := len(mounts) - 1; i >= 0; i-- {
						if mounts[i] != mo && strings.HasPrefix(mounts[i]+"/", mo+"/") {
							// do it in the background, otherwise we block the transaction
							logrus.WithField("mount", m).WithField("sub", m).Debug("Removing sub since it is already included in mount")
							mounts = append(mounts[:i], mounts[i+1:]...)
						}
					}

					b, err := json.Marshal(mounts)
					if err != nil {
						return err
					}
					err = bucket.Put([]byte(EntryKey), b)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	return err
}

// Remove will remove the given mounts from Sledgehammer if possible.
// It requires an exact match for the mount
func (m *Mounts) Remove(mount string) error {
	logrus.WithField("mount", mount).Debug("Removing mount")
	mounts := []string{}
	return m.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketKey))
		if bucket != nil {
			byteMounts := bucket.Get([]byte(EntryKey))
			if byteMounts != nil {
				err := json.Unmarshal(byteMounts, &mounts)
				if err != nil {
					return err
				}
			}
			if hasMount(mounts, mount) {
				for i := len(mounts) - 1; i >= 0; i-- {
					if mounts[i] == mount {
						mounts = append(mounts[:i], mounts[i+1:]...)
					}
				}
				b, err := json.Marshal(mounts)
				if err != nil {
					return err
				}
				logrus.WithField("mount", mount).Debug("Removed mount")
				bucket.Put([]byte(EntryKey), b)
			}
		}
		return nil
	})
}

func hasMount(mounts []string, mount string) bool {
	for _, m := range mounts {
		if m == mount || strings.HasPrefix(mount+"/", m+"/") {
			logrus.WithField("toFind", mount).WithField("mounts", mounts).Debug("Found mount in mounts")
			return true
		}
	}
	logrus.WithField("toFind", mount).WithField("mounts", mounts).Debug("Mount not found in mounts")
	return false
}
