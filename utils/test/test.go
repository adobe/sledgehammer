/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsouza/go-dockerclient"

	"github.com/adobe/sledgehammer/slh/cmd"
	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/utils/contracts"
	"github.com/coreos/bbolt"
	"github.com/stretchr/testify/assert"
)

// Step is a single step for integration testing the executable. It consists of a command and the output
type Step struct {
	Cmd      string
	Has      []string
	Not      []string
	DoBefore func(cfg *config.Config)
	DoAfter  func(cfg *config.Config)
}

// NewTestDB returns a TestDB using a temporary path.
func NewTestDB(t *testing.T) *bolt.DB {
	// Retrieve a temporary path.
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	os.Remove(path)
	// Open the database.
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Return wrapped type.
	return db
}

// Close and delete Bolt database.
func Close(db *bolt.DB, t *testing.T) {
	path := db.Path()
	err := db.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Remove(path)
	if err != nil {
		t.Fatal(err)
	}
}

// NewTmpDir creates a new temp dir for the tests
func NewTmpDir(t *testing.T) string {
	path, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	return path

}

// DeleteTmpDir will delete the created tmp directory
func DeleteTmpDir(path string, t *testing.T) {
	err := os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

type TestCase struct {
	Name             string
	Steps            []*Step
	ShouldInitialize bool
}

func DoTest(t *testing.T, cases []*TestCase) {
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			path := NewTmpDir(t)
			defer DeleteTmpDir(path, t)

			for _, st := range tt.Steps {
				stdOut := &bytes.Buffer{}
				pr, pw := io.Pipe()
				defer pr.Close()
				defer pw.Close()

				stdErr := &bytes.Buffer{}

				docker, err := docker.NewClientFromEnv()
				if err != nil {
					t.Fatal(err)
				}

				cfg := &config.Config{
					IO: &config.IO{
						Out: stdOut,
						In:  pr,
						Err: stdErr,
					},
					Docker: config.Docker{
						Docker: docker,
					},
					ConfigDir:   path,
					Initialized: !tt.ShouldInitialize,
				}
				if st.DoBefore != nil {
					st.DoBefore(cfg)
				}
				cmd := cmd.CreateRootCommand(cfg)
				cmd.SetArgs(strings.Split(st.Cmd, " "))
				cmd.Execute()
				if st.DoAfter != nil {
					st.DoAfter(cfg)
				}
				if len(st.Has) > 0 {
					for _, has := range st.Has {
						assert.Contains(t, stdOut.String(), has)
					}
				}
				if len(st.Not) > 0 {
					for _, not := range st.Not {
						assert.NotContains(t, stdOut.String(), not)
					}
				}
			}
		})
	}
}

