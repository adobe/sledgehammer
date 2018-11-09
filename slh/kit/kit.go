/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package kit

import (
	"errors"
	"strings"

	"github.com/adobe/sledgehammer/slh/tool"
)

var (
	// ErrorKitNotFound will be thrown if the tool kit cannot be found
	ErrorKitNotFound = errors.New("Tool kit could not be found")
)

// ParseError will be thrown if the kit cannot be parsed
type ParseError struct {
	ToParse string
}

func (p *ParseError) Error() string {
	return "Could not parse '" + p.ToParse + "'"
}

// Kit is a collection of tools bundled under a common name.
type Kit struct {
	Name        string
	Description string
	Tools       []Tool
}

// Tool is the entry in a kit, it will determine the tool and the version to use
type Tool struct {
	Name    string
	Version string
	Alias   string
}

// ParseTool will parse the tool string of a kit and returns the name and the version of the tool to use
func ParseTool(toParse string) (*Tool, error) {
	splitted := strings.Split(toParse, ":")
	var name string
	var version string
	if len(splitted) > 2 {
		return nil, &ParseError{ToParse: toParse}
	}
	if len(splitted) > 0 {
		name = splitted[0]
	}
	version = "latest"
	if len(splitted) > 1 {
		version = splitted[1]
		if version == "" {
			version = "latest"
		}
	}
	if !tool.HasValidName(name) {
		return nil, tool.ErrorNameInvalid
	}
	return &Tool{
		Name:    name,
		Version: version,
	}, nil
}
