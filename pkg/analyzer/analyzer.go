package analyzer

import "errors"

type Analy struct {
	masks      map[string]int
	DIRECTIVES map[string]string
	CONTEXT    map[string][]string
}

func newAnaly() {
	a = new(Analy)
	a.masks = map[string]int{
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
}

func analyze(fname string, stmt statement, ctx []string, strict bool, check_ctx bool, check_arg bool) {
	directive := stmt.directive
	line := stmt.line
	dir := check_directive(directive)
	if strict && !dir {
		errors.New("unknown directive " + directive)
	}

	ct := check_context(ctx)

	if !ct && !dir {
		return
	}

	args := stmtm.args || []string{}
	n_args := len(args)

	masks = DIRECTIVES[directive]

	if check_ctx {
		masks := func() []string {
			for _, m := range masks {
				if m && CONTEXT[ctx]
			}
		}
	}

}
