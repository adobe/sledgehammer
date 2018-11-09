/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package kit_test

import (
	"testing"

	"github.com/adobe/sledgehammer/slh/tool"

	"github.com/adobe/sledgehammer/slh/kit"
	"github.com/stretchr/testify/assert"
)

func TestParseTool(t *testing.T) {
	cases := []struct {
		name     string
		toParse  string
		expected *kit.Tool
		err      error
	}{
		{
			name:    "empty tool",
			toParse: "",
			err:     tool.ErrorNameInvalid,
		},
		{
			name:    "illegal tool name",
			toParse: "fo/bar:latest",
			err:     tool.ErrorNameInvalid,
		},
		{
			name:    "malformed input",
			toParse: "fo:bar:latest",
			err:     &kit.ParseError{ToParse: "fo:bar:latest"},
		},
		{
			name:     "empty version",
			toParse:  "foobar:",
			expected: &kit.Tool{Name: "foobar", Version: "latest"},
		},
		{
			name:     "success #1",
			toParse:  "foobar:*",
			expected: &kit.Tool{Name: "foobar", Version: "*"},
		},
		{
			name:     "success #2",
			toParse:  "foobar:edge",
			expected: &kit.Tool{Name: "foobar", Version: "edge"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			output, err := kit.ParseTool(tt.toParse)
			if tt.err != nil {
				assert.Equal(t, tt.err.Error(), err.Error())
				return
			}
			assert.Equal(t, tt.expected, output)
		})
	}
}
