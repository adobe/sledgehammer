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
	"time"
)

// Data is the data each tool contains
type Data struct {
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Registry      string    `json:"registry"`
	ImageRegistry string    `json:"imageRegistry,omitempty"`
	Image         string    `json:"image"`
	Default       bool      `json:"default"`
	Type          string    `json:"type,omitempty"`
	Entry         []string  `json:"entry,omitempty"`
	Added         time.Time `json:"added"`
	Versions      []string  `json:"versions,omitempty"`
	Daemon        *Daemon   `json:"daemon,omitempty"`
}

// Daemon defines the entry point when the container should be started as a daemon
type Daemon struct {
	Entry []string `json:"entry,omitempty"`
}

// FullImage will return the full name of the image including repository and version if possible
func FullImage(tool Tool, version string) string {
	var fullName string
	if len(tool.Data().ImageRegistry) > 0 {
		fullName += tool.Data().ImageRegistry + "/"
	}
	fullName += tool.Data().Image
	if len(version) > 0 {
		fullName += ":" + version
	}
	return fullName
}
