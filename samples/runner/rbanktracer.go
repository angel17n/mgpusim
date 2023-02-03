package runner

import (
	"gitlab.com/akita/akita/v3/tracing"
)

// rBankTracer can trace the number of instruction completed.
type rBankTracer struct {
	readScalarCount     uint64
	writeScalarCount    uint64
	readVectorialCount  uint64
	writeVectorialCount uint64
	accessType          string

	inflightInst map[string]tracing.Task
}

// newrBankTracer creates a tracer that can count the number of instructions.
func newRBankTracer() *rBankTracer {
	t := &rBankTracer{
		inflightInst: map[string]tracing.Task{},
	}
	return t
}

func (t *rBankTracer) StartTask(task tracing.Task) {
	if task.Kind != "registeraccess" {
		return
	}

	t.accessType = task.What

	t.inflightInst[task.ID] = task
}

func (t *rBankTracer) StepTask(task tracing.Task) {
	// Do nothing
}

func (t *rBankTracer) EndTask(task tracing.Task) {
	_, found := t.inflightInst[task.ID]
	if !found {
		return
	}

	switch t.accessType {
	case "readS":
		t.readScalarCount++
		break

	case "readV":
		t.readVectorialCount++
		break

	case "writeS":
		t.writeScalarCount++
		break

	case "writeV":
		t.writeVectorialCount++
		break
	}

	delete(t.inflightInst, task.ID)
}
