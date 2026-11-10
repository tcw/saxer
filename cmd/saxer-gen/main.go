// Command saxer-gen writes a large XML file for benchmarking saxer.
//
//	go run ./cmd/saxer-gen -size 1G -o /tmp/cars.xml
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/tcw/saxer/internal/xmlgen"
)

func main() {
	size := flag.String("size", "1G", "approximate size of the document, e.g. 500M, 1G, 4G")
	out := flag.String("o", "", "output file (default stdout)")
	depth := flag.Int("depth", 40, "nesting depth of the deep option chains")
	seed := flag.Uint64("seed", 1, "random seed")
	flag.Parse()

	n, err := parseSize(*size)
	if err != nil {
		fmt.Fprintln(os.Stderr, "saxer-gen:", err)
		os.Exit(2)
	}
	w := os.Stdout
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			fmt.Fprintln(os.Stderr, "saxer-gen:", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}
	if err := xmlgen.Write(w, xmlgen.Options{Size: n, Depth: *depth, Seed: *seed}); err != nil {
		fmt.Fprintln(os.Stderr, "saxer-gen:", err)
		os.Exit(1)
	}
}

// parseSize parses sizes such as 1024, 64K, 500M and 2G (powers of 1024).
func parseSize(s string) (int64, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	mult := int64(1)
	switch {
	case strings.HasSuffix(s, "K"):
		mult, s = 1<<10, s[:len(s)-1]
	case strings.HasSuffix(s, "M"):
		mult, s = 1<<20, s[:len(s)-1]
	case strings.HasSuffix(s, "G"):
		mult, s = 1<<30, s[:len(s)-1]
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid size %q", s)
	}
	return n * mult, nil
}
