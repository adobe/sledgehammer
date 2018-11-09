/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package db

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/coreos/bbolt"

	"github.com/sirupsen/logrus"
)

// Open will try to open the database in the given path and error out if any problem occurs.
func Open(configDir string) (*bolt.DB, error) {
	path, err := filepath.Abs(filepath.Join(configDir, "data.db"))
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Creating path at %s", configDir)
	err = os.MkdirAll(configDir, 0766)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Opening database at %s", path)
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, errors.New("Error while opening database: " + err.Error())
	}
	return db, err
}
