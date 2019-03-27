package analyzer

import (
	"errors"
	"strings"
)

type Analy struct {
	MASKS      map[string]int
	DIRECTIVES map[string][]string
	CONTEXT    map[[3]string]string
	term       string
}

func newAnaly() *Analy {
	a := new(Analy)
	a.term = ";"
	a.MASKS = map[string]int{
		"NGX_DIRECT_CONF":      0x00010000, // main file (not used)
		"NGX_MAIN_CONF":        0x00040000, // main context
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
	a.CONTEXT = map[[3]string]string{
		[3]string{}:                                   "NGX_MAIN_CONF",
		[3]string{"events"}:                           "NGX_EVENT_CONF",
		[3]string{"mail"}:                             "NGX_MAIL_MAIN_CONF",
		[3]string{"mail", "server"}:                   "NGX_MAIL_SRV_CONF",
		[3]string{"stream"}:                           "NGX_STREAM_MAIN_CONF",
		[3]string{"stream", "server"}:                 "NGX_STREAM_SRV_CONF",
		[3]string{"stream", "upstream"}:               "NGX_STREAM_UPS_CONF",
		[3]string{"http"}:                             "NGX_HTTP_MAIN_CONF",
		[3]string{"http", "server"}:                   "NGX_HTTP_SRV_CONF",
		[3]string{"http", "location"}:                 "NGX_HTTP_LOC_CONF",
		[3]string{"http", "upstream"}:                 "NGX_HTTP_UPS_CONF",
		[3]string{"http", "server", "if"}:             "NGX_HTTP_SIF_CONF",
		[3]string{"http", "location", "if"}:           "NGX_HTTP_LIF_CONF",
		[3]string{"http", "location", "limit_except"}: "NGX_HTTP_LMT_CONF",
	}

	return a
}

func analyze(fname string, stmt statement, term string, ctx [3]string, strict bool, checkCtx bool, checkArg bool) error {
	directive := stmt.directive
	a := newAnaly()
	a.term = term
	//line := stmt.line
	dir := checkDirective(directive, a.DIRECTIVES)

	// if strict and directive isn't recognized then throw error
	if strict && !dir {
		errors.New("unknown directive " + directive)
	}

	ct := checkContext(ctx, a.CONTEXT)

	// if we don't know where this directive is allowed and how
	// many arguments it can take then don't bother analyzing it
	if !ct && !dir {
		return errors.New("problem here")
	}

	args := stmt.args

	//  makes numArgs an unsigned int for bit shifting later
	numArgs := uint(len(args))

	masks := a.DIRECTIVES[directive]

	// if this directive can't be used in this context then throw an error
	if checkCtx {
		for _, mask := range masks {
			bitmask := a.CONTEXT[ctx]
			if a.MASKS[mask]&a.MASKS[bitmask] != 0x00000000 {
				masks = append(masks, mask)
			}
		}
		if len(masks) == 0 {
			errors.New(directive + " directive is not allowed here")
		}
	}

	if !checkArg {
		return errors.New("Big yikes")
	}

	validFlags := func(x string) bool {
		x = strings.ToLower(x)
		for _, v := range [2]string{"on", "off"} {
			if x == v {
				return true
			}
		}
		return false
	}
	// do this in reverse because we only throw errors at the end if no masks
	// are valid, and typically the first bit mask is what the parser expects
	reason := ""
	for i := len(masks); i >= 0; i-- {
		msk := masks[i]
		// if the directive isn't a block but should be according to the mask
		if a.MASKS[msk]&a.MASKS["NGX_CONF_BLOCK"] != 0x00000000 && a.term != "{" {
			reason = "directive " + directive + " has no opening '{'"
			continue
		}
		//if the directive is a block but shouldn't be according to the mask
		if a.MASKS[msk]&a.MASKS["NGX_CONF_BLOCK"] != 0x00000000 && a.term != ";" {
			reason = "directive " + directive + " is not terminated by ';'"
			continue
		}
		// use mask to check the directive's arguments
		if ((a.MASKS[msk]>>numArgs)&1 != 0x00000000 && numArgs <= 7) || //NOARGS to TAKE7
			(a.MASKS[msk]&a.MASKS["NGX_CONF_FLAG"] != 0x00000000 && numArgs == 1 && validFlags(stmt.args[0])) ||
			(a.MASKS[msk]&a.MASKS["NGX_CONF_ANY"] != 0x00000000 && numArgs >= 0) ||
			(a.MASKS[msk]&a.MASKS["NGX_CONF_1MORE"] != 0x00000000 && numArgs >= 1) ||
			(a.MASKS[msk]&a.MASKS["NGX_CONF_2MORE"] != 0x00000000 && numArgs >= 2) {
			return errors.New("Big Problem")
		} else if a.MASKS[msk]&a.MASKS["NGX_CONF_FLAG"] != 0x00000000 && numArgs == 1 && !validFlags(stmt.args[0]) {
			reason = "invalid value " + stmt.args[0] + " in " + stmt.directive + " directive, it must be 'on' or 'off'"
			continue
		} else {
			reason = "invalid number of arguements in " + directive
			continue
		}
	}
	return errors.New(reason)
}

func checkContext(cont [3]string, contexts map[[3]string]string) bool {
	if contexts[cont] != "" {
		return true
	}
	return false
}

func checkDirective(dir string, direct map[string][]string) bool {
	for d := range direct {
		if d == dir {
			return true
		}
	}
	return false
}
