# Performance

saxer went from about 175 MiB/s to 690–880 MiB/s (4.4× to 5× faster), and from about a million allocations per 64 MiB document to under 60.

All numbers are from an Intel Xeon @ 2.30GHz cloud VM with 4 cores (saxer uses one) and Go 1.27.1. Benchmark times are medians of 6 runs compared with [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat). The VM is noisy (±5–10%), so changes under 5% are within noise unless benchstat marks them significant.

## How to measure

    go test -run '^$' -bench BenchmarkSaxer -benchmem -count 6 . > new.txt
    benchstat old.txt new.txt

`BenchmarkSaxer` runs whole queries, as the command line does, over a 64 MiB document from `internal/xmlgen` (set `SAXER_BENCH_MB=1024` for 1 GiB). The cases are:

| Case | Query | What it stresses |
| --- | --- | --- |
| plain | `Fuel` | many small matches |
| attribute | `car?man=Saab` | attribute parsing and matching |
| inner-xml | `-i color` | inner xml of small matches |
| count | `-n car` | counting large matches |
| no-match | `nomatch` | the bare scan |
| large-matches | `car` | copying and printing most of the file |

For a whole file, generate one with `go run ./cmd/saxer-gen -size 1G -o cars.xml` and profile a run with `saxer --profile-cpu -n car cars.xml && go tool pprof -top saxer cpu.pprof`.

## Where the time went before

A profile of `saxer Fuel` over 1 GiB (6.2 s) on the old reader:

| Function | Share of CPU | Why |
| --- | --- | --- |
| `parser.next` + `SaxReader.Read` | 37% | a function call and a state switch for every input byte |
| `tagmatcher.AddTag` | 22% | every tag converted to a string, then scanned rune by rune for attributes |
| `TagBuffer.Add` | 10% | every tag byte copied with a bounds check |
| `parser.tagByte` | 9% | per byte quote tracking in tags |
| `runtime.mallocgc` | 7% | 1.1 million allocations per 64 MiB: tag name strings, attribute strings, a string per match |

## Where it goes now

A profile of the same query (1.2 s):

| Function | Share of CPU | Notes |
| --- | --- | --- |
| `parser.scanTag` | 48% | finding the `>` that ends each tag; quoted values are skipped with `IndexByte` |
| `indexbytebody` | 25% | SIMD search for the next `<` and for closing quotes |
| `Syscall6` | 13% | reading the file: copying from the page cache |
| `markup.IsSpace` | 8% | finding the end of each tag name |

What is left is mostly a pass over each tag that cannot be avoided without parsing in parallel, which stays out of scope until the changes below have settled.

## Before and after

Old reader (commit 7590a54) against the new one, same benchmark and document:

| Case | Before | After | Change |
| --- | --- | --- | --- |
| plain | 365 ms (175 MiB/s) | 80 ms (799 MiB/s) | −78% |
| attribute | 368 ms (174 MiB/s) | 99 ms (645 MiB/s) | −73% |
| inner-xml | 367 ms (174 MiB/s) | 76 ms (844 MiB/s) | −79% |
| count | 392 ms (163 MiB/s) | 89 ms (715 MiB/s) | −77% |
| no-match | 333 ms (192 MiB/s) | 73 ms (880 MiB/s) | −78% |
| large-matches | 412 ms (155 MiB/s) | 93 ms (689 MiB/s) | −77% |

| Case | Allocations before | after | Bytes allocated before | after |
| --- | --- | --- | --- | --- |
| plain | 1,152,729 | 43 | 43.8 MiB | 132 KiB |
| count | 1,029,065 | 48 | 105.1 MiB | 134 KiB |
| large-matches | 1,130,182 | 51 | 106.7 MiB | 140 KiB |

The command line over a 1 GiB file, best of 3, output to /dev/null:

| Command | Before | After |
| --- | --- | --- |
| `saxer Fuel` | 6.12 s | 1.19 s |
| `saxer "car?man=Saab"` | 5.82 s | 1.56 s |
| `saxer -i color` | 6.04 s | 1.24 s |
| `saxer -n car` | 6.62 s | 1.45 s |
| `saxer car` | 7.10 s | 1.51 s |
| `saxer nomatch` | 5.65 s | 1.16 s |
| `saxer -j engine` (new) | | 2.00 s |
| `saxer -n car` on the file compressed with `gzip -1` | | 4.75 s |

Peak resident memory over all of these runs was 14 MB. Gzip input is limited by decompression in `compress/gzip`.

## Each change on its own

Each optimisation was measured by undoing only that one in the new reader and comparing (`old` = without the change, `new` = with it). The geomean is over the cases listed.

| Change | Cases | Without | With | Change |
| --- | --- | --- | --- | --- |
| Name first, attributes lazily: parse attributes only when a step's name matches | all | 102.9 ms | 83.6 ms | −18.8% (plain −30%, no-match −26%) |
| Bulk copy when recording: copy runs of bytes into the match at once | large-matches | 129.7 ms | 94.2 ms | −27.4% |
| No copy on emit: write matches from the buffer instead of a string | large-matches | 115.8 ms | 94.2 ms | −18.7% |
| Avoid per-tag strings: tag names as byte slices | all | 93.1 ms | 84.8 ms | −8.9% |
| Skip ahead with `bytes.IndexByte` outside a match | all | 90.5 ms | 85.9 ms | −5.2% |
| Count mode does not record content | count | 97.2 ms | 92.1 ms | −5.2% |
| Read buffer 64 KB instead of 4 KB, no `bufio` on stdin | all | 86.9 ms | 83.1 ms | −4.5% |

Notes:

- Skip-ahead pays in proportion to text length. On a first version of the generator, where text between tags was only indentation, the change was within noise; with the free-text `<description>` of the current generator it is 5%. Documents with long text gain more.
- Bulk copy and no-copy emit only matter when matches are a large part of the input: on the `plain` and `inner-xml` cases, which match small elements, they are within noise (−1.6% and −3.6% geomean).
- The read buffer saving comes from 16 times fewer read calls; with stdin the old code also copied every byte through an extra `bufio.Reader`.
- The largest gain, not separable into one switch, is the parser structure: the old reader handled every byte through a function call and a state switch; the new one works on a window of the input and only looks at bytes inside tags.
