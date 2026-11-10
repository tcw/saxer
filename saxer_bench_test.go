package main

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/tcw/saxer/internal/xmlgen"
)

// benchSize is the size of the generated benchmark document. Override it
// with SAXER_BENCH_MB, e.g. SAXER_BENCH_MB=1024 for a 1 GB document.
func benchSize() int64 {
	if mb, err := strconv.Atoi(os.Getenv("SAXER_BENCH_MB")); err == nil && mb > 0 {
		return int64(mb) << 20
	}
	return 64 << 20
}

var benchDoc = sync.OnceValue(func() []byte {
	var b bytes.Buffer
	if err := xmlgen.Write(&b, xmlgen.Options{Size: benchSize(), Seed: 1}); err != nil {
		panic(err)
	}
	return b.Bytes()
})

// BenchmarkSaxer runs complete queries over a generated document, the way
// the command line does, and discards the output.
func BenchmarkSaxer(b *testing.B) {
	doc := benchDoc()
	cases := []struct {
		name string
		opts options
	}{
		{"plain", options{query: "Fuel"}},
		{"attribute", options{query: "car?man=Saab"}},
		{"inner-xml", options{query: "color", isInnerXml: true}},
		{"count", options{query: "car", count: true}},
		{"no-match", options{query: "nomatch"}},
		{"large-matches", options{query: "car"}},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.SetBytes(int64(len(doc)))
			for b.Loop() {
				opts := c.opts
				opts.tagBuffer = 4
				opts.contentBuf = 4
				if _, err := SaxXmlInput(bytes.NewReader(doc), io.Discard, &opts); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
