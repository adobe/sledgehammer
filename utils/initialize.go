/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package utils

import (
	"github.com/adobe/sledgehammer/slh/config"
	"github.com/coreos/bbolt"
	"github.com/sirupsen/logrus"
)

var (
	// InitializeFlag will indicate if this is the first run of Sledgehammer
	InitializeFlag = "initialized"
)

// ShouldInitialize will return if this is the first run of the tool
func ShouldInitialize(cfg *config.Config) (bool, error) {
	logrus.Info("Checking if initialize is needed")
	if cfg.Initialized {
		return false, nil
	}
	db, err := cfg.OpenDatabase()
	if err != nil {
		return false, err
	}
	defer cfg.CloseDatabase()

	doInit := false
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(InitializeFlag))
		if err != nil {
			return err
		}
		if bucket != nil {
			initialized := bucket.Get([]byte(InitializeFlag))
			if initialized == nil {
				logrus.Info("Initialized flag is not set")
				doInit = true
				return bucket.Put([]byte(InitializeFlag), []byte(InitializeFlag))
			}
		}
		//  fetch initialize flag, if not there yet set it and call functions
		return nil
	})

	return doInit, err
}
