/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package alias_test

import (
	"testing"

	"github.com/adobe/sledgehammer/slh/alias"
	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/utils/test"
	"github.com/stretchr/testify/assert"
)

func TestAliasList(t *testing.T) {
	cases := []struct {
		name     string
		expected []alias.Alias
		previous []alias.Alias
	}{
		{
			name:     "Empty list",
			expected: []alias.Alias{},
			previous: []alias.Alias{},
		},
		{
			name: "One entry",
			expected: []alias.Alias{
				{
					Name: "foo",
				},
			},
			previous: []alias.Alias{
				{
					Name: "foo",
				},
			},
		},
		{
			name: "Multiple Entry",
			expected: []alias.Alias{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
			},
			previous: []alias.Alias{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			aliases := alias.New(config.Database{DB: db})

			if len(tt.previous) > 0 {
				for _, m := range tt.previous {
					aliases.Add(m)
				}
			}

			ms, err := aliases.List()
			if err != nil {
				t.Fatal(err)
			}
			assert.ElementsMatch(t, tt.expected, ms)
		})
	}
}

func TestAliasAdd(t *testing.T) {
	cases := []struct {
		name     string
		expected []alias.Alias
		toAdd    alias.Alias
		previous []alias.Alias
		err      error
	}{
		{
			name:     "Add nothing",
			expected: []alias.Alias{},
			toAdd:    alias.Alias{},
			previous: []alias.Alias{},
			err:      alias.ErrorNoName,
		},
		{
			name: "Add single",
			expected: []alias.Alias{
				{
					Name: "foo",
				},
			},
			toAdd: alias.Alias{
				Name: "foo",
			},
			previous: []alias.Alias{},
		},
		{
			name: "Add already existing",
			expected: []alias.Alias{
				{
					Name: "foo",
				},
			},
			toAdd: alias.Alias{
				Name: "foo",
			},
			previous: []alias.Alias{
				{
					Name: "foo",
				},
			},
		},
		{
			name: "Add another alias",
			expected: []alias.Alias{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
			},
			toAdd: alias.Alias{
				Name: "bar",
			},
			previous: []alias.Alias{
				{
					Name: "foo",
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			aliases := alias.New(config.Database{DB: db})

			if len(tt.previous) > 0 {
				for _, m := range tt.previous {
					aliases.Add(m)
				}
			}

			err := aliases.Add(tt.toAdd)
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
			alis, err := aliases.List()
			if err != nil {
				t.Fatal(err)
			}
			assert.ElementsMatch(t, tt.expected, alis)
		})
	}
}

func TestAliasGet(t *testing.T) {
	cases := []struct {
		name     string
		expected *alias.Alias
		toFind   string
		previous []alias.Alias
		err      error
	}{
		{
			name:     "Get nothing",
			expected: &alias.Alias{},
			toFind:   "",
			previous: []alias.Alias{},
			err:      alias.ErrorNotFound,
		},
		{
			name:     "Get nonexistend",
			expected: &alias.Alias{},
			toFind:   "foo",
			previous: []alias.Alias{},
			err:      alias.ErrorNotFound,
		},
		{
			name:     "Get nonexistend with entries",
			expected: &alias.Alias{},
			toFind:   "foo",
			previous: []alias.Alias{
				{
					Name: "bar",
				},
			},
			err: alias.ErrorNotFound,
		},
		{
			name: "Get one with entry",
			expected: &alias.Alias{
				Name: "foo",
			},
			toFind: "foo",
			previous: []alias.Alias{
				{
					Name: "foo",
				},
			},
		},
		{
			name: "Get one with entries",
			expected: &alias.Alias{
				Name: "foo",
			},
			toFind: "foo",
			previous: []alias.Alias{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			aliases := alias.New(config.Database{DB: db})

			if len(tt.previous) > 0 {
				for _, m := range tt.previous {
					aliases.Add(m)
				}
			}

			ali, err := aliases.Get(tt.toFind)
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
			assert.Equal(t, tt.expected, ali)
		})
	}
}

func TestAliasHas(t *testing.T) {
	cases := []struct {
		name     string
		expected bool
		toFind   string
		previous []alias.Alias
	}{
		{
			name:     "Has nothing",
			expected: false,
			toFind:   "",
			previous: []alias.Alias{},
		},
		{
			name:     "Has nonexistend",
			expected: false,
			toFind:   "foo",
			previous: []alias.Alias{},
		},
		{
			name:     "Has nonexistend with entries",
			expected: false,
			toFind:   "foo",
			previous: []alias.Alias{
				{
					Name: "bar",
				},
			},
		},
		{
			name:     "Has one with entry",
			expected: true,
			toFind:   "foo",
			previous: []alias.Alias{
				{
					Name: "foo",
				},
			},
		},
		{
			name:     "Has one with entries",
			expected: true,
			toFind:   "foo",
			previous: []alias.Alias{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			aliases := alias.New(config.Database{DB: db})

			if len(tt.previous) > 0 {
				for _, m := range tt.previous {
					aliases.Add(m)
				}
			}

			ali, err := aliases.Has(tt.toFind)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.expected, ali)
		})
	}
}

func TestAliasRemove(t *testing.T) {
	cases := []struct {
		name     string
		expected []alias.Alias
		toRemove string
		previous []alias.Alias
	}{
		{
			name:     "Remove nothing",
			expected: []alias.Alias{},
			toRemove: "",
			previous: []alias.Alias{},
		},
		{
			name: "Remove nothing with entry",
			expected: []alias.Alias{
				{
					Name: "foo",
				},
			},
			toRemove: "",
			previous: []alias.Alias{
				{
					Name: "foo",
				},
			},
		},
		{
			name:     "Remove one with one entry",
			expected: []alias.Alias{},
			toRemove: "foo",
			previous: []alias.Alias{
				{
					Name: "foo",
				},
			},
		},
		{
			name: "Remove one with multiple entry",
			expected: []alias.Alias{
				{
					Name: "bar",
				},
			},
			toRemove: "foo",
			previous: []alias.Alias{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			aliases := alias.New(config.Database{DB: db})

			if len(tt.previous) > 0 {
				for _, m := range tt.previous {
					aliases.Add(m)
				}
			}

			err := aliases.Remove(tt.toRemove)
			if err != nil {
				t.Fatal(err)
			}
			alis, err := aliases.List()
			if err != nil {
				t.Fatal(err)
			}
			assert.ElementsMatch(t, tt.expected, alis)
		})
	}
}

// TestFrom

// Test symlinks
