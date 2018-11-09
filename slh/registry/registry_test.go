/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package registry_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/registry"
	"github.com/adobe/sledgehammer/utils/test"
)

func TestList(t *testing.T) {
	pathToAdd := test.NewTmpDir(t)
	defer test.DeleteTmpDir(pathToAdd, t)

	test.PrepareLocalRegistries(pathToAdd)

	cases := []struct {
		name       string
		registries []registry.Registry
		expected   []registry.Registry
	}{
		{
			name:       "List, no entries",
			registries: []registry.Registry{},
			expected:   []registry.Registry{},
		},
		{
			name: "List, one entry",
			registries: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "foo.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "foo",
					},
				},
			},
			expected: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "foo.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "foo",
					},
				},
			},
		},
		{
			name: "List, multiple entries",
			registries: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "foo.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "foo",
					},
				},
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
			},
			expected: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "foo.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "foo",
					},
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			registries := registry.New(config.Database{DB: db})

			if len(tt.registries) > 0 {
				for _, reg := range tt.registries {
					err := registries.Add(reg)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			regs, err := registries.List()
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, tt.expected, regs)
		})
	}
}

func TestExists(t *testing.T) {
	pathToAdd := test.NewTmpDir(t)
	defer test.DeleteTmpDir(pathToAdd, t)

	test.PrepareLocalRegistries(pathToAdd)

	cases := []struct {
		name       string
		registries []registry.Registry
		toFind     string
		found      bool
	}{
		{
			name:       "No entry, nothing to find",
			registries: []registry.Registry{},
			toFind:     "",
			found:      false,
		},
		{
			name:       "No entry, one to find",
			registries: []registry.Registry{},
			toFind:     "foo",
			found:      false,
		},
		{
			name: "One entry, nothing to find",
			registries: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
			},
			toFind: "foo",
			found:  false,
		},
		{
			name: "One entry, one to find",
			registries: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "foo.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "foo",
					},
				},
			},
			toFind: "foo",
			found:  true,
		},
		{
			name: "Multiple entry, one to find",
			registries: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "foo.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "foo",
					},
				},
			},
			toFind: "foo",
			found:  true,
		},
		{
			name: "Multiple entry, nothing to find",
			registries: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "foo.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "foo",
					},
				},
			},
			toFind: "baz",
			found:  false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			registries := registry.New(config.Database{DB: db})

			if len(tt.registries) > 0 {
				for _, reg := range tt.registries {
					err := registries.Add(reg)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			exists, err := registries.Exists(tt.toFind)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.found, exists)
		})
	}
}

func TestRemove(t *testing.T) {

	pathToAdd := test.NewTmpDir(t)
	defer test.DeleteTmpDir(pathToAdd, t)

	test.PrepareLocalRegistries(pathToAdd)

	cases := []struct {
		name       string
		registries []registry.Registry
		toRemove   string
		expected   []registry.Registry
		err        error
	}{
		{
			name:       "Nothing",
			registries: []registry.Registry{},
			toRemove:   "",
			err:        registry.ErrorNoName,
			expected:   []registry.Registry{},
		},
		{
			name: "Nothing to remove",
			registries: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
			},
			toRemove: "",
			err:      registry.ErrorNoName,
			expected: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
			},
		},
		{
			name: "One entry to remove",
			registries: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
			},
			toRemove: "bar",
			expected: []registry.Registry{},
		},
		{
			name: "Two entries to remove",
			registries: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "foo.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "foo",
					},
				},
			},
			toRemove: "bar",
			expected: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "foo.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "foo",
					},
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			registries := registry.New(config.Database{DB: db})

			if len(tt.registries) > 0 {
				for _, reg := range tt.registries {
					err := registries.Add(reg)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			err := registries.Remove(tt.toRemove)
			if err != nil {
				assert.Equal(t, tt.err, err)
			}
			regs, err := registries.List()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.expected, regs)
		})
	}
}

func TestAdd(t *testing.T) {

	pathToAdd := test.NewTmpDir(t)
	defer test.DeleteTmpDir(pathToAdd, t)

	test.PrepareLocalRegistries(pathToAdd)

	cases := []struct {
		name       string
		registries []registry.Registry
		toAdd      registry.Registry
		expected   []registry.Registry
		err        error
	}{
		{
			name:       "Nothing to add",
			registries: []registry.Registry{},
			toAdd:      &registry.LocalRegistry{Core: registry.Data{}},
			expected:   []registry.Registry{},
			err:        registry.ErrorNoName,
		},
		{
			name:       "Nothing to add",
			registries: []registry.Registry{},
			toAdd:      &registry.LocalRegistry{Core: registry.Data{}},
			expected:   []registry.Registry{},
			err:        registry.ErrorNoName,
		},
		{
			name:       "Normal add, no entries",
			registries: []registry.Registry{},
			toAdd: &registry.LocalRegistry{
				Location: filepath.Join(pathToAdd, "foo.json"),
				Core: registry.Data{
					Type: registry.RegTypeLocal,
					Name: "foo",
				},
			},
			expected: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "foo.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "foo",
					},
				},
			},
		},
		{
			name: "Normal add, one entry",
			registries: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
			},
			toAdd: &registry.LocalRegistry{
				Location: filepath.Join(pathToAdd, "foo.json"),
				Core: registry.Data{
					Type: registry.RegTypeLocal,
					Name: "foo",
				},
			},
			expected: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "foo.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "foo",
					},
				},
			},
		},
		{
			name: "Normal add, multiple entry",
			registries: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "baz.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "baz",
					},
				},
			},
			toAdd: &registry.LocalRegistry{
				Location: filepath.Join(pathToAdd, "foo.json"),
				Core: registry.Data{
					Type: registry.RegTypeLocal,
					Name: "foo",
				},
			},
			expected: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "baz.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "baz",
					},
				},
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "foo.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "foo",
					},
				},
			},
		},
		{
			name: "Duplicate add",
			registries: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
			},
			toAdd: &registry.LocalRegistry{
				Location: filepath.Join(pathToAdd, "bar.json"),
				Core: registry.Data{
					Type: registry.RegTypeLocal,
					Name: "bar",
				},
			},
			expected: []registry.Registry{
				&registry.LocalRegistry{
					Location: filepath.Join(pathToAdd, "bar.json"),
					Core: registry.Data{
						Type: registry.RegTypeLocal,
						Name: "bar",
					},
				},
			},
			err: registry.ErrorAlreadyExists,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			registries := registry.New(config.Database{DB: db})

			if len(tt.registries) > 0 {
				for _, reg := range tt.registries {
					err := registries.Add(reg)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			err := registries.Add(tt.toAdd)
			if err != nil {
				if tt.err != nil {
					assert.Equal(t, tt.err, err)
					return
				}
				t.Fatal(err)
			}
			if tt.err != nil {
				t.Fatal("Error should have thrown but didn't:", tt.err)
			}
			regs, err := registries.List()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.expected, regs)
		})
	}
}
