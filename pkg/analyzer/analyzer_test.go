package analyzer

import (
	"testing"
)

func TestAnalyze(t *testing.T) {
	fname := "/path/to/nginx.conf"
	ctx := [3]string{"http", "upstream"}

	// testing state Directive

	// the state Directive should not cause errors if it's in these contexts
	statement1 := Statement{Directive: "state",
		Args: []string{"/path/to/state/file.conf"},
		Line: 5,
	}
	// the state Directive should not cause errors if it's in these contexts
	goodContexts := [2][3]string{
		{"http", "upstream"},
		{"stream", "upstream"},
	}

	for _, v1 := range goodContexts {
		if err := Analyze(fname, statement1, ";", v1, true, true, false); err != nil {
			t.Errorf("Throwing an error on contexts %v", v1)
		}
	}

	badContext := [5][3]string{
		{"noevents"},
		{"femail"},
		{"femail", "waitress"},
		{"origin"},
		{"https"},
	}
	for _, v2 := range badContext {
		if err := Analyze(fname, statement1, ";", v2, true, true, false); err == nil {
			t.Error("Not throwing an error on contexts : ", v2)
		}

	}

	// test flag Directive Args

	// an NGINX_CONF_FLAG Directive
	statement2 := Statement{
		Directive: "accept_mutex",
		Args:      []string{},
		Line:      2,
	}

	goodArgs := [6][]string{
		{"on"},
		{"off"},
		{"On"},
		{"Off"},
		{"ON"},
		{"OFF"},
	}

	for _, v := range goodArgs {
		statement2.Args = v
		if err := Analyze(fname, statement2, ";", ctx, true, false, true); err != nil {
			t.Errorf("Throwing an error on good Args: %v", v)
		}

	}
	badArgs := [4][]string{
		{""},
		{"0"},
		{"true"},
		{"okay"},
	}

	for _, v := range badArgs {
		statement2.Args = v
		if err := Analyze(fname, statement2, ";", ctx, true, false, true); err == nil {
			t.Errorf("Not failing on bad Args: %v", v)
		}
	}
}
