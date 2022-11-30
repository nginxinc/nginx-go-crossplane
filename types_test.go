/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package crossplane

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirective_String(t *testing.T) {
	for _, tf := range []struct {
		directive *Directive
		expected  string
	}{
		{
			directive: &Directive{
				Directive: "location",
				Args:      []string{"/foo"},
			},
			expected: "location /foo",
		},
		{
			directive: &Directive{
				Directive: "location",
				Args:      []string{"~", "\\.(gif|jpg|png)$"},
				Block: []*Directive{
					{
						Directive: "root",
						Args:      []string{"/data/images"},
					},
				},
			},
			expected: "location ~ \\.(gif|jpg|png)$ {...}",
		},
	} {
		s := tf.directive.String()
		assert.Equal(t, s, tf.expected)
	}
}

// nolint:funlen
func TestDirective_Equal(t *testing.T) {
	commentPtr := pStr("foo")

	for _, ef := range []struct {
		a     *Directive
		b     *Directive
		equal bool
	}{
		{
			a: &Directive{
				Directive: "location",
			},
			b: &Directive{
				Directive: "location",
			},
			equal: true,
		},
		{
			a: &Directive{
				Directive: "location",
			},
			b: &Directive{
				Directive: "server",
			},
			equal: false,
		},
		{
			a: &Directive{
				Directive: "location",
				Args:      []string{},
			},
			b: &Directive{
				Directive: "location",
				Args:      []string{"b"},
			},
			equal: false,
		},
		{
			a: &Directive{
				Directive: "location",
				Args:      []string{"b"},
			},
			b: &Directive{
				Directive: "location",
				Args:      []string{"b"},
			},
			equal: true,
		},
		{
			a: &Directive{
				Directive: "location",
				Line:      1,
				Args:      []string{"b"},
				File:      "",
			},
			b: &Directive{
				Directive: "location",
				Args:      []string{"b"},
			},
			equal: false,
		},
		{
			a: &Directive{
				Directive: "location",
				Line:      1,
				Args:      []string{"b"},
				File:      "",
			},
			b: &Directive{
				Directive: "location",
				Args:      []string{"b"},
				Line:      1,
			},
			equal: true,
		},
		{
			a: &Directive{
				Directive: "location",
				Line:      1,
				Args:      []string{"b"},
				File:      "/",
			},
			b: &Directive{
				Directive: "location",
				Args:      []string{"b"},
				Line:      1,
			},
			equal: false,
		},
		{
			a: &Directive{
				Directive: "location",
				Args:      []string{"b"},
				Block:     []*Directive{},
			},
			b: &Directive{
				Directive: "location",
				Args:      []string{"b"},
				Block:     []*Directive{},
			},
			equal: true,
		},
		{
			a: &Directive{
				Directive: "location",
				Args:      []string{"c"},
				Block:     []*Directive{},
			},
			b: &Directive{
				Directive: "location",
				Args:      []string{"b"},
				Block:     []*Directive{},
			},
			equal: false,
		}, {
			a: &Directive{
				Directive: "location",
				Args:      []string{"b"},
				Block: []*Directive{
					{
						Directive: "root",
						Args:      []string{"/data/images"},
					},
				},
			},
			b: &Directive{
				Directive: "location",
				Args:      []string{"b"},
				Block:     []*Directive{},
			},
			equal: false,
		},
		{
			a: &Directive{
				Directive: "location",
				Comment:   pStr("a"),
			},
			b: &Directive{
				Directive: "location",
			},
			equal: false,
		},
		{
			a: &Directive{
				Directive: "location",
				Comment:   pStr("a"),
			},
			b: &Directive{
				Directive: "location",
				Comment:   pStr("a"),
			},
			equal: true,
		},
		{
			a: &Directive{
				Directive: "location",
				Comment:   pStr("b"),
			},
			b: &Directive{
				Directive: "location",
				Comment:   pStr("a"),
			},
			equal: false,
		},
		{
			a: &Directive{
				Directive: "location",
				Comment:   commentPtr,
			},
			b: &Directive{
				Directive: "location",
				Comment:   commentPtr,
			},
			equal: true,
		},
		{
			a: &Directive{
				Directive: "location",
				Includes:  []int{},
			},
			b: &Directive{
				Directive: "location",
			},
			equal: true,
		},
		{
			a: &Directive{
				Directive: "location",
				Includes:  []int{},
			},
			b: &Directive{
				Directive: "location",
				Includes:  []int{1},
			},
			equal: false,
		},
		{
			a: &Directive{
				Directive: "location",
				Includes:  []int{1},
			},
			b: &Directive{
				Directive: "location",
				Includes:  []int{1},
			},
			equal: true,
		},
		{
			a: &Directive{
				Directive: "location",
				Includes:  []int{19},
			},
			b: &Directive{
				Directive: "location",
				Includes:  []int{1},
			},
			equal: false,
		},
		{
			a: nil,
			b: &Directive{
				Directive: "location",
				Includes:  []int{1},
			},
			equal: false,
		},
		{
			a: &Directive{
				Directive: "location",
				Includes:  []int{1},
			},
			b:     nil,
			equal: false,
		},
		{
			a:     nil,
			b:     nil,
			equal: true,
		},
	} {
		eq := ef.a.Equal(ef.b)
		if eq != ef.equal {
			t.Logf("a=%#v, b=%#v", ef.a, ef.b)
		}

		assert.Equal(t, eq, ef.equal)
	}
}