// PrepareLocalRegistries will write some real registries to the given path
func PrepareLocalRegistries(path string) {
	reg1 := contracts.Registry{
		Description: "Foo description",
		Maintainer:  "plaschke@adobe.com",
		Tools: []contracts.Tool{
			{
				Description: "This is a foo tool",
				Image:       "foo-tools/foo",
				Name:        "foo",
				Type:        "local",
			},
		},
		Kits: []contracts.Kit{
			{
				Name:        "foo-kit",
				Description: "foo description",
				Tools: []contracts.KitTool{
					{
						Name:    "foo",
						Version: "*",
						Alias:   "foo",
					},
				},
			},
		},
	}
	reg2 := contracts.Registry{
		Description: "Bar description",
		Maintainer:  "plaschke@adobe.com",
		Tools: []contracts.Tool{
			{
				Description: "This is a bar tool",
				Image:       "bar-tools/bar",
				Name:        "bar",
				Type:        "local",
			},
		},
	}
	reg3 := contracts.Registry{
		Description: "Baz description",
		Maintainer:  "plaschke@adobe.com",
		Tools: []contracts.Tool{
			{
				Description: "This is a baz tool",
				Image:       "baz-tools/baz",
				Name:        "baz",
				Type:        "local",
			},
			{
				Description: "This is a foo tool",
				Image:       "baz-tools/foo",
				Name:        "foo",
				Type:        "local",
			},
		},
	}

	real := contracts.Registry{
		Description: "Real description",
		Maintainer:  "plaschke@adobe.com",
		Tools: []contracts.Tool{
			{
				Description: "This is a real tool",
				Image:       "alpine",
				Name:        "real",
				Type:        "hub",
				Entry:       []string{"sh", "-c", "echo Hello from the real tool"},
			},
			{
				Description: "This is a real alpine tool",
				Image:       "alpine",
				Name:        "alpine-test-version",
				Type:        "hub",
				Entry:       []string{"sh", "-c", "cat /etc/alpine-release"},
			},
			{
				Name:        "real-daemon",
				Description: "This is a real daemon tool",
				Image:       "alpine",
				Entry:       []string{"sh", "-c", "echo Hello from the daemon tool"},
				Daemon: &contracts.ToolDaemon{
					Entry: []string{"/bin/ash"},
				},
				Type: "hub",
			},
			{
				Name:        "argument-test",
				Description: "This is a real daemon tool",
				Image:       "alpine",
				Entry:       []string{"sh", "-c", "echo \"Foo: ${0}\""},
				Daemon: &contracts.ToolDaemon{
					Entry: []string{"/bin/ash"},
				},
				Type: "hub",
			},
		},
	}

	b1, _ := json.Marshal(reg1)
	b2, _ := json.Marshal(reg2)
	b3, _ := json.Marshal(reg3)
	bReal, _ := json.Marshal(real)

	ioutil.WriteFile(filepath.Join(path, "foo.json"), b1, 0777)
	ioutil.WriteFile(filepath.Join(path, "bar.json"), b2, 0777)
	ioutil.WriteFile(filepath.Join(path, "baz.json"), b3, 0777)
	ioutil.WriteFile(filepath.Join(path, "real.json"), bReal, 0777)
}

func PrepareChangedRegistries(path string) {
	reg1 := contracts.Registry{
		Description: "Foo description",
		Maintainer:  "plaschke@adobe.com",
		Tools: []contracts.Tool{
			{
				Description: "This is a foo tool",
				Image:       "foo-tools/foo",
				Name:        "foo",
				Type:        "local",
			},
		},
	}
	reg2 := contracts.Registry{
		Description: "Bar description",
		Maintainer:  "plaschke@adobe.com",
		Tools: []contracts.Tool{
			{
				Description: "This is a bar tool",
				Image:       "bar-tools/bar",
				Name:        "bar",
				Type:        "local",
			},
		},
	}
	reg3 := contracts.Registry{
		Description: "Baz description",
		Maintainer:  "plaschke@adobe.com",
		Tools: []contracts.Tool{
			{
				Description: "This is a baz tool",
				Image:       "baz-tools/baz",
				Name:        "baz",
				Type:        "local",
			},
			{
				Description: "This is a foo tool",
				Image:       "baz-tools/foo",
				Name:        "foo",
				Type:        "local",
			},
		},
	}

	real := contracts.Registry{
		Description: "Real description",
		Maintainer:  "plaschke@adobe.com",
		Tools: []contracts.Tool{
			{
				Description: "This is a real tool",
				Image:       "alpine",
				Name:        "real",
				Type:        "hub",
				Entry:       []string{"sh", "-c", "echo Hello from the real tool, now updated"},
			},
			{
				Description: "This is a real alpine tool",
				Image:       "alpine",
				Name:        "alpine-version",
				Type:        "hub",
				Entry:       []string{"sh", "-c", " cat /etc/alpine-release"},
			},
		},
	}

	b1, _ := json.Marshal(reg1)
	b2, _ := json.Marshal(reg2)
	b3, _ := json.Marshal(reg3)
	bReal, _ := json.Marshal(real)

	ioutil.WriteFile(filepath.Join(path, "foo.json"), b1, 0777)
	ioutil.WriteFile(filepath.Join(path, "bar.json"), b2, 0777)
	ioutil.WriteFile(filepath.Join(path, "baz.json"), b3, 0777)
	ioutil.WriteFile(filepath.Join(path, "real.json"), bReal, 0777)
}
