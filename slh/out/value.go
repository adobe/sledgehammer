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
	"strings"
	"text/tabwriter"
)

func NewValue(name string, element interface{}) *Value {
	return &Value{
		Name:    name,
		Element: element,
	}
}

type Value struct {
	Name    string
	Element interface{}
}

func (v *Value) Table(tw *tabwriter.Writer) {
	tw.Write([]byte(fmt.Sprintf("%v\t%v\n", v.Name, v.Element)))
}

func (v *Value) JSON(ms map[string]interface{}) {
	ms[strings.Replace(strings.ToLower(v.Name), " ", "_", -1)] = v.Element
}
