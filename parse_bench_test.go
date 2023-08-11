/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package crossplane

import (
	"bytes"
	"compress/bzip2"
	"flag"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals
var (
	runBenchLocally = flag.Bool("local-parse-bench", false, "perform local parse benchmark test")

	once       sync.Once
	cfgPath    string
	removeFunc func()
)

func getLargeConfig(b *testing.B) (string, func()) {
	if !*runBenchLocally {
		b.Skip("getLargeConfig is only run locally when -local-parse-bench is specified")
	}

	path := getTestConfigPath("large-config", "nginx.conf.bz2")

	// unpack compressed log file and place in a temporary directory
	file, err := os.Open(path)
	if err != nil {
		b.Skip("cannot open input file")
	}
	defer file.Close()

	// Open output file
	tmpdir, err := os.MkdirTemp("", "alog")
	if err != nil {
		b.Skip("cannot create output dir")
	}
	remove := func() {
		b.Logf("removing temporary dir %s", tmpdir)
		_ = os.RemoveAll(tmpdir)
	}

	tmpFile := filepath.Join(tmpdir, "nginx.conf")
	truncFile, err := os.Create(tmpFile)
	if err != nil {
		b.Skip("cannot create output file")
	}

	bz2r := bzip2.NewReader(file)

	_, err = io.Copy(truncFile, bz2r)
	if err != nil {
		b.Skip("cannot copy to output file")
	}

	_ = truncFile.Close()

	b.Logf("Opened large config file %s", tmpFile)
	return tmpFile, remove
}

func getLargeConfigOnce(b *testing.B) string {
	once.Do(func() {
		cfgPath, removeFunc = getLargeConfig(b)
	})
	return cfgPath
}

func benchmarkParseLargeConfig(b *testing.B, size int) {
	defer func() { SetTokenChanCap(TokenChanCap) }()

	path := getLargeConfigOnce(b)

	b.ReportAllocs()
	b.ResetTimer()
	b.StopTimer()
	b.StartTimer()

	SetTokenChanCap(size)
	for i := 0; i < b.N; i++ {
		_, _ = Parse(path, &ParseOptions{SingleFile: true, StopParsingOnError: true})
	}
}

func TestMain(b *testing.M) {
	b.Run()
	if removeFunc != nil {
		removeFunc()
	}
	os.Exit(0)
}

func BenchmarkParseLargeConfig_Slow_TokBuf_0(b *testing.B)   { benchmarkParseLargeConfig(b, 0) }
func BenchmarkParseLargeConfig_Slow_TokBuf_1(b *testing.B)   { benchmarkParseLargeConfig(b, 1) }
func BenchmarkParseLargeConfig_Slow_TokBuf_8(b *testing.B)   { benchmarkParseLargeConfig(b, 8) }
func BenchmarkParseLargeConfig_Slow_TokBuf_64(b *testing.B)  { benchmarkParseLargeConfig(b, 64) }
func BenchmarkParseLargeConfig_Slow_TokBuf_512(b *testing.B) { benchmarkParseLargeConfig(b, 512) }
func BenchmarkParseLargeConfig_Slow_TokBuf_1024(b *testing.B) {
	benchmarkParseLargeConfig(b, 1024)
}
func BenchmarkParseLargeConfig_Slow_TokBuf_2048(b *testing.B) {
	benchmarkParseLargeConfig(b, 2048)
}
func BenchmarkParseLargeConfig_Slow_TokBuf_4096(b *testing.B) {
	benchmarkParseLargeConfig(b, 4096)
}

func BenchmarkParseLargeConfig_TokBuf_0(b *testing.B)    { benchmarkParseLargeConfig(b, 0) }
func BenchmarkParseLargeConfig_TokBuf_1(b *testing.B)    { benchmarkParseLargeConfig(b, 1) }
func BenchmarkParseLargeConfig_TokBuf_8(b *testing.B)    { benchmarkParseLargeConfig(b, 8) }
func BenchmarkParseLargeConfig_TokBuf_64(b *testing.B)   { benchmarkParseLargeConfig(b, 64) }
func BenchmarkParseLargeConfig_TokBuf_512(b *testing.B)  { benchmarkParseLargeConfig(b, 512) }
func BenchmarkParseLargeConfig_TokBuf_1024(b *testing.B) { benchmarkParseLargeConfig(b, 1024) }
func BenchmarkParseLargeConfig_TokBuf_2048(b *testing.B) { benchmarkParseLargeConfig(b, 2048) }
func BenchmarkParseLargeConfig_TokBuf_4096(b *testing.B) { benchmarkParseLargeConfig(b, 4096) }

func benchmarkParseBuildLargeConfig(b *testing.B, inclParse bool, size int,
	build func(w io.Writer, config Config, options *BuildOptions) error) {
	defer func() { SetTokenChanCap(TokenChanCap) }()
	path := getLargeConfigOnce(b)

	payload, err := Parse(path, &ParseOptions{SingleFile: true, StopParsingOnError: true})
	require.NoError(b, err)

	b.ReportAllocs()
	b.StopTimer()
	b.ResetTimer()
	b.StartTimer()
	SetTokenChanCap(size)
	for i := 0; i < b.N; i++ {
		if inclParse {
			payload, err = Parse(path, &ParseOptions{SingleFile: true, StopParsingOnError: true})
			require.NoError(b, err)
		}
		b := &bytes.Buffer{}
		bo := &BuildOptions{Tabs: true}

		_ = build(b, payload.Config[0], bo)
	}
}

func BenchmarkBuildLargeConfigInPlace(b *testing.B) {
	benchmarkParseBuildLargeConfig(b, false, TokenChanCap, Build)
}

func BenchmarkParseAndBuildLargeConfigInPlace(b *testing.B) {
	benchmarkParseBuildLargeConfig(b, true, TokenChanCap, Build)
}
