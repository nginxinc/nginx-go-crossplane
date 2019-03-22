package analyse

import (
	"testing"
)

func TestAnalyzer(t *testing.T) {
	fname := "/path/to/nginx.conf"
	ctx := [1]string{"events"}

	// testing state directive

	// the state directive should not cause errors if it's in these contexts
	statement1 := struct {
		directive string
		args      [1]string
		line      int // this is arbitrary
	}{
		"state",
		[1]string{"/path/to/state/file.conf"},
		5,
	}
	// the state directive should not cause errors if it's in these contexts
	goodContexts := [3][]string{
		[]string{"http", "upstream"},
		[]string{"stream", "upstream"},
		[]string{"some_third_part_context"},
	}

	for _, v1 := range goodContexts {
		crossplane.analyzer.analyze(fname, statement1, ";", v1)
	}

	//the state directive should not be in any of these contexts
	badContext := []string{}
	cntx := crossplane.analyser.CONTEXTS
	next := 0
	for key, value := range cntx {
		isIn := false
		for v := range goodContexts {
			if v == key {
				isIn = true
				break
			}
		}
		if isIn {
			badContext[next] = key
			next++
		}

	}

	for _, v2 := range badContext {
		err := crossplane.analyzer.analyze(fname, statement1, ";", v2)
		if err != nil {
			t.Errorf("Error %v", err)
		}
	}

	// test flag directive args

	// an NGINX_CONF_FLAG directive
	statement2 := struct {
		directive string
		args      [1]string
		line      int // this is arbitrary
	}{
		"accept_mutex",
		[1]string{},
		2,
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
		crossplane.analyzer.analyze(fname, statement2, ";", ctx)

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
		err := crossplane.analyzer.analyze(fname, statement2, ";", ctx)
		if err != nil {
			t.Errorf("Error %v", err)
		}
	}

}
