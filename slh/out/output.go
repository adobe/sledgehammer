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
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
)

type Renderer interface {
	Table(*tabwriter.Writer)
	JSON(map[string]interface{})
}

type Output struct {
	Writer     io.Writer
	RenderFunc func()
	Element    Renderer
	ExitCode   int
}

func (o *Output) Render() {
	if o.Element != nil {
		o.RenderFunc()
	}
}

func (l *Output) RenderTable() {
	tw := &tabwriter.Writer{}
	tw = tw.Init(l.Writer, 0, 0, 3, ' ', 0)
	l.Element.Table(tw)
	tw.Flush()
}

func (l *Output) RenderJSON() {
	jsonMap := map[string]interface{}{}
	l.Element.JSON(jsonMap)
	bb, _ := json.MarshalIndent(jsonMap, "", "  ")
	fmt.Fprint(l.Writer, string(bb)+"\n")
}

func (o *Output) Set(el Renderer) {
	o.Element = el
}
