# saxer

[![CI](https://github.com/tcw/saxer/actions/workflows/ci.yml/badge.svg)](https://github.com/tcw/saxer/actions/workflows/ci.yml)
[![Go version](https://img.shields.io/github/go-mod/go-version/tcw/saxer)](go.mod)
[![Go Reference](https://pkg.go.dev/badge/github.com/tcw/saxer.svg)](https://pkg.go.dev/github.com/tcw/saxer)

**grep for XML.** Query XML files that are too big for other tools: saxer streams the input, uses a few megabytes of memory whatever the file size, and gets through 650–900 MB a second.

## Installation

### Download a release binary

Every tagged version has ready-built binaries on the [Releases page](https://github.com/tcw/saxer/releases), no Go needed. Pick the archive for your system:

| OS | Intel / AMD (amd64) | ARM (arm64) |
| --- | --- | --- |
| Linux | `saxer_<version>_linux_amd64.tar.gz` | `saxer_<version>_linux_arm64.tar.gz` |
| macOS | `saxer_<version>_darwin_amd64.tar.gz` | `saxer_<version>_darwin_arm64.tar.gz` (Apple silicon) |
| Windows | `saxer_<version>_windows_amd64.zip` | `saxer_<version>_windows_arm64.zip` |

Extract it and put the `saxer` binary somewhere on your `PATH`, for example on Linux or macOS:

    tar -xzf saxer_<version>_linux_amd64.tar.gz
    sudo mv saxer /usr/local/bin/
    saxer --version

On Windows, unzip the archive and move `saxer.exe` to a folder listed in your `Path` environment variable.

To verify the download, fetch `checksums.txt` from the same release and check the archive against it:

    sha256sum --ignore-missing -c checksums.txt              # Linux
    shasum -a 256 --ignore-missing -c checksums.txt          # macOS

The binaries are not signed. On macOS, if Gatekeeper refuses to run saxer, remove the quarantine flag:

    xattr -d com.apple.quarantine saxer

### Install from source

    go install github.com/tcw/saxer@latest

This needs Go 1.27 or later and puts `saxer` in `$(go env GOPATH)/bin`.

## Quick start

Try it on the bundled [testdata/example.xml](testdata/example.xml), a file of three cars:

    $ saxer Fuel testdata/example.xml
    <Fuel>Gasoline</Fuel>
    <Fuel>Diesel</Fuel>
    <Fuel>Diesel</Fuel>

    $ saxer "car?man=Saab" testdata/example.xml
    <car vin="wp033" man="Saab">
    		<color>yellow</color>
    		<xs:doors>4</xs:doors>
    		<engine nr="003">
    			<Fuel>Diesel</Fuel>
    		</engine>
    	</car>

    $ saxer -n "Fuel~Diesel" testdata/example.xml
    2

Every match is printed as the original XML, so the output can go straight to another tool. Without a file, saxer reads stdin. Gzip-compressed input is detected and decompressed automatically, so `saxer car dump.xml.gz` just works.

## Common tasks

Find elements by name, or by a path of names. A query matches anywhere in the document, and `*` matches any single element:

    $ saxer engine/Fuel testdata/example.xml
    <Fuel>Gasoline</Fuel>
    <Fuel>Diesel</Fuel>
    <Fuel>Diesel</Fuel>

    $ saxer -i "car/*/Fuel" testdata/example.xml
    Gasoline
    Diesel
    Diesel

Filter by attribute, or by text:

    $ saxer "?nr=002" testdata/example.xml
    <engine nr="002">
    			<Fuel>Diesel</Fuel>
    		</engine>

    $ saxer "color~red" testdata/example.xml
    <color>red</color>

Count matches:

    $ saxer -n car testdata/example.xml
    3

Get line numbers and paths, tab separated for `cut` and `awk`:

    $ saxer -m engine testdata/example.xml
    5-7	cars/car/engine
    12-14	cars/car/engine
    19-21	cars/car/engine

One node per line, for `grep`, `sort` and friends:

    $ saxer --squash engine testdata/example.xml
    <engine nr="001"> <Fuel>Gasoline</Fuel> </engine>
    <engine nr="002"> <Fuel>Diesel</Fuel> </engine>
    <engine nr="003"> <Fuel>Diesel</Fuel> </engine>

JSON lines, one object per match, for `jq` and scripts:

    $ saxer -j color testdata/example.xml
    {"content":"<color>blue</color>","line_start":3,"line_end":3,"path":"cars/car/color","attributes":{}}
    {"content":"<color>red</color>","line_start":10,"line_end":10,"path":"cars/car/color","attributes":{}}
    {"content":"<color>yellow</color>","line_start":17,"line_end":17,"path":"cars/car/color","attributes":{}}

For example, list the VIN of every Volvo:

    saxer -j "car?man=Volvo" big.xml | jq -r .attributes.vin

## Query syntax

    step/step/...          each step is   name[~text][?key[=value][&key[=value]]...]

| Query | Matches |
| --- | --- |
| `car` | every `car` element |
| `engine/Fuel` | `Fuel` elements whose parent is an `engine` |
| `car/*/Fuel` | `Fuel` elements whose grandparent is a `car` |
| `*` | every element (the outermost ones; matches never nest) |
| `?nr` | any element with an `nr` attribute |
| `engine?nr=003` | `engine` elements with `nr="003"` |
| `car?vin=wp031&man=Volvo` | `car` elements with both attributes |
| `color~red` | `color` elements whose text is `red` |
| `?name=Owner \& Co` | a literal `&`; a backslash escapes any of `/ ? & = ~ * \` |

- A query matches at any depth: `engine/Fuel` finds that pair wherever it is.
- Attribute values and text are compared after decoding entities, so `?name=a\&b` matches `name="a&amp;b"`.
- The text of an element is all the text inside it, that of child elements and CDATA included, with white space trimmed at both ends. So `car~red` only matches a car with no other text, while `-c "car~red"` matches a car with red anywhere in it. Text can only be matched on the last step.
- `-s` makes names, attributes and text case insensitive, `-c` matches them by substring, and `-o` ignores namespace prefixes on names.
- When matches are nested, only the outermost is printed. With text matching, an inner element is printed when the outer one does not match.
- A query can have up to 64 steps; documents can nest to any depth and have any number of attributes.
- Remember to quote queries with `?`, `&`, `*` or `~` in the shell.

## Flags

    saxer [flags] <query> [file]

Matching:

    -s, --case          Turn on case insensitivity
    -o, --omit-ns       Omit namespace in tag-name matches
    -c, --contains      Match tag names, attributes and text by contains (not equals)
    -f, --firstN int    First n matches (0 = all matches)

Output:

    -i, --inner         Inner-xml of selected element
    -n, --count         Number of matches
    -m, --meta          Line numbers and path of matches, tab separated
    -j, --json          One JSON object per match and line: content, line_start, line_end, path, attributes
    -u, --unescape      Unescape xml escape tokens (&lt; &gt; &#38; ...)
    -l, --single-line   Each node will have a single line (line endings become spaces)
        --squash        Squash runs of white space into one space (implies --single-line)
    -w, --wrap          Wrap result in Xml tag

Buffers and diagnostics:

        --tag-buf int   Size of element tag buffer in KB - largest tag (default 4)
        --cont-buf int  Size of content buffer in MB - largest returned element (default 4)
        --profile-cpu   Write a CPU profile to cpu.pprof
    -h, --help          help for saxer
    -v, --version       version for saxer

## Performance

saxer reads 650–900 MB of XML a second on a single core, in about 14 MB of memory whatever the file size. Gzip input is limited by decompression, about 220 MB a second.

| Query (1 GiB generated file) | Time | Throughput (MiB/s) |
| --- | --- | --- |
| `saxer Fuel`, many small matches | 1.19 s | 860 |
| `saxer "car?man=Saab"`, attribute match | 1.56 s | 656 |
| `saxer -i color`, inner xml | 1.24 s | 826 |
| `saxer -n car`, count | 1.45 s | 706 |
| `saxer car`, matches that make up most of the file | 1.51 s | 678 |
| `saxer nomatch`, a scan with no matches | 1.16 s | 883 |
| `saxer -j engine`, JSON lines | 2.00 s | 512 |

Measured on one core of an Intel Xeon @ 2.30GHz cloud VM, with the file in the page cache; peak memory was 14 MB. The same version before these optimisations took 5.7–7.1 s for the same queries. The file was written by the bundled generator, `go run ./cmd/saxer-gen -size 1G -o cars.xml`: car records with long text, many attributes, comments, CDATA, processing instructions, a DOCTYPE and deep nesting. Before and after numbers for each optimisation are in [docs/performance.md](docs/performance.md), with profiles.

## Troubleshooting

**`tag is larger than the tag buffer`**: a single tag, with its attributes, is larger than 4 KB. Raise the limit with `--tag-buf`, in KB, e.g. `--tag-buf 64`.

**`content buffer is full`**: a matched element is larger than 4 MB, and saxer holds each match in memory before printing it. Raise the limit with `--cont-buf`, in MB, or query a smaller element. Counting (`-n`) and line numbers (`-m`) do not hold matches in memory, so they have no such limit.

**`end tag </x> does not match start tag <y>`** and similar errors mean the input is not well formed XML; the error names the line.

**Exit codes**, as for grep:

| Code | Meaning |
| --- | --- |
| 0 | something matched |
| 1 | nothing matched |
| 2 | an error, such as a missing file, a bad query or malformed XML |

## Example file

The examples above and below use [testdata/example.xml](testdata/example.xml):

    <cars>
    	<car vin="wp031" man="Volvo">
    		<color>blue</color>
    		<xs:doors>4</xs:doors>
    		<engine nr="001">
    			<Fuel>Gasoline</Fuel>
    		</engine>
    	</car>
    	<car vin="wp032" man="Volvo">
    		<color>red</color>
    		<xs:doors>2</xs:doors>
    		<engine nr="002">
    			<Fuel>Diesel</Fuel>
    		</engine>
    	</car>
    	<car vin="wp033" man="Saab">
    		<color>yellow</color>
    		<xs:doors>4</xs:doors>
    		<engine nr="003">
    			<Fuel>Diesel</Fuel>
    		</engine>
    	</car>
    	<info>&lt;some-xml>data&lt;/some-xml></info>
    </cars>

## More examples

    $ saxer "?vin=wp031&man=Volvo" testdata/example.xml
    <car vin="wp031" man="Volvo">
    		<color>blue</color>
    		<xs:doors>4</xs:doors>
    		<engine nr="001">
    			<Fuel>Gasoline</Fuel>
    		</engine>
    	</car>

    $ saxer -l engine testdata/example.xml
    <engine nr="001"> 			<Fuel>Gasoline</Fuel> 		</engine>
    <engine nr="002"> 			<Fuel>Diesel</Fuel> 		</engine>
    <engine nr="003"> 			<Fuel>Diesel</Fuel> 		</engine>

    $ saxer -f 2 Fuel testdata/example.xml
    <Fuel>Gasoline</Fuel>
    <Fuel>Diesel</Fuel>

    $ saxer -u info testdata/example.xml
    <info><some-xml>data</some-xml></info>

    $ saxer -s fuel testdata/example.xml
    <Fuel>Gasoline</Fuel>
    <Fuel>Diesel</Fuel>
    <Fuel>Diesel</Fuel>

    $ saxer -o doors testdata/example.xml
    <xs:doors>4</xs:doors>
    <xs:doors>2</xs:doors>
    <xs:doors>4</xs:doors>

    $ saxer -c or testdata/example.xml
    <color>blue</color>
    <xs:doors>4</xs:doors>
    <color>red</color>
    <xs:doors>2</xs:doors>
    <color>yellow</color>
    <xs:doors>4</xs:doors>

    $ saxer -w Fuel testdata/example.xml
    <saxer-result>
    <Fuel>Gasoline</Fuel>
    <Fuel>Diesel</Fuel>
    <Fuel>Diesel</Fuel>
    </saxer-result>

    $ saxer -j -i "car?man=Saab" testdata/example.xml
    {"content":"\n\t\t<color>yellow</color>\n\t\t<xs:doors>4</xs:doors>\n\t\t<engine nr=\"003\">\n\t\t\t<Fuel>Diesel</Fuel>\n\t\t</engine>\n\t","line_start":16,"line_end":22,"path":"cars/car","attributes":{"vin":"wp033","man":"Saab"}}

## Contributing

Issues and pull requests are welcome. Before sending a change, run:

    go vet ./...
    go test -race ./...                                     # unit tests, including every example in this README
    go test -run '^$' -bench . -benchmem .                  # benchmarks over a generated 64 MB document
    SAXER_BENCH_MB=1024 go test -run '^$' -bench . .        # the same over 1 GB
    go test ./saxreader -run '^$' -fuzz FuzzRead -fuzztime 30s

Compare benchmark runs with [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat), and include before and after numbers in pull requests that aim to make saxer faster. A CPU profile of a real run is one flag away: `saxer --profile-cpu -n car cars.xml && go tool pprof -top saxer cpu.pprof`.

### Releasing

Pushing a tag that starts with `v` (for example `v1.2.0`, or `v1.2.0-rc.1` for a prerelease) runs the [release workflow](.github/workflows/release.yml): it runs the tests, then [GoReleaser](https://goreleaser.com) builds binaries for linux, darwin and windows on amd64 and arm64, and publishes them with `checksums.txt` and a changelog on a GitHub Release. Commits starting with `docs:`, `test:` or `chore:` are left out of the changelog. You can also release from the Actions tab (or the GitHub mobile app): run the Release workflow by hand and enter a version such as `v1.2.0` in the **tag** field; the workflow creates the tag on the chosen branch and publishes the release. Leave the field empty for a snapshot build that publishes nothing and attaches the archives to the run as a `saxer-snapshot` artifact; locally the same is:

    goreleaser check
    goreleaser release --snapshot --clean   # archives land in dist/

