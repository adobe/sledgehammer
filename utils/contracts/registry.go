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

// Registry - Defines how a registry has to be structured to be recognized by sledgehammer
type Registry struct {
	Description string `json:"description,omitempty"`
	Maintainer  string `json:"maintainer,omitempty"`
	Tools       []Tool `json:"tools,omitempty"`
	Kits        []Kit  `json:"kits,omitempty"`
}
