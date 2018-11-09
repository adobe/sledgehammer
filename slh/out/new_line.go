/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package out

import (
	"fmt"
	"text/tabwriter"
)

func NewNewLine() *NewLine {
	return &NewLine{}
}

type NewLine struct{}

func (v *NewLine) Table(tw *tabwriter.Writer) {
	tw.Write([]byte(fmt.Sprintf("\n")))
}

func (v *NewLine) JSON(ms map[string]interface{}) {
}
