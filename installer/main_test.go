/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/adobe/sledgehammer/mocks"
	"github.com/adobe/sledgehammer/utils/test"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
)

var dockerTests = []struct {
	name     string
	expects  func(mock *mocks.MockClient, systemMock *mocks.MockENV)
	contains string
	err      error
}{
	{
		name: "Docker not available",
		expects: func(mock *mocks.MockClient, systemMock *mocks.MockENV) {
			mock.EXPECT().ListImages(gomock.Any()).Times(1).Return(nil, errors.New("TestError"))
		},
		contains: "STEP 1 of 2:",
		err:      ErrorDockerMountMissing,
	},
	{
		name: "Docker still not available",
		expects: func(mock *mocks.MockClient, systemMock *mocks.MockENV) {
			mock.EXPECT().ListImages(gomock.Any()).Times(1).Return([]docker.APIImages{}, nil)
			mock.EXPECT().InspectContainer(gomock.Any()).Times(1).Return(nil, errors.New("TestError"))
		},
		contains: "Could not inspect the running container,",
		err:      ErrorContainerNotFound,
	},
	{
		name: ":/bin not mounted",
		expects: func(mock *mocks.MockClient, systemMock *mocks.MockENV) {
			mock.EXPECT().ListImages(gomock.Any()).Times(1).Return([]docker.APIImages{}, nil)
			mock.EXPECT().InspectContainer(gomock.Any()).Times(1).Return(&docker.Container{
				Mounts: []docker.Mount{},
			}, nil)
		},
		contains: ":/bin",
		err:      ErrorNoVolumeMounted,
	},
	{
		name: "server version not fetchable",
		expects: func(mock *mocks.MockClient, systemMock *mocks.MockENV) {
			mock.EXPECT().ListImages(gomock.Any()).Times(1).Return([]docker.APIImages{}, nil)
			mock.EXPECT().InspectContainer(gomock.Any()).Times(1).Return(&docker.Container{
				Mounts: []docker.Mount{
					docker.Mount{
						Destination: "/bin",
					},
				},
			}, nil)
			mock.EXPECT().Version().Times(1).Return(&docker.Env{}, errors.New("TestError"))
			systemMock.EXPECT().GetSystem().Times(1).Return("", true)
		},
		contains: "Could not fetch the server version",
		err:      ErrorServerVersionNotFound,
	},
	{
		name: "client version not fetchable",
		expects: func(mock *mocks.MockClient, systemMock *mocks.MockENV) {
			mock.EXPECT().ListImages(gomock.Any()).Times(1).Return([]docker.APIImages{}, nil)
			mock.EXPECT().InspectContainer(gomock.Any()).Times(1).Return(&docker.Container{
				Mounts: []docker.Mount{
					docker.Mount{
						Destination: "/bin",
					},
				},
			}, nil)
			mock.EXPECT().Version().Times(1).Return(&docker.Env{}, nil)
			mock.EXPECT().Info().Times(1).Return(&docker.DockerInfo{}, errors.New("TestError"))
			systemMock.EXPECT().GetSystem().Times(1).Return("", true)
		},
		contains: "Could not fetch the client info",
		err:      ErrorClientVersionNotFound,
	},
	{
		name: "system not detectable",
		expects: func(mock *mocks.MockClient, systemMock *mocks.MockENV) {
			mock.EXPECT().ListImages(gomock.Any()).Times(1).Return([]docker.APIImages{}, nil)
			mock.EXPECT().InspectContainer(gomock.Any()).Times(1).Return(&docker.Container{
				Mounts: []docker.Mount{
					docker.Mount{
						Destination: "/bin",
					},
				},
			}, nil)
			mock.EXPECT().Version().Times(1).Return(&docker.Env{"Os=foo3", "Arch=bar"}, nil)
			mock.EXPECT().Info().Times(1).Return(&docker.DockerInfo{}, nil)
			systemMock.EXPECT().GetSystem().Times(1).Return("", true)
		},
		contains: "We could not autodetect the system to install.",
		err:      ErrorSystemNotDetected,
	},
	{
		name: "selected version does not match",
		expects: func(mock *mocks.MockClient, systemMock *mocks.MockENV) {
			mock.EXPECT().ListImages(gomock.Any()).Times(1).Return([]docker.APIImages{}, nil)
			mock.EXPECT().InspectContainer(gomock.Any()).Times(1).Return(&docker.Container{
				Mounts: []docker.Mount{
					docker.Mount{
						Destination: "/bin",
					},
				},
			}, nil)
			systemMock.EXPECT().GetSystem().Times(1).Return("foobar", true)
		},
		contains: "The selected system is not available.",
		err:      ErrorSystemNotAvailable,
	},
	{
		name: "selected system can not be copied",
		expects: func(mock *mocks.MockClient, systemMock *mocks.MockENV) {
			mock.EXPECT().ListImages(gomock.Any()).Times(1).Return([]docker.APIImages{}, nil)
			mock.EXPECT().InspectContainer(gomock.Any()).Times(1).Return(&docker.Container{
				Mounts: []docker.Mount{
					docker.Mount{
						Destination: "/bin",
					},
				},
			}, nil)
			systemMock.EXPECT().GetSystem().Times(1).Return("foo2-bar", true)
		},
		contains: "Could not open the binary",
		err:      ErrorCopyFailed,
	},
	{
		name: "positive install - auto detect",
		expects: func(mock *mocks.MockClient, systemMock *mocks.MockENV) {
			mock.EXPECT().ListImages(gomock.Any()).Times(1).Return([]docker.APIImages{}, nil)
			mock.EXPECT().InspectContainer(gomock.Any()).Times(1).Return(&docker.Container{
				Mounts: []docker.Mount{
					docker.Mount{
						Destination: "/bin",
					},
				},
			}, nil)
			systemMock.EXPECT().GetSystem().Times(1).Return("foo-bar", true)
		},
		contains: "Sledgehammer installed, call it with 'slh'",
	},
	{
		name: "positive install - auto detect with data",
		expects: func(mock *mocks.MockClient, systemMock *mocks.MockENV) {
			mock.EXPECT().ListImages(gomock.Any()).Times(1).Return([]docker.APIImages{}, nil)
			mock.EXPECT().InspectContainer(gomock.Any()).Times(1).Return(&docker.Container{
				Mounts: []docker.Mount{
					docker.Mount{
						Destination: "/data",
					},
				},
			}, nil)
			systemMock.EXPECT().GetSystem().Times(1).Return("foo-bar", true)
		},
		contains: "Sledgehammer installed, call it with 'slh'",
	},
	{
		name: "positive install - custom system",
		expects: func(mock *mocks.MockClient, systemMock *mocks.MockENV) {
			mock.EXPECT().ListImages(gomock.Any()).Times(1).Return([]docker.APIImages{}, nil)
			mock.EXPECT().InspectContainer(gomock.Any()).Times(1).Return(&docker.Container{
				Mounts: []docker.Mount{
					docker.Mount{
						Destination: "/bin",
					},
					docker.Mount{
						Source:      "/var/run/docker.sock",
						Destination: "/var/run/docker.sock",
					},
				},
			}, nil)
			systemMock.EXPECT().GetSystem().Times(1).Return("", true)
			mock.EXPECT().Version().Times(1).Return(&docker.Env{"Os=foo", "Arch=bar"}, nil)
			mock.EXPECT().Info().Times(1).Return(&docker.DockerInfo{}, nil)
		},
		contains: "Sledgehammer installed, call it with 'slh'",
	},
}

