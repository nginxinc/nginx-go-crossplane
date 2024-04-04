package crossplane

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanner(t *testing.T) {
	t.Parallel()

	for _, f := range lexFixtures {
		f := f
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()

			path := getTestConfigPath(f.name, "nginx.conf")
			file, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			s := NewScanner(file)

			i := 0
			for {
				got, err := s.Scan()
				if err == io.EOF {
					if i < len(f.tokens)-1 {
						t.Fatal("unexpected end of file")
					}
					return
				}

				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				want := f.tokens[i]
				require.Equal(t, want.value, got.Text)
				require.Equal(t, want.line, got.Line)
				i++
			}
		})
	}
}

func TestScanner_unhappy(t *testing.T) {
	t.Parallel()

	for name, c := range unhappyFixtures {
		c := c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := NewScanner(strings.NewReader(c))
			for {
				_, err := s.Scan()
				if err == io.EOF {
					t.Fatal("reached end of string")
				}

				if err != nil {
					t.Logf("got error: %v", err)

					if gotErr := s.Err(); !errors.Is(gotErr, err) {
						t.Fatalf("error do not match: have=%+v, want=%+v", gotErr, err)
					}

					if _, gotErr := s.Scan(); !errors.Is(gotErr, err) {
						t.Fatalf("error after scan does not match: have=%+v, want=%+v", gotErr, err)
					}

					break
				}
			}
		})
	}
}

var t Token //nolint: gochecknoglobals // trying to avoid return value being optimzed away

func BenchmarkScan(b *testing.B) {
	for _, bm := range lexFixtures {
		b.Run(bm.name, func(b *testing.B) {
			path := getTestConfigPath(bm.name, "nginx.conf")
			file, err := os.Open(path)
			if err != nil {
				b.Fatal(err)
			}
			defer file.Close()

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if _, err := file.Seek(0, 0); err != nil {
					b.Fatal(err)
				}

				s := NewScanner(file)

				for {
					tok, err := s.Scan()
					if err == io.EOF {
						break
					}
					if err != nil {
						b.Fatal(err)
					}
					t = tok
				}
			}
		})
	}
}
