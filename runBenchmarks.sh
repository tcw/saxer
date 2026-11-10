#!/bin/sh
# Benchmarks over a generated document; set SAXER_BENCH_MB=1024 for 1 GB.
go test -run '^$' -bench . -benchmem . ./saxreader
