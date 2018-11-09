/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package tool

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/docker/docker-credential-helpers/credentials"

	"github.com/adobe/sledgehammer/utils/docker"
	"github.com/sirupsen/logrus"
)

// JFrogFactory is the factory for the JFrogTool
type JFrogFactory struct{}

// Raw will return a raw JFrogFactory struct for populating from the db
func (g *JFrogFactory) Raw() Tool {
	return &JFrogTool{}
}

// Create will take data and return a JFrogTool from the given arguments
func (g *JFrogFactory) Create(dt Data) Tool {
	return &JFrogTool{
		Core: dt,
	}
}

// JFrogTool represents a sledgehammer tool which is stored in any JFrog artifacory repository
type JFrogTool struct {
	Core Data `json:"code"`
}

// JFrogTagResponse is the response the repository returns when asked for tags
type JFrogTagResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

var (
	// JFrogTagsURL is the URL that will return all tags for a given image
	JFrogTagsURL = "%s/artifactory/v2/%s/tags/list"
)

// Data will return the inner data for the tool
func (t *JFrogTool) Data() *Data {
	return &t.Core
}

// Versions returns the version of the tool while it fetches them from the official docker repository
func (t *JFrogTool) Versions() ([]string, error) {
	logrus.Info("Checking remote image versions of JFrogTool")
	versions := []string{}
	if len(t.Data().ImageRegistry) == 0 {
		return versions, nil
	}

	var httpClient = &http.Client{
		Timeout: time.Second * 10,
	}

	creds, err := docker.GetCredentials(t.Data().ImageRegistry)
	if err != nil {
		logrus.WithField("registry", FullImage(t, "")).Warnln(err.Error())
	}

	versions, err = t.getTags(httpClient, creds)
	if err != nil {
		return versions, err
	}
	return versions, nil
}

func (t *JFrogTool) getTags(client *http.Client, creds *credentials.Credentials) ([]string, error) {
	url := fmt.Sprintf(JFrogTagsURL, t.Data().ImageRegistry, t.Data().Image)
	if !strings.HasPrefix(url, "http") {
		url = fmt.Sprintf("http://%s", url)
	}

	versions := []string{}
	logrus.WithField("url", url).Debug("Fetching tag page")
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return versions, err
	}
	if creds != nil && len(creds.Secret) > 0 {
		req.Header.Add("X-JFrog-Art-Api", creds.Secret)
	}
	resp, err := client.Do(req)
	if err != nil {
		return versions, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return versions, err
	}

	// check for any error
	if resp.StatusCode != 200 {
		return versions, errors.New(string(body))
	}

	tagResponse := JFrogTagResponse{}
	err = json.Unmarshal(body, &tagResponse)
	if err != nil {
		return versions, err
	}
	versions = tagResponse.Tags
	return versions, nil
}
