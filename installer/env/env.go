/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package env

import (
	"os"
)

// ENV is the interface to get the selected system variable from the installation
type ENV interface {
	GetSystem() (string, bool)
}

// OSNEV is the default that will just return the environment variable
type OSENV struct{}

// GetSystem will return the content of the SYSTEM environment variable
func (st *OSENV) GetSystem() (string, bool) {
	return os.LookupEnv("SYSTEM")
}
