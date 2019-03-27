package analyzer

import (
	"testing"
)

type statement struct {
	directive string
	args      [1]string
	line      int
}

func TestAnalyze(t *testing.T) {
	fname := "/path/to/nginx.conf"
	ctx := [3]string{"events"}

	// testing state directive

	// the state directive should not cause errors if it's in these contexts
	statement1 := statement{directive: "state",
		args: [1]string{"/path/to/state/file.conf"},
		line: 5,
	}
	// the state directive should not cause errors if it's in these contexts
	goodContexts := [3][3]string{
		[3]string{"http", "upstream"},
		[3]string{"stream", "upstream"},
		[3]string{"some_third_part_context"},
	}

	for _, v1 := range goodContexts {
		analyze(fname, statement1, ";", v1, true, true, false)
	}

	//the state directive should not be in any of these contexts
	badContext := [11][3]string{
		[3]string{"events"},
		[3]string{"mail"},
		[3]string{"mail", "server"},
		[3]string{"stream"},
		[3]string{"stream", "server"},
		[3]string{"http"},
		[3]string{"http", "server"},
		[3]string{"http", "location"},
		[3]string{"http", "server", "if"},
		[3]string{"http", "location", "if"},
		[3]string{"http", "location", "limit_except"},
	}

	for _, v2 := range badContext {
		if err := analyze(fname, statement1, ";", v2, true, true, false); err != nil {
			t.Errorf("Error %v", err)
		}
	}

	// test flag directive args

	// an NGINX_CONF_FLAG directive
	statement2 := statement{directive: "accept_mutex",
		args: [1]string{},
		line: 2,
	}

	goodArgs := [6][1]string{
		[1]string{"on"},
		[1]string{"off"},
		[1]string{"On"},
		[1]string{"Off"},
		[1]string{"ON"},
		[1]string{"OFF"},
	}

	for _, v := range goodArgs {
		statement2.args = v
		analyze(fname, statement2, ";", ctx, true, false, true)

	}
	badArgs := [5][1]string{
		[1]string{"1"},
		[1]string{"0"},
		[1]string{"true"},
		[1]string{"okay"},
		[1]string{""},
	}

	for _, v := range badArgs {
		statement2.args = v
		if err := analyze(fname, statement2, ";", ctx, true, false, true); err != nil {
			t.Errorf("Error %v", err)
		}
	}

}
