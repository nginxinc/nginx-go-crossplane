package analyzer

import (
	"errors"
	"strings"
)

type Analy struct {
	MASKS      map[string]int
	DIRECTIVES map[string]string
	CONTEXT    map[string][]string
	term string 	
}

func newAnaly() *Analy{
	a := new(Analy)
	a.term = ";"
	a.MASKS = map[string]int{
		"NGX_DIRECT_CONF":      0x00010000,
		"NGX_MAIN_CONF":        0x00040000,
		"NGX_EVENT_CONF":       0x00080000, // events
		"NGX_MAIL_MAIN_CONF":   0x00100000, // mail
		"NGX_MAIL_SRV_CONF":    0x00200000, // mail > server
		"NGX_STREAM_MAIN_CONF": 0x00400000, // stream
		"NGX_STREAM_SRV_CONF":  0x00800000, // stream > server
		"NGX_STREAM_UPS_CONF":  0x01000000, // stream > upstream
		"NGX_HTTP_MAIN_CONF":   0x02000000, // http
		"NGX_HTTP_SRV_CONF":    0x04000000, // http > server
		"NGX_HTTP_LOC_CONF":    0x08000000, // http > location
		"NGX_HTTP_UPS_CONF":    0x10000000, // http > upstream
		"NGX_HTTP_SIF_CONF":    0x20000000, // http > server > if
		"NGX_HTTP_LIF_CONF":    0x40000000, // http > location > if
		"NGX_HTTP_LMT_CONF":    0x80000000,
	}
	a.CONTEXT = map[string][]string{
		"NGX_MAIN_CONF":        []string{},
		"NGX_EVENT_CONF":       []string{"events"},
		"NGX_MAIL_MAIN_CONF":   []string{"mail"},
		"NGX_MAIL_SRV_CONF":    []string{"mail", "server"},
		"NGX_STREAM_MAIN_CONF": []string{"stream"},
		"NGX_STREAM_SRV_CONF":  []string{"stream", "server"},
		"NGX_STREAM_UPS_CONF":  []string{"stream", "upstream"},
		"NGX_HTTP_MAIN_CONF":   []string{"http"},
		"NGX_HTTP_SRV_CONF":    []string{"http", "server"},
		"NGX_HTTP_LOC_CONF":    []string{"http", "location"},
		"NGX_HTTP_UPS_CONF":    []string{"http", "upstream"},
		"NGX_HTTP_SIF_CONF":    []string{"http", "server", "if"},
		"NGX_HTTP_LIF_CONF":    []string{"http", "location", "if"},
		"NGX_HTTP_LMT_CONF":    []string{"http", "location", "limit_except"},
	}

	return a 
}

func analyze(fname string, stmt statement,term string ctx []string, strict bool, check_ctx bool, check_arg bool) {
	directive := stmt.directive
	a := newAnaly()
	a.term = term 
	line := stmt.line
	dir := checkDirective(directive)
	if strict && !dir {
		errors.New("unknown directive " + directive)
	}

	ct := checkContext(ctx)

	if !ct && !dir {
		return
	}
	if len(stmt.args) != 0{
		args := stmt.args 
	} else {
		args := [1]string{}
	}
	
	numArgs := len(args)

	masks := a.DIRECTIVES[directive]

	if check_ctx {
		masks := func() []string {
			b := []string{}
			for m,b  := range masks {
				if m & CONTEXT[ctx] != 0x00000000 {
					b.append(m)
				}
			}
		}	return b 
		if len(masks) == 0{
			errors.New(directive + " directive is not allowed here")
		}
	}

	if !check_arg{
		return 
	}

	validFlags := func(x string) bool{
		x = strings.ToLower(x)
		for _,v := range [2]string{"on", "off"}{
			if x == v{
				return true 
			}
		}
		return false 
	}

	reason := ""
	for i := len(masks); i >= 0; i--{
		if masks[i] & a.MASK["NGX_CONF_BLOCK"] == 0x00000000 && a.term != "{"{
			reason = "directive " + directive + " has no opening '{'"
			continue 
		}

		if masks[i] & a.MASKS["NGX_CONF_BLOCK"] != 0x00000000 && a.term != ";"{
			reason = "directive " + directive + " is not terminated by ';'"
			continue 
		}

		if (masks[i] >> numArgs & 1 != 0x00000000 && numArgs <= 7) || 
		(masks[i] & a.MASKS["NGX_CONF_FLAG"] != 0x00000000 && numArgs ==1 && validFlags(args[0])) ||
		(masks[i] & a.MASKS["NGX_CONF_ANY"] != 0x00000000 && numArgs >= 0) ||
		(masks[i] & a.MASKS["NGX_CONF_1MORE"] != 0x00000000 && numArgs >= 1) ||
		(masks[i] & a.MASKS["NGX_CONF_2MORE"] != 0x00000000 && numArgs >= 2){
			return
		} else if masks[i] & a.MASKS["NGX_CONF_FLAG"] != 0x00000000 && numArgs == 1 && !validFlags(args[0]){
			reason = "invalid value "+ args[0] +" in "+ directive + " directive, it must be 'on' or 'off'"
			continue 
		} else {
			reason = "invalid number of arguements in "+ directive
			continue 
		}

	}
	return errors.New(reason)

}

func checkContext(cont []string, contexts map[string][]string){
	for _,c := cont {
		for k,v := range contexts{
			if c == v {
				return true 
			}
		}
	}
	return false 
}

func checkDirective(dir string, direct map[string]string){
	for d,v := direct {
		if d == dir {
			return true 
		}
	}
	return false 
}
