// Package sharedvectoradd implements the addition of a vector using local memory as a benchmark.
package elementwise

import (
	"log"

	// embed hsaco files
	_ "embed"

	"gitlab.com/akita/mgpusim/v3/driver"
	"gitlab.com/akita/mgpusim/v3/insts"
	"gitlab.com/akita/mgpusim/v3/kernels"
)

// KernelArgs defines kernel arguments
type KernelArgs struct {
	A                   driver.Ptr
	B                   driver.Ptr
	C                   driver.Ptr
	HiddenGlobalOffsetX int64
	HiddenGlobalOffsetY int64
	HiddenGlobalOffsetZ int64
}

// Benchmark defines a benchmark
type Benchmark struct {
	driver  *driver.Driver
	context *driver.Context
	gpus    []int
	hsaco   *insts.HsaCo

	Length      int
	A           []float32
	B           []float32
	C           []float32
	gA          driver.Ptr
	gB          driver.Ptr
	gC          driver.Ptr

	useUnifiedMemory bool
}

//go:embed kernels.hsaco
var hsacoBytes []byte

// NewBenchmark returns a benchmark
func NewBenchmark(driver *driver.Driver) *Benchmark {
	b := new(Benchmark)

	b.driver = driver
	b.context = driver.Init()
	b.hsaco = kernels.LoadProgramFromMemory(hsacoBytes, "elementwise")

	return b
}

// SelectGPU selects GPU
func (b *Benchmark) SelectGPU(gpus []int) {
	b.gpus = gpus
}

// SetUnifiedMemory uses Unified Memory
func (b *Benchmark) SetUnifiedMemory() {
	b.useUnifiedMemory = true
}

// Run runs
func (b *Benchmark) Run() {
	b.driver.SelectGPU(b.context, b.gpus[0])
	b.initMem()
	b.exec()
}

func (b *Benchmark) initMem() {
	if b.useUnifiedMemory {
		b.gA = b.driver.AllocateUnifiedMemory(b.context, uint64(b.Length*4))
		b.gB = b.driver.AllocateUnifiedMemory(b.context, uint64(b.Length*4))
		b.gC = b.driver.AllocateUnifiedMemory(b.context, uint64(b.Length*4))
	} else {
		b.gA = b.driver.AllocateMemory(b.context, uint64(b.Length*4))
		b.driver.Distribute(b.context, b.gA, uint64(b.Length*4), b.gpus)

		b.gB = b.driver.AllocateMemory(b.context, uint64(b.Length*4))
		b.driver.Distribute(b.context, b.gB, uint64(b.Length*4), b.gpus)

		b.gC = b.driver.AllocateMemory(b.context, uint64(b.Length*4))
		b.driver.Distribute(b.context, b.gC, uint64(b.Length*4), b.gpus)
	}

	b.A = make([]float32, b.Length)
	b.B = make([]float32, b.Length)
	b.C = make([]float32, b.Length)
	for i := 0; i < b.Length; i++ {
		b.A[i] = float32(i)
		b.B[i] = float32(i)
	}

	b.driver.MemCopyH2D(b.context, b.gA, b.A)
	b.driver.MemCopyH2D(b.context, b.gB, b.B)
}

func (b *Benchmark) exec() {
	queues := make([]*driver.CommandQueue, len(b.gpus))

	for i, gpu := range b.gpus {
		b.driver.SelectGPU(b.context, gpu)
		q := b.driver.CreateCommandQueue(b.context)
		queues[i] = q

		numWI := b.Length / len(b.gpus)

		kernArg := KernelArgs{
			b.gA, b.gB, b.gC,
			int64(numWI * i), 0, 0,
		}

		b.driver.EnqueueLaunchKernel(
			q,
			b.hsaco,
			[3]uint32{uint32(numWI), 1, 1}, // Global grid size (number of WorkItems)
			[3]uint16{512, 1, 1}, // Work-group size (64 workItems = 1 wavefront).
			&kernArg,
		)
	}

	for _, q := range queues {
		b.driver.DrainCommandQueue(q)
	}

	b.driver.MemCopyD2H(b.context, b.C, b.gC)
}

// Verify verifies
func (b *Benchmark) Verify() {
	for i := 0; i < b.Length; i++ {
		if b.A[i] * b.B[i] != b.C[i] {
			log.Panicf("mismatch at %d, input %fx%f, output %f", i,
				b.A[i], b.B[i], b.C[i])
		}
	}

	log.Printf("Passed!\n")
}
