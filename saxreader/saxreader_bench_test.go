package saxreader

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/tcw/saxer/contentbuffer"
	"github.com/tcw/saxer/tagmatcher"
)

func benchmarkDocument() []byte {
	var b bytes.Buffer
	b.WriteString("<?xml version=\"1.0\"?>\n<!-- generated -->\n<root>\n")
	for i := 0; i < 10000; i++ {
		fmt.Fprintf(&b, "  <item id=\"%d\" kind='x'>\n    <name>n%d</name>\n    <v><![CDATA[<raw>]]></v>\n  </item>\n", i, i)
	}
	b.WriteString("</root>\n")
	return b.Bytes()
}

func BenchmarkRead(b *testing.B) {
	doc := benchmarkDocument()
	for _, query := range []string{"name", "item?id=5000", "nomatch"} {
		b.Run(query, func(b *testing.B) {
			b.SetBytes(int64(len(doc)))
			for i := 0; i < b.N; i++ {
				sr := New()
				sr.EmitterFn = func(*contentbuffer.EmitterData) bool { return false }
				tm, err := tagmatcher.NewTagMatcher(query, tagmatcher.Options{})
				if err != nil {
					b.Fatal(err)
				}
				if err := sr.Read(bytes.NewReader(doc), &tm); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
