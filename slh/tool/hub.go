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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/adobe/sledgehammer/utils/docker"
	"github.com/sirupsen/logrus"
)

// HubFactory is the factory for the HubTool
type HubFactory struct{}

// Raw will return a raw HubFactory struct for populating from the db
func (g *HubFactory) Raw() Tool {
	return &HubTool{}
}

// Create will take data and return a HubFactory from the given arguments
func (g *HubFactory) Create(dt Data) Tool {
	return &HubTool{
		Core: dt,
	}
}

// HubTool represents a sledgehammer tool which is stored in the official docker hub repository
type HubTool struct {
	Core Data `json:"code"`
}

// HubTagResponse is the response the repository returns when asked for tags
type HubTagResponse struct {
	Next    string         `json:"next"`
	Results []HubTagResult `json:"results"`
}

// HubTokenResponse is the response for a login request
type HubTokenResponse struct {
	Token string `json:"token"`
}

// HubTagResult is the tag name and part of the TagResult
type HubTagResult struct {
	Name string `json:"name"`
}

var (
	// HubHost is the main URL for the repository
	HubHost = "https://index.docker.io/v1/"
	// HubLoginURL is the URL were we can login into the repository
	HubLoginURL = "https://hub.docker.com/v2/users/login/"
	// HubTagsURL is the url where we can fetch the tags of an image
	HubTagsURL = "https://hub.docker.com/v2/repositories/%s/tags/?page=%d&page_size=%d"
)

// Data will return the inner data for the tool
func (t *HubTool) Data() *Data {
	return &t.Core
}

// Versions returns the version of the tool while it fetches them from the official docker repository
func (t *HubTool) Versions() ([]string, error) {
	logrus.Info("Checking remote image versions of HubTool")
	versions := []string{}
	var httpClient = &http.Client{
		Timeout: time.Second * 10,
	}

	token, err := t.getToken(httpClient)
	if err != nil {
		return versions, err
	}

	versions, err = t.getTags(httpClient, token)
	if err != nil {
		return versions, err
	}
	return versions, nil
}

func (t *HubTool) getToken(client *http.Client) (string, error) {
	// login to docker hub with credentials if they can be found
	// get credentials for https://index.docker.io/v1/
	logrus.WithField("image", t.Core.Image).Info("Trying to get token")

	creds, err := docker.GetCredentials(HubHost)
	if err != nil {
		logrus.WithField("registry", FullImage(t, "")).Warnln(err.Error())
	}
	var resp *http.Response
	if creds != nil {
		logrus.WithField("registry", FullImage(t, "")).WithField("username", creds.Username).Info("Found credentials for registry")
		values := map[string]string{"username": creds.Username, "password": creds.Secret}
		jsonValue, _ := json.Marshal(values)
		resp, err = client.Post(HubLoginURL, "application/json", bytes.NewBuffer(jsonValue))
	} else {
		resp, err = client.Post(HubLoginURL, "application/json", bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	tokenResponse := HubTokenResponse{}
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return "", err
	}
	if len(tokenResponse.Token) > 15 {
		logrus.WithField("token", tokenResponse.Token[:10]).Info("Got token")
	}
	return tokenResponse.Token, nil
}

func (t *HubTool) getTags(client *http.Client, token string) ([]string, error) {
	imageName := t.Core.Image
	if !strings.Contains(t.Core.Image, "/") {
		imageName = "library/" + t.Core.Image
	}
	url := fmt.Sprintf(HubTagsURL, imageName, 1, 100)
	return t.getTagsFromPage(client, url, token)
}

func (t *HubTool) getTagsFromPage(client *http.Client, url string, token string) ([]string, error) {
	versions := []string{}
	logrus.WithField("url", url).Debug("Fetching tag page")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return versions, err
	}
	if len(token) > 0 {
		req.Header.Add("Authorization", "JWT "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return versions, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return versions, err
	}
	hubTagResponse := HubTagResponse{}
	err = json.Unmarshal(body, &hubTagResponse)
	if err != nil {
		return versions, err
	}
	for _, tag := range hubTagResponse.Results {
		versions = append(versions, tag.Name)
	}

	if hubTagResponse.Next != "null" && len(hubTagResponse.Next) > 0 {
		nextPageVersions, err := t.getTagsFromPage(client, hubTagResponse.Next, token)
		if err != nil {
			return versions, err
		}
		logrus.WithField("versions", nextPageVersions).Debug("Fetched tags from page")
		versions = append(versions, nextPageVersions...)
	}
	return versions, nil
}
