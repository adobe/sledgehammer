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

// Kit - Defines how a collection of tools, called kit should be structured
type Kit struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Tools       []KitTool `json:"tools,omitempty"`
}

// KitTool is an entry in a tool kit.
type KitTool struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Alias   string `json:"alias,omitempty"`
}
