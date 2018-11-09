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

type Table struct {
	Name       string
	Headers    []string
	Rows       [][]interface{}
	mergeCells string
	hideHeader bool
}

func (t *Table) MergeCells(replace string) {
	t.mergeCells = replace
}

func (t *Table) HideHeader() {
	t.hideHeader = true
}

func (t *Table) Table(tw *tabwriter.Writer) {
	if !t.hideHeader {
		for _, r := range t.Headers {
			tw.Write([]byte(fmt.Sprintf("%s\t", r)))
		}
		tw.Write([]byte("\n"))
		for _, r := range t.Headers {
			tw.Write([]byte(strings.Repeat("-", len(r)) + "\t"))
		}
		tw.Write([]byte("\n"))
	}
	for y := 0; y < len(t.Rows); y++ {
		for x := 0; x < len(t.Rows[y]); x++ {
			if len(t.mergeCells) > 0 && y-1 >= 0 && t.Rows[y-1][x] == t.Rows[y][x] {
				tw.Write([]byte(fmt.Sprintf("%v\t", t.mergeCells)))
			} else {
				tw.Write([]byte(fmt.Sprintf("%v\t", t.Rows[y][x])))
			}
		}
		tw.Write([]byte("\n"))
	}
}

func (t *Table) Add(cols ...interface{}) {
	t.Rows = append(t.Rows, cols)
}

func (t *Table) JSON(ms map[string]interface{}) {
	jsonTable := []interface{}{}
	for _, r := range t.Rows {
		jsonRow := map[string]interface{}{}
		for i, c := range r {
			jsonRow[strings.Replace(strings.ToLower(t.Headers[i]), " ", "_", -1)] = c
		}
		jsonTable = append(jsonTable, jsonRow)
	}
	ms[strings.ToLower(t.Name)] = jsonTable
}

func NewTable(name string, headers ...string) *Table {
	return &Table{
		Name:    name,
		Headers: headers,
		Rows:    [][]interface{}{},
	}
}
