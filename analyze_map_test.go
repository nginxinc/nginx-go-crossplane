/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package crossplane

import (
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:funlen,exhaustruct
func TestAnalyzeMapBody(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		mapDirective string
		parameter    *Directive
		term         string
		wantErr      *ParseError
	}{
		"valid map": {
			mapDirective: "map",
			parameter: &Directive{
				Directive: "default",
				Args:      []string{"0"},
				Line:      5,
				Block:     Directives{},
			},
			term: ";",
		},
		"valid map volatile parameter": {
			mapDirective: "map",
			parameter: &Directive{
				Directive: "volatile",
				Line:      5,
				Block:     Directives{},
			},
			term: ";",
		},
		"invalid map volatile parameter": {
			mapDirective: "map",
			parameter: &Directive{
				Directive: "volatile",
				Args:      []string{"1"},
				Line:      5,
				Block:     Directives{},
			},
			term:    ";",
			wantErr: &ParseError{What: "invalid number of parameters", BlockCtx: "map"},
		},
		"valid map hostnames parameter": {
			mapDirective: "map",
			parameter: &Directive{
				Directive: "hostnames",
				Line:      5,
				Block:     Directives{},
			},
			term: ";",
		},
		"invalid map hostnames parameter": {
			mapDirective: "map",
			parameter: &Directive{
				Directive: "hostnames",
				Args:      []string{"foo"},
				Line:      5,
				Block:     Directives{},
			},
			term:    ";",
			wantErr: &ParseError{What: "invalid number of parameters", BlockCtx: "map"},
		},
		"valid geo proxy_recursive parameter": {
			mapDirective: "geo",
			parameter: &Directive{
				Directive: "proxy_recursive",
				Line:      5,
				Block:     Directives{},
			},
			term: ";",
		},
		"valid types": {
			mapDirective: "types",
			parameter: &Directive{
				Directive: "text/html",
				Args:      []string{"html htm shtml"},
				Line:      5,
				Block:     Directives{},
			},
			term: ";",
		},
		"invalid types with special parameter": {
			mapDirective: "types",
			parameter: &Directive{
				Directive: "hostnames",
				Line:      5,
				Block:     Directives{},
			},
			term:    ";",
			wantErr: &ParseError{What: "invalid number of parameters", BlockCtx: "types"},
		},
		"invalid geo proxy_recursive parameter": {
			mapDirective: "geo",
			parameter: &Directive{
				Directive: "proxy_recursive",
				Args:      []string{"1"},
				Line:      5,
				Block:     Directives{},
			},
			term:    ";",
			wantErr: &ParseError{What: "invalid number of parameters", BlockCtx: "geo"},
		},
		"valid geo ranges parameter": {
			mapDirective: "geo",
			parameter: &Directive{
				Directive: "ranges",
				Line:      5,
				Block:     Directives{},
			},
			term: ";",
		},
		"invalid geo ranges parameter": {
			mapDirective: "geo",
			parameter: &Directive{
				Directive: "ranges",
				Args:      []string{"0", "0", "0"},
				Line:      5,
				Block:     Directives{},
			},
			term:    ";",
			wantErr: &ParseError{What: "invalid number of parameters", BlockCtx: "geo"},
		},
		"invalid number of parameters in map": {
			mapDirective: "map",
			parameter: &Directive{
				Directive: "default",
				Args:      []string{"0", "0"},
				Line:      5,
				Block:     Directives{},
			},
			term:    ";",
			wantErr: &ParseError{What: "invalid number of parameters", BlockCtx: "map"},
		},
		"valid split_clients": {
			mapDirective: "split_clients",
			parameter: &Directive{
				Directive: "0.5%",
				Args:      []string{"google.com"},
				Line:      5,
				Block:     Directives{},
			},
			term: ";",
		},
		"invalid split_clients": {
			mapDirective: "split_clients",
			parameter: &Directive{
				Directive: "0.5%",
				Args:      []string{"google.com", "testme"},
				Line:      5,
				Block:     Directives{},
			},
			term:    ";",
			wantErr: &ParseError{What: "invalid number of parameters", BlockCtx: "split_clients"},
		},
		"valid geoip2": {
			mapDirective: "geoip2",
			parameter: &Directive{
				Directive: "$geoip2_data_continent_code",
				Args:      []string{"continent", "code"},
				Line:      5,
				Block:     Directives{},
			},
			term: ";",
		},
		"invalid geoip2": {
			mapDirective: "geoip2",
			parameter: &Directive{
				Directive: "$geoip2_data_city_name",
				Args:      []string{},
				Line:      5,
				Block:     Directives{},
			},
			term:    ";",
			wantErr: &ParseError{What: "invalid number of parameters", BlockCtx: "geoip2"},
		},
		"valid geoip2 auto_reload": {
			mapDirective: "geoip2",
			parameter: &Directive{
				Directive: "auto_reload",
				Args:      []string{"5m"},
				Line:      5,
				Block:     Directives{},
			},
			term: ";",
		},
		"invalid geoip2 auto_reload": {
			mapDirective: "geoip2",
			parameter: &Directive{
				Directive: "auto_reload",
				Args:      []string{"5m", "10m"},
				Line:      5,
				Block:     Directives{},
			},
			term:    ";",
			wantErr: &ParseError{What: "invalid number of parameters", BlockCtx: "geoip2"},
		},
		"valid otel_exporter": {
			mapDirective: "otel_exporter",
			parameter: &Directive{
				Directive: "endpoint",
				Args:      []string{"localhost:4317"},
				Line:      5,
				Block:     Directives{},
			},
			term: ";",
		},
		"invalid otel_exporter": {
			mapDirective: "otel_exporter",
			parameter: &Directive{
				Directive: "endpoint",
				Args:      []string{"localhost:4317", "localhost:1234"},
				Line:      5,
				Block:     Directives{},
			},
			term:    ";",
			wantErr: &ParseError{What: "invalid number of parameters", BlockCtx: "otel_exporter"},
		},
		"missing semicolon": {
			mapDirective: "map",
			parameter: &Directive{
				Directive: "default",
				Args:      []string{"0", "0"},
				Line:      5,
				Block:     Directives{},
			},
			term:    "}",
			wantErr: &ParseError{What: `unexpected "}"`, BlockCtx: "map"},
		},
	}

	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := analyzeMapBody("nginx.conf", tc.parameter, tc.term, tc.mapDirective)
			if tc.wantErr == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)

			var perr *ParseError
			require.ErrorAs(t, err, &perr)
			require.Equal(t, tc.wantErr.What, perr.What)
			require.Equal(t, tc.wantErr.BlockCtx, perr.BlockCtx)
		})
	}
}
