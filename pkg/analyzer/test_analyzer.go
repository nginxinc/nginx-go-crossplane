package analyzer

import (
	"testing"
)

type statement struct {
	directive string
	args      [1]string
	line      int
}

func TestAnalyzer(t *testing.T) {
	fname := "/path/to/nginx.conf"
	ctx := []string{"events"}
	a := newAnalyzer()

	// testing state directive

	// the state directive should not cause errors if it's in these contexts
	statement1 := statement{directive: "state",
		args: [1]string{"/path/to/state/file.conf"},
		line: 5,
	}
	// the state directive should not cause errors if it's in these contexts
	goodContexts := [3][]string{
		[]string{"http", "upstream"},
		[]string{"stream", "upstream"},
		[]string{"some_third_part_context"},
	}

	for _, v1 := range goodContexts {
		a.analyse(fname, statement1, v1)
	}

	//the state directive should not be in any of these contexts
	badContext := [11][]string{
		[]string{"events"},
		[]string{"mail"},
		[]string{"mail", "server"},
		[]string{"stream"},
		[]string{"stream", "server"},
		[]string{"http"},
		[]string{"http", "server"},
		[]string{"http", "location"},
		[]string{"http", "server", "if"},
		[]string{"http", "location", "if"},
		[]string{"http", "location", "limit_except"},
	}

	for _, v2 := range badContext {
		if err := a.analyse(fname, statement1, v2); err != nil {
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
		a.analyse(fname, statement2, ctx)

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
		if err := a.analyse(fname, statement2, ctx); err != nil {
			t.Errorf("Error %v", err)
		}
	}

}
