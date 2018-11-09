/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package cache_test

import (
	"testing"
	"time"

	"github.com/fsouza/go-dockerclient"

	"github.com/adobe/sledgehammer/slh/tool"

	"github.com/stretchr/testify/assert"

	"github.com/adobe/sledgehammer/mocks"
	"github.com/adobe/sledgehammer/slh/cache"
	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/utils/test"
	"github.com/golang/mock/gomock"
)

func TestContainer(t *testing.T) {
	cases := []struct {
		name     string
		before   func(*cache.Cache, *mocks.MockClient, *mocks.MockTool)
		expected string
		err      error
	}{
		{
			name: "No daemon tool",
			before: func(c *cache.Cache, m *mocks.MockClient, t *mocks.MockTool) {
				t.EXPECT().Data().Return(&tool.Data{
					Registry: "foo",
					Name:     "bar",
				}).AnyTimes()
			},
			expected: "",
		},
		{
			name: "No container ID in cache yet",
			before: func(c *cache.Cache, m *mocks.MockClient, t *mocks.MockTool) {
				m.EXPECT().CreateContainer(gomock.Any()).Return(&docker.Container{ID: "foobar"}, nil)
				m.EXPECT().StartContainer(gomock.Any(), gomock.Any())
				// t.EXPECT().Versions().Return([]string{"1"}, nil)
				t.EXPECT().Data().Return(&tool.Data{
					Registry: "foo",
					Name:     "bar",
					Daemon: &tool.Daemon{
						Entry: []string{"foo"},
					},
				}).AnyTimes()
			},
			expected: "foobar",
		},
		{
			name: "Container ID in cache",
			before: func(c *cache.Cache, m *mocks.MockClient, t *mocks.MockTool) {
				m.EXPECT().CreateContainer(gomock.Any()).Return(&docker.Container{ID: "foobar"}, nil)
				m.EXPECT().StartContainer(gomock.Any(), gomock.Any())
				// t.EXPECT().Versions().Return([]string{"1"}, nil)
				t.EXPECT().Data().Return(&tool.Data{
					Registry: "foo",
					Name:     "bar",
					Daemon: &tool.Daemon{
						Entry: []string{"foo"},
					},
				}).AnyTimes()
				c.Container.Get(t, "1", &config.Config{
					Docker: config.Docker{
						Docker: m,
					},
				}, []string{})
			},
			expected: "foobar",
		},
		{
			name: "Container ID in cache expired",
			before: func(c *cache.Cache, m *mocks.MockClient, t *mocks.MockTool) {
				cache.ContainerIDTTL = 0 * time.Microsecond
				m.EXPECT().CreateContainer(gomock.Any()).Return(&docker.Container{ID: "foobar"}, nil)
				m.EXPECT().StartContainer(gomock.Any(), gomock.Any())
				// t.EXPECT().Versions().Return([]string{"1"}, nil)
				t.EXPECT().Data().Return(&tool.Data{
					Registry: "foo",
					Name:     "bar",
					Daemon: &tool.Daemon{
						Entry: []string{"foo"},
					},
				}).AnyTimes()
				c.Container.Get(t, "1", &config.Config{
					Docker: config.Docker{
						Docker: m,
					},
				}, []string{})
				m.EXPECT().RemoveContainer(gomock.Any())
				m.EXPECT().CreateContainer(gomock.Any()).Return(&docker.Container{ID: "foobar2"}, nil)
				m.EXPECT().StartContainer(gomock.Any(), gomock.Any())
			},
			expected: "foobar2",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// create db
			db := test.NewTestDB(t)
			defer test.Close(db, t)
			c := cache.New(config.Database{DB: db})

			// create mock docker
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			dockerMock := mocks.NewMockClient(mockCtrl)
			toolMock := mocks.NewMockTool(mockCtrl)

			// execute before
			if tt.before != nil {
				tt.before(c, dockerMock, toolMock)
			}

			// do GET
			id, err := c.Container.Get(
				toolMock,
				"1",
				&config.Config{
					Docker: config.Docker{
						Docker: dockerMock,
					},
				},
				[]string{},
			)

			// execute after
			if err != nil {
				assert.Equal(t, tt.err, err)
				return
			}
			assert.Equal(t, tt.expected, id)
		})
	}
}

// test clear

// test registry

// test version
