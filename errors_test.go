package crossplane

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorString(t *testing.T) {
	t.Parallel()
	tf := "test.conf"
	tl := 1
	var te error

	tcs := []struct {
		f *string
		w string
		l *int
		e error

		exp string
	}{
		{&tf, "error", &tl, te, "error in test.conf:1"},
		{nil, "error", &tl, te, "error in (nofile):1"},
		{nil, "error", nil, te, "error in (nofile)"},
		{&tf, "error", nil, te, "error in test.conf"},
		{nil, "", nil, te, " in (nofile)"},
	}

	for _, tc := range tcs {
		e := &ParseError{File: tc.f, Line: tc.l, What: tc.w, originalErr: tc.e}

		assert.Equal(t, tc.exp, e.Error())
	}
}
