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

type List struct {
	Name     string
	Elements []interface{}
}

func (l *List) Table(tw *tabwriter.Writer) {
	for i := 0; i < len(l.Elements); i++ {
		if i == 0 && len(l.Name) > 0 {
			tw.Write([]byte(fmt.Sprintf("%s\t%s\n", l.Name, l.Elements[i])))
		} else {
			tw.Write([]byte(fmt.Sprintf("\t%s\n", l.Elements[i])))
		}
	}
}

func (l *List) JSON(ms map[string]interface{}) {
	ms[strings.ToLower(l.Name)] = l.Elements
}

func NewList(name string) *List {
	return &List{
		Name:     name,
		Elements: []interface{}{},
	}
}

func (l *List) Add(el interface{}) {
	l.Elements = append(l.Elements, el)
}