func TestInstaller(t *testing.T) {

	for _, tt := range dockerTests {
		t.Run(tt.name, func(t *testing.T) {

			var b bytes.Buffer

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			dockerMock := mocks.NewMockClient(mockCtrl)
			envMock := mocks.NewMockENV(mockCtrl)

			db := test.NewTestDB(t)
			defer test.Close(db, t)

			path := test.NewTmpDir(t)
			defer test.DeleteTmpDir(path, t)

			outputPath := filepath.Join(path, "result")
			err := os.MkdirAll(outputPath, 0777)
			if err != nil {
				t.Fatal(t)
			}

			config := Config{
				Docker:      dockerMock,
				Env:         envMock,
				Writer:      &b,
				Systems:     []string{"foo-bar", "foo2-bar"},
				DB:          test.NewTestDB(t),
				WorkingPath: path,
				InstallPath: outputPath,
			}

			defer config.DB.Close()

			err = os.Mkdir(filepath.Join(path, "slh"), 0777)
			if err != nil {
				t.Fatal(t)
			}
			// create a dummy executable so we can test the happy path
			_, err = os.Create(filepath.Join(path, "slh/slh-foo-bar"))
			if err != nil {
				t.Fatal(t)
			}

			// expects
			tt.expects(dockerMock, envMock)

			// call
			err = InstallSledgehammer(config)

			if tt.err != nil {
				assert.Equal(t, tt.err, err)
			} else if err != nil {
				t.Fatal(err)
			}
			// assert
			assert.Contains(t, b.String(), tt.contains)
		})
	}
}
