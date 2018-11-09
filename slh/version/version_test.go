/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package version_test

import (
	"testing"

	"github.com/adobe/sledgehammer/slh/version"
	"github.com/stretchr/testify/assert"
)

func TestShouldPull(t *testing.T) {
	cases := []struct {
		local    string
		remote   string
		expected bool
	}{
		{
			local:    "",
			remote:   "",
			expected: false,
		},
		{
			local:    "1.2.0",
			remote:   "1.2.0",
			expected: false,
		},
		{
			local:    "1.2.3",
			remote:   "1.2.0",
			expected: false,
		},
		{
			local:    "1.2.0",
			remote:   "1.2.3",
			expected: true,
		},
	}
	for _, tt := range cases {
		t.Run("ShouldPull", func(t *testing.T) {
			should := version.ShouldPull(tt.local, tt.remote)
			assert.Equal(t, tt.expected, should)
		})
	}
}
func TestHas(t *testing.T) {
	cases := []struct {
		version           string
		included_versions []string
		expected          bool
	}{
		{
			version:           "",
			included_versions: []string{},
			expected:          false,
		},
		{
			version:           "1.2.3",
			included_versions: []string{"1.2.0", "1.2.4"},
			expected:          false,
		},
		{
			version:           "1.2.3",
			included_versions: []string{"1.2.0", "1.2.3", "1.2.4"},
			expected:          true,
		},
	}
	for _, tt := range cases {
		t.Run("Has", func(t *testing.T) {
			should := version.Has(tt.included_versions, tt.version)
			assert.Equal(t, tt.expected, should)
		})
	}
}

func TestMerge(t *testing.T) {
	cases := []struct {
		listA    []string
		listB    []string
		expected []string
	}{
		{
			listA:    []string{},
			listB:    []string{},
			expected: []string{},
		},
		{
			listA:    []string{"foo"},
			listB:    []string{},
			expected: []string{"foo"},
		},
		{
			listA:    []string{},
			listB:    []string{"bar"},
			expected: []string{"bar"},
		},
		{
			listA:    []string{"foo"},
			listB:    []string{"bar"},
			expected: []string{"foo", "bar"},
		},
		{
			listA:    []string{"foo"},
			listB:    []string{"foo"},
			expected: []string{"foo"},
		},
	}
	for _, tt := range cases {
		t.Run("Merge", func(t *testing.T) {
			should := version.Merge(tt.listA, tt.listB)
			assert.ElementsMatch(t, tt.expected, should)
		})
	}
}
func TestSelectVersions(t *testing.T) {
	cases := []struct {
		versions []string
		toSearch string
		expected string
	}{
		{
			versions: []string{},
			toSearch: "",
			expected: "",
		},
		{
			versions: []string{"1.2.3", "1.2", "1.2.4", "1", "2"},
			toSearch: "1",
			expected: "1.2.4",
		},
		{
			versions: []string{"1.2.3", "1.2", "1.3.4", "1", "2"},
			toSearch: "1.2.X",
			expected: "1.2.3",
		},
		{
			versions: []string{"1.2.3", "1.2", "1.3.4", "1", "2"},
			toSearch: "1.x",
			expected: "1.3.4",
		},
		{
			versions: []string{"1.2.3", "1.2", "1.3.4", "1", "2"},
			toSearch: "^1",
			expected: "1.3.4",
		},
		{
			versions: []string{"1.2.3", "1.2", "1.3.4", "1", "2", "latest"},
			toSearch: "",
			expected: "2",
		},
		{
			versions: []string{"latest"},
			toSearch: "",
			expected: "latest",
		},
		{
			versions: []string{"1.2.3", "1.2", "1.3.4", "1", "2", "foobar"},
			toSearch: "foobar",
			expected: "foobar",
		},
		{
			versions: []string{"1.2.3", "1.2", "1.3.4", "1", "2", "foobar"},
			toSearch: "baz",
			expected: "",
		},
		{
			versions: []string{"0.1.0-1", "latest", "0.1.0-2"},
			toSearch: "*",
			expected: "0.1.0-2",
		},
		{
			versions: []string{"0.1.0-1", "latest", "0.1.0-2"},
			toSearch: "",
			expected: "0.1.0-2",
		},
		{
			versions: []string{"1.1.0-1", "latest", "1.1.0-2"},
			toSearch: "*",
			expected: "1.1.0-2",
		},
		{
			versions: []string{"1.1.0-1", "latest", "1.1.0-2"},
			toSearch: "",
			expected: "1.1.0-2",
		},
		{
			versions: []string{"1.1.0", "1.1.0-1", "latest", "1.1.0-2"},
			toSearch: "*",
			expected: "1.1.0",
		},
		{
			versions: []string{"1.1.0", "1.1.0-1", "latest", "1.1.0-2"},
			toSearch: "",
			expected: "1.1.0",
		},
	}
	for _, tt := range cases {
		t.Run("versions", func(t *testing.T) {
			should := version.Select(tt.versions, tt.toSearch)
			assert.Equal(t, tt.expected, should)
		})
	}

}
