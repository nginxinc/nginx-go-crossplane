package main 

import (
	"fmt"
	"log"
)

func main(){
	test_state_directive()
}

func test_state_directive(){
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
	good_contexts := [3][]string{
		[2]string{"http","upstream"}, 
		[2]string{"stream","upstream"}, 
		[1]string{"some_third_part_context"},
	}
	
	for _,v1 := range good_contexts{
		err := crossplane.analyzer.analyze(fname, statement1,";", v1)
	}

	//the state directive should not be in any of these contexts
	bad_context := []

	for _,v2 := range bad_contexts{
		err := crossplane.analyzer.analyze(fname, statement1, ";", v2)
		if err != nil{
			log.Fatal(err)
		}
	}

	// test flag directive args 

	// an NGINX_CONF_FLAG directive
	statement2 := struct {
		directive string
		args []string
        line      int // this is arbitrary
    }{
		"accept_mutex",
		nil,
        2,
	}

	good_args := [6][1]string{
		[1]string{"on"},
		[1]string{"off"},
		[1]string{"On"},
		[1]string{"Off"},
		[1]string{"ON"},
		[1]string{"OFF"},
	}

	for _,v := range good_args{
		statement2.args = v 
		crossplane.analyzer.analyze(fname, statement2, ";", ctx)

	}
	bad_args := [5][1]string{
		[1]string{"1"},
		[1]string{"0"},
		[1]string{"true"},
		[1]string{"okay"},
		[1]string{""},
	}

	for _,v := range bad_args{
		statement2.args = v 
		err := crossplane.analyzer.analyze(fname, statement2, ";", ctx)
		if err != nil{
			log.Fatal(err)
		}
	}
	
}
