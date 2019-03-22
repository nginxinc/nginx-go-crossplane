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

	
	
}
