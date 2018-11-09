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

// LocalFactory is the factory for the LocalRegistry
type LocalFactory struct{}

// Raw will return a raw LocalRepository struct for populating from the db
func (g *LocalFactory) Raw() Tool {
	return &LocalTool{}
}

// Create will take data and return a LocalRegistry from the given arguments
func (g *LocalFactory) Create(dt Data) Tool {
	return &LocalTool{
		Core: dt,
	}
}

// LocalTool represents a tool that can only be found on the local computer, this should not be used in prodution
// It is mainly used for local development
type LocalTool struct {
	Core Data `json:"code"`
}

// Versions will return all available remote versions, in this case none
func (t *LocalTool) Versions() ([]string, error) {
	return []string{}, nil
}

// Data will return the inner data for the tool
func (t *LocalTool) Data() *Data {
	return &t.Core
}
