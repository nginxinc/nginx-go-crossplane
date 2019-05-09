package analyzer

import (
	"testing"
)

func TestAnalyze(t *testing.T) {
	fname := "/path/to/nginx.conf"
	ctx := [3]string{"events"}

	// testing state directive

	// the state directive should not cause errors if it's in these contexts
	statement1 := Statement{directive: "state",
		args: [1]string{"/path/to/state/file.conf"},
		line: 5,
	}
	// the state directive should not cause errors if it's in these contexts
	goodContexts := [3][3]string{
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
		if err := Analyze(fname, statement1, ";", v2, true, true, false); err != nil {
			continue
		} else {
			t.Errorf("Not throwing an error on contexts")
		}
	}

	// test flag directive args

	// an NGINX_CONF_FLAG directive
	statement2 := Statement{
		directive: "accept_mutex",
		args:      [1]string{},
		line:      2,
	}

	goodArgs := [6][1]string{
		{"on"},
		{"off"},
		{"On"},
		{"Off"},
		{"ON"},
		{"OFF"},
	}

	for _, v := range goodArgs {
		statement2.args = v
		if err := Analyze(fname, statement2, ";", ctx, true, false, true); err != nil {
			t.Errorf("Throwing an error on good args: %v", v)
		}

	}
	badArgs := [5][1]string{
		{"1"},
		{"0"},
		{"true"},
		{"okay"},
		{""},
	}

	for _, v := range badArgs {
		statement2.args = v
		if err := Analyze(fname, statement2, ";", ctx, true, false, true); err != nil {
			continue
		} else {
			t.Errorf("Not failing on bad args: %v", v)
		}
	}

}
