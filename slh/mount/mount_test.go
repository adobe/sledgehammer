/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package mount_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/mount"
	"github.com/adobe/sledgehammer/utils/test"
	homedir "github.com/mitchellh/go-homedir"
)

func TestList(t *testing.T) {

	home, err := homedir.Dir()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name     string
		expected []string
		previous []string
	}{
		{
			name:     "Empty list",
			expected: []string{},
		},
		{
			name:     "Single entry list - home",
			previous: []string{home},
			expected: []string{home},
		},
		{
			name:     "Single entry list - not home",
			previous: []string{"/tmp"},
			expected: []string{"/tmp"},
		},
		{
			name:     "Multi entry list - not home",
			previous: []string{"/foobar", "/tmp"},
			expected: []string{"/foobar", "/tmp"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			mounts := mount.New(config.Database{DB: db})

			if len(tt.previous) > 0 {
				for _, m := range tt.previous {
					mounts.Add(m)
				}
			}

			ms, err := mounts.List()
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, tt.expected, ms)
		})
	}
}

func TestAdd(t *testing.T) {

	home, err := homedir.Dir()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name     string
		expected []string
		previous []string
		adding   []string
	}{
		{
			name:     "Adding none",
			expected: []string{},
		},
		{
			name:     "Add home when it is not there yet",
			expected: []string{home},
			adding:   []string{home},
		},
		{
			name:     "Adding /tmp",
			expected: []string{"/tmp"},
			adding:   []string{"/tmp"},
		},
		{
			name:     "Adding /tmp when there is already a subfolder",
			expected: []string{"/tmp"},
			adding:   []string{"/tmp"},
			previous: []string{"/tmp/foo"},
		},
		{
			name:     "Adding /tmp when there are already subfolders",
			expected: []string{"/tmp"},
			adding:   []string{"/tmp"},
			previous: []string{"/tmp/foo", "/tmp/bar"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			mounts := mount.New(config.Database{DB: db})

			if len(tt.previous) > 0 {
				for _, m := range tt.previous {
					mounts.Add(m)
				}
			}

			err := mounts.Add(tt.adding...)
			if err != nil {
				t.Fatal(err)
			}

			ms, err := mounts.List()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.expected, ms)
		})
	}
}

func TestRemove(t *testing.T) {

	home, err := homedir.Dir()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name     string
		expected []string
		previous []string
		removing []string
	}{
		{
			name:     "Removing none",
			expected: []string{},
		},
		{
			name:     "Remove home when it is not there yet",
			expected: []string{},
			removing: []string{home},
		},
		{
			name:     "Remove tmp when home is there",
			previous: []string{"/tmp", home},
			expected: []string{home},
			removing: []string{"/tmp"},
		},
		{
			name:     "Remove tmp when home is not there",
			previous: []string{"/tmp"},
			expected: []string{},
			removing: []string{"/tmp"},
		},
		{
			name:     "Removing non existing path /foo",
			expected: []string{"/tmp"},
			previous: []string{"/tmp"},
			removing: []string{"/foo"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			mounts := mount.New(config.Database{DB: db})

			if len(tt.previous) > 0 {
				for _, m := range tt.previous {
					mounts.Add(m)
				}
			}

			for _, m := range tt.removing {
				err := mounts.Remove(m)
				if err != nil {
					t.Fatal(err)
				}
			}
			ms, err := mounts.List()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.expected, ms)
		})
	}
}
