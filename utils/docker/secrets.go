/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package docker

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"

	cred "github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/credentials"
	client "github.com/fsouza/go-dockerclient"
)

// SecretJSON represents the credsStore entry in a docker config
type SecretJSON struct {
	CredsStore string `json:"credsStore"`
	// TODO: Implement credsHelper per registry
	CredsHelper map[string]string `json:"credsHelper"`
}

// credStore will determine if the user has a credsStore defined or not
func credStore() (string, error) {
	credsStore := SecretJSON{}
	paths := cfgPaths(os.Getenv("DOCKER_CONFIG"), os.Getenv("HOME"))
	for _, path := range paths {
		r, err := os.Open(path)
		if err == nil {
			// parse json
			b, _ := ioutil.ReadAll(r)
			_ = json.Unmarshal(b, &credsStore)
			if len(credsStore.CredsStore) > 0 {
				return credsStore.CredsStore, nil
			}
		}
	}
	return "", errors.New("No credStore defined")
}

// GetCredentials is the main function to call when credentials are required,
// it will switch between the config or the credentials store to fetch the requested credentials
func GetCredentials(server string) (*credentials.Credentials, error) {
	credStore, err := credStore()
	if err == nil {
		return getSecretFromCredStore(credStore, server)
	}
	auth, err := client.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		return nil, err
	}
	serverConfig, found := auth.Configs[server]
	if found {
		return &credentials.Credentials{
			Secret:    serverConfig.Password,
			ServerURL: serverConfig.ServerAddress,
			Username:  serverConfig.Username,
		}, nil
	}
	return nil, errors.New("No credentials found")
}

func getSecretFromCredStore(store string, server string) (*credentials.Credentials, error) {
	credFunc := cred.NewShellProgramFunc("docker-credential-" + store)
	creds, err := cred.Get(credFunc, server)
	if err != nil {
		return nil, err
	}
	return creds, nil
}

func cfgPaths(dockerConfigEnv string, homeEnv string) []string {
	var paths []string
	if dockerConfigEnv != "" {
		paths = append(paths, path.Join(dockerConfigEnv, "config.json"))
	}
	if homeEnv != "" {
		paths = append(paths, path.Join(homeEnv, ".docker", "config.json"))
		paths = append(paths, path.Join(homeEnv, ".dockercfg"))
	}
	return paths
}
