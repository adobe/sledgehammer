/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package contracts

// Tool - Defines how a tool needs to be represented so that sledgehmmer can recognize it
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Registry    string      `json:"registry,omitempty"`
	Image       string      `json:"image"`
	Entry       []string    `json:"entry,omitempty"`
	Type        string      `json:"type,omitempty"`
	Daemon      *ToolDaemon `json:"daemon,omitempty"`
}

// ToolDaemon defines a tool as daemon. The entry will be the main entrypoint that will be called to keep the container in a daemon state.
type ToolDaemon struct {
	// The daemon is expected to terminate itself after a given amount of time.
	// This time should be higher than the stated ttl, sledgehammer will make sure
	Entry []string `json:"entry,omitempty"`
	// TTL   int      `json:"ttl,omitempty"`
}
