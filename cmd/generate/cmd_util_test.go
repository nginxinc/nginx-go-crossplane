/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"testing"

	"github.com/nginxinc/nginx-go-crossplane/internal/generator"
	"github.com/stretchr/testify/require"
)

//nolint:funlen
func TestOverrideParser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected overrideItem
		wantErr  bool
	}{
		{
			name:  "normalFormat_pass",
			input: "location:ngxHTTPMainConf|ngxConfTake12,ngxStreamMainConf",

			expected: overrideItem{
				directive: "location",
				masks: []generator.Mask{
					{"ngxHTTPMainConf", "ngxConfTake12"},
					{"ngxStreamMainConf"},
				},
			},
			wantErr: false,
		},
		{
			name:  "withSpaces_pass",
			input: "hash:ngxHTTPUpsConf | ngxConfTake12, ngxStreamUpsConf | ngxConfTake12",
			expected: overrideItem{
				directive: "hash",
				masks: []generator.Mask{
					{"ngxHTTPUpsConf", "ngxConfTake12"},
					{"ngxStreamUpsConf", "ngxConfTake12"},
				},
			},
			wantErr: false,
		},
		{
			name:    "withoutColon_fail",
			input:   "hashngxHTTPUpsConf | ngxConfTake12,ngxStreamUpsConf | ngxConfTake12",
			wantErr: true,
		},
		{
			name:    "colonLeftsideEmpty_fail",
			input:   " :ngxHTTPUpsConf | ngxConfTake12,ngxStreamUpsConf | ngxConfTake12",
			wantErr: true,
		},
		{
			name:    "colonRightsideEmpty_fail",
			input:   "hash:  ",
			wantErr: true,
		},
		{
			name:    "emptyBitmask_fail",
			input:   "hash: ngxHTTPUpsConf| ",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got overrideItem
			err := got.UnmarshalText([]byte(tc.input))

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// If the testcase wants an error and there is an error, skip the output file validation.
			// Output makes no sense when there is an error.
			if err != nil {
				return
			}

			require.Equal(t, tc.expected, got)
		})
	}
}
