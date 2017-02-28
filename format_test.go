// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

package bindata

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"golang.org/x/tools/imports"
)

var testCases = [...]struct {
	name   string
	config func(*Config)
}{
	{"default", func(*Config) {}},
	{"old-default", func(c *Config) {
		c.Package = "main"
		c.MemCopy = true
		c.Compress = true
		c.Metadata = true
		c.HashLength = 16
		// The AssetDir API currently produces
		// wrongly formatted code. We're going
		// to skip it for now.
		/*c.AssetDir = true
		c.Restore = true*/
		c.DecompressOnce = true
	}},
	{"debug", func(c *Config) { c.Debug = true }},
	{"dev", func(c *Config) { c.Dev = true }},
	{"tags", func(c *Config) { c.Tags = "!x" }},
	{"package", func(c *Config) { c.Package = "test" }},
	{"prefix", func(c *Config) { c.Prefix = "testdata" }},
	{"compress", func(c *Config) { c.Compress = true }},
	{"copy", func(c *Config) { c.MemCopy = true }},
	{"metadata", func(c *Config) { c.Metadata = true }},
	{"decompress-once", func(c *Config) { c.DecompressOnce = true }},
	{"hash-dir", func(c *Config) { c.HashFormat = DirHash; c.HashLength = 16 }},
	{"hash-suffix", func(c *Config) { c.HashFormat = NameHashSuffix; c.HashLength = 16 }},
	{"hash-hashext", func(c *Config) { c.HashFormat = HashWithExt; c.HashLength = 16 }},
	{"hash-unchanged", func(c *Config) { c.HashFormat = NameUnchanged; c.HashLength = 16 }},
	{"hash-enc-b32", func(c *Config) { c.HashEncoding = Base32Hash; c.HashFormat = DirHash; c.HashLength = 16 }},
	{"hash-enc-b64", func(c *Config) { c.HashEncoding = Base64Hash; c.HashFormat = DirHash; c.HashLength = 16 }},
	{"hash-key", func(c *Config) { c.HashKey = []byte{0x00, 0x11, 0x22, 0x33}; c.HashFormat = DirHash; c.HashLength = 16 }},
}

var testPaths = [...]struct {
	path      string
	recursive bool
}{
	{"testdata", true},
	{"LICENSE", false},
	{"README.md", false},
}

var gencode = flag.String("gencode", "", "write generated code to specified directory")

func testGenerate(w io.Writer, c *Config) error {
	g, err := New(c)
	if err != nil {
		return err
	}

	for _, path := range testPaths {
		if err = g.FindFiles(path.path, path.recursive); err != nil {
			return err
		}
	}

	_, err = g.WriteTo(w)
	return err
}

func TestGenerate(t *testing.T) {
	if *gencode == "" {
		t.Skip("skipping test as -gencode flag not provided")
	}

	if err := os.Mkdir(*gencode, 0777); err != nil && !os.IsExist(err) {
		t.Fatal(err)
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			c := &Config{Package: "main"}
			test.config(c)

			f, err := os.Create(filepath.Join(*gencode, test.name+".go"))
			if err != nil {
				t.Fatal(err)
			}

			err = testGenerate(f, c)
			f.Close()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestFormatting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			c := &Config{Package: "main"}
			test.config(c)

			var buf bytes.Buffer
			if err := testGenerate(&buf, c); err != nil {
				t.Fatal(err)
			}

			out, err := imports.Process("bindata.go", buf.Bytes(), nil)
			if err != nil {
				t.Fatal(err)
			}

			if bytes.Equal(buf.Bytes(), out) {
				return
			}

			t.Error("not correctly formatted.")

			var diff bytes.Buffer
			diff.WriteString("diff:\n")

			if err := difflib.WriteUnifiedDiff(&diff, difflib.UnifiedDiff{
				A:       difflib.SplitLines(buf.String()),
				B:       difflib.SplitLines(string(out)),
				Context: 2,
				Eol:     "",
			}); err != nil {
				t.Fatal(err)
			}

			t.Log(diff.String())
		})
	}
}

func BenchmarkFindFiles(b *testing.B) {
	for _, test := range testCases {
		test := test
		b.Run(test.name, func(b *testing.B) {
			c := &Config{Package: "main"}
			test.config(c)

			g, err := New(c)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()

			for n := 0; n < b.N; n++ {
				if err = g.FindFiles("testdata", true); err != nil {
					b.Fatal(err)
				}

				g.toc = nil
			}
		})
	}
}

func BenchmarkWriteTo(b *testing.B) {
	for _, test := range testCases {
		test := test
		b.Run(test.name, func(b *testing.B) {
			c := &Config{Package: "main"}
			test.config(c)

			g, err := New(c)
			if err != nil {
				b.Fatal(err)
			}

			for _, path := range testPaths {
				if err = g.FindFiles(path.path, path.recursive); err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()

			for n := 0; n < b.N; n++ {
				if _, err = g.WriteTo(nopWriter{}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
