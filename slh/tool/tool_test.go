/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package tool_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/slh/tool"
	"github.com/adobe/sledgehammer/utils/test"
)

func TestList(t *testing.T) {
	cases := []struct {
		name     string
		expected []string
		previous []tool.Tool
	}{
		{
			name:     "Empty list",
			expected: []string{},
		},
		{
			name: "Single entry list - foo tool",
			previous: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
			},
			expected: []string{"foo/foo"},
		},
		{
			name: "Multiple entry list - foo + bar tool",
			previous: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "bar",
						Registry: "bar",
					},
				},
			},
			expected: []string{"bar/bar", "foo/foo"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			tools := tool.New(config.Database{DB: db})

			if len(tt.previous) > 0 {
				for _, m := range tt.previous {
					err := tools.Add(m)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			sorgedTools, toolMap, err := tools.List()
			if err != nil {
				t.Fatal(err)
			}

			v := make([]string, 0, len(toolMap))

			for _, value := range sorgedTools {
				for _, value1 := range toolMap[value] {
					v = append(v, value1.Data().Registry+"/"+value1.Data().Name)
				}
			}

			assert.EqualValues(t, tt.expected, v)
		})
	}
}

func TestFrom(t *testing.T) {

	cases := []struct {
		name     string
		expected []string
		previous []tool.Tool
		registry string
	}{
		{
			name:     "Empty list",
			expected: []string{},
		},
		{
			name: "Single entry list - no registry",
			previous: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
			},
			expected: []string{},
		},
		{
			name: "Single entry list - foo registry",
			previous: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
			},
			registry: "foo",
			expected: []string{"foo/foo"},
		},
		{
			name: "Single entry list - bar registry",
			previous: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
			},
			registry: "bar",
			expected: []string{},
		},
		{
			name: "Multiple entry list - no registry",
			previous: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "bar",
						Registry: "bar",
					},
				},
			},
			expected: []string{},
		},
		{
			name: "Multiple entry list - foo registry",
			previous: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "bar",
						Registry: "bar",
					},
				},
			},
			registry: "foo",
			expected: []string{"foo/foo"},
		},
		{
			name: "Multiple entry list - bar registry",
			previous: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "bar",
						Registry: "bar",
					},
				},
			},
			registry: "bar",
			expected: []string{"bar/bar"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			tools := tool.New(config.Database{DB: db})

			if len(tt.previous) > 0 {
				for _, m := range tt.previous {
					err := tools.Add(m)
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			outTools, err := tools.From(tt.registry)
			stringTools := []string{}
			if err != nil {
				t.Fatal(err)
			}

			for _, t := range outTools {
				stringTools = append(stringTools, t.Data().Registry+"/"+t.Data().Name)
			}

			assert.EqualValues(t, tt.expected, stringTools)
		})
	}
}

type remove struct {
	registry string
	name     string
}

func TestRemove(t *testing.T) {

	cases := []struct {
		name     string
		expected []string
		previous []tool.Tool
		removing remove
		err      error
	}{
		{
			name: "Empty Registry - wrong arguments #1",
			err:  tool.ErrorRegistryEmpty,
			removing: remove{
				name: "foo",
			},
		},
		{
			name: "Empty Registry - wrong arguments #2",
			err:  tool.ErrorNameEmpty,
			removing: remove{
				registry: "foo",
			},
		},
		{
			name: "Empty Registry - valid tool",
			removing: remove{
				registry: "foo",
				name:     "foo",
			},
			expected: []string{},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			tools := tool.New(config.Database{DB: db})

			if len(tt.previous) > 0 {
				for _, m := range tt.previous {
					err := tools.Add(m)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			err := tools.Remove(tt.removing.registry, tt.removing.name)
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

			sortedTools, toolMap, err := tools.List()
			v := make([]string, 0, len(toolMap))
			if err != nil {
				t.Fatal(err)
			}

			for _, value := range sortedTools {
				for _, value1 := range toolMap[value] {
					v = append(v, value1.Data().Registry+"/"+value1.Data().Name)
				}
			}

			assert.EqualValues(t, tt.expected, v)
		})
	}
}

func TestAdd(t *testing.T) {
	cases := []struct {
		name     string
		expected []string
		previous []tool.Tool
		adding   []tool.Tool
		err      error
	}{
		{
			name:     "Simple Add - no entry",
			expected: []string{"foo/foo"},
			previous: []tool.Tool{},
			adding: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
			},
		},
		{
			name:     "Simple Add - one entry",
			expected: []string{"bar/bar", "foo/foo"},
			previous: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
			},
			adding: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "bar",
						Registry: "bar",
					},
				},
			},
		},
		{
			name:     "Duplicate Add - one entry",
			expected: []string{"foo/foo"},
			previous: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
			},
			adding: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
			},
		},
		{
			name:     "Multiple add - no entry",
			expected: []string{"bar/bar", "foo/foo"},
			previous: []tool.Tool{},
			adding: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "bar",
						Registry: "bar",
					},
				},
			},
		},
		{
			name:     "Multiple add - one duplicate entry",
			expected: []string{"bar/bar", "foo/foo"},
			previous: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
			},
			adding: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "bar",
						Registry: "bar",
					},
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			tools := tool.New(config.Database{DB: db})

			if len(tt.previous) > 0 {
				for _, m := range tt.previous {
					err := tools.Add(m)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			err := tools.Add(tt.adding...)
			if err != nil {
				t.Fatalf(err.Error())
			}
			sortedTools, toolMap, err := tools.List()
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

			sortedTools, toolMap, err = tools.List()
			v := make([]string, 0, len(toolMap))
			if err != nil {
				t.Fatal(err)
			}

			for _, value := range sortedTools {
				for _, value1 := range toolMap[value] {
					v = append(v, value1.Data().Registry+"/"+value1.Data().Name)
				}
			}

			assert.EqualValues(t, tt.expected, v)
		})
	}
}

