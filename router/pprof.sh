#!/bin/sh
set -e

go test -bench . -run=none -cpuprofile cpu.out
go tool pprof -http=":8081" cpu.out
