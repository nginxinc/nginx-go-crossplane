package crossplane

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorString(t *testing.T) {
	t.Parallel()
	tf := "test.conf"
	tl := 1

	tcs := []struct {
		f *string
		w string
		l *int

		exp string
	}{
		{&tf, "error", &tl, "error in test.conf:1"},
		{nil, "error", &tl, "error in (nofile):1"},
		{nil, "error", nil, "error in (nofile)"},
		{&tf, "error", nil, "error in test.conf"},
		{nil, "", nil, " in (nofile)"},
	}

	for _, tc := range tcs {
		e := &ParseError{File: tc.f, Line: tc.l, What: tc.w}

		assert.Equal(t, tc.exp, e.Error())
	}
}
