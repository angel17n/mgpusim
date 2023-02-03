package main

import (
	"flag"

	"gitlab.com/akita/mgpusim/v3/benchmarks/memoryaccess/elementwise"
	"gitlab.com/akita/mgpusim/v3/samples/runner"
)

var numData = flag.Int("length", 64, "The length of the array.")

func main() {
	flag.Parse()

	runner := new(runner.Runner).ParseFlag().Init()

	benchmark := elementwise.NewBenchmark(runner.Driver())
	benchmark.Length = *numData

	runner.AddBenchmark(benchmark)

	runner.Run()
}
