package main

import (
	"flag"

	"gitlab.com/akita/mgpusim/v3/benchmarks/memoryaccess/memoryread_explicit"
	"gitlab.com/akita/mgpusim/v3/samples/runner"
)

var numData = flag.Int("length", 163840, "The length of the array.")

func main() {
	flag.Parse()

	runner := new(runner.Runner).ParseFlag().Init()

	benchmark := memoryread_explicit.NewBenchmark(runner.Driver())
	benchmark.Length = *numData

	runner.AddBenchmark(benchmark)

	runner.Run()
}