func TestGet(t *testing.T) {
	cases := []struct {
		name     string
		expected []string
		previous []tool.Tool
		get      remove
		err      error
	}{
		{
			name:     "Simple get - no entry",
			expected: []string{},
			previous: []tool.Tool{},
			get: remove{
				name:     "foo",
				registry: "foo",
			},
			err: tool.ErrorToolNotFound,
		},
		{
			name:     "Simple get - one entry",
			expected: []string{"foo/foo"},
			previous: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
			},
			get: remove{
				name:     "foo",
				registry: "foo",
			},
		},
		{
			name:     "Simple get - multiple entry",
			expected: []string{"foo/foo"},
			previous: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "bar",
						Registry: "bar",
					},
				},
			},
			get: remove{
				name:     "foo",
				registry: "foo",
			},
		},
		{
			name:     "Default get - single entry",
			expected: []string{"foo/foo"},
			previous: []tool.Tool{
				&tool.LocalTool{
					Core: tool.Data{
						Name:     "foo",
						Registry: "foo",
					},
				},
			},
			get: remove{
				name: "foo",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			tools := tool.New(config.Database{DB: db})

			if len(tt.previous) > 0 {
				for _, m := range tt.previous {
					err := tools.Add(m)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			getTool, err := tools.Get(tt.get.registry, tt.get.name)
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

			v := make([]string, 0, 1)

			if getTool != nil {
				v = append(v, getTool.Data().Registry+"/"+getTool.Data().Name)
			}

			assert.EqualValues(t, tt.expected, v)
		})
	}
}

func TestDefaultHandling(t *testing.T) {
	cases := []struct {
		name    string
		execute func(tools *tool.Tools)
	}{
		{
			name: "Single tool as default",
			execute: func(tools *tool.Tools) {
				tools.Add(&tool.LocalTool{
					Core: tool.Data{
						Registry: "foo",
						Name:     "foo",
					},
				})
				to, _ := tools.Get("foo", "foo")
				assert.Equal(t, true, to.Data().Default)
			},
		},
		{
			name: "Single tool with existing tool",
			execute: func(tools *tool.Tools) {
				tools.Add(&tool.LocalTool{
					Core: tool.Data{
						Registry: "foo",
						Name:     "foo",
					},
				})
				tools.Add(&tool.LocalTool{
					Core: tool.Data{
						Registry: "bar",
						Name:     "foo",
					},
				})
				to, _ := tools.Get("foo", "foo")
				assert.Equal(t, true, to.Data().Default)
				to, _ = tools.Get("bar", "foo")
				assert.Equal(t, false, to.Data().Default)
			},
		},
		{
			name: "Single tool after removing",
			execute: func(tools *tool.Tools) {
				tools.Add(&tool.LocalTool{
					Core: tool.Data{
						Registry: "foo",
						Name:     "foo",
					},
				})
				tools.Add(&tool.LocalTool{
					Core: tool.Data{
						Registry: "bar",
						Name:     "foo",
					},
				})
				to, _ := tools.Get("foo", "foo")
				assert.Equal(t, true, to.Data().Default)
				to, _ = tools.Get("bar", "foo")
				assert.Equal(t, false, to.Data().Default)
				tools.Remove("foo", "foo")
				to, _ = tools.Get("bar", "foo")
				assert.Equal(t, true, to.Data().Default)
				l, _, _ := tools.List()
				assert.Len(t, l, 1)
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			defer test.Close(db, t)

			tls := tool.New(config.Database{DB: db})
			tt.execute(tls)
		})
	}
}
