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

import "text/tabwriter"

type Container struct {
	Name     string
	Elements []Renderer
}

func (b *Container) Table(tw *tabwriter.Writer) {
	for _, e := range b.Elements {
		e.Table(tw)
	}
}

func (b *Container) JSON(mp map[string]interface{}) {
	for _, e := range b.Elements {
		e.JSON(mp)
	}
}

func (b *Container) Add(el Renderer) {
	b.Elements = append(b.Elements, el)
}

func NewContainer(name string) *Container {
	return &Container{
		Name:     name,
		Elements: []Renderer{},
	}
}
