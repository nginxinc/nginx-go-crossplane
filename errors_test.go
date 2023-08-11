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

func TestErrorString(t *testing.T) {
	t.Parallel()
	testconfig := "test.conf"
	line := 1
	var err error

	tcs := []struct {
		f *string
		w string
		l *int
		e error

		exp string
	}{
		{&testconfig, "error", &line, err, "error in test.conf:1"},
		{nil, "error", &line, err, "error in (nofile):1"},
		{nil, "error", nil, err, "error in (nofile)"},
		{&testconfig, "error", nil, err, "error in test.conf"},
		{nil, "", nil, err, " in (nofile)"},
	}

	for _, tc := range tcs {
		e := &ParseError{File: tc.f, Line: tc.l, What: tc.w, originalErr: tc.e}

		assert.Equal(t, tc.exp, e.Error())
	}
}
