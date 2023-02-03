package runner

import (
	"gitlab.com/akita/akita/v3/tracing"
)

const DEFAULT_VREG_SIZE = 16384

// rBankTracerPlus can trace the number of instruction completed.
type rBankTracerPlus struct {
	readVectorialCount  [DEFAULT_VREG_SIZE * 4]uint64
	writeVectorialCount [DEFAULT_VREG_SIZE * 4]uint64
	accessType          string

	inflightInst map[string]tracing.Task
}

// newrBankTracerPlus creates a tracer that can count the number of instructions.
func newRBankTracerPlus() *rBankTracerPlus {
	t := &rBankTracerPlus{
		inflightInst: map[string]tracing.Task{},
	}
	return t
}

func (t *rBankTracerPlus) StartTask(task tracing.Task) {
	if task.Kind != "registeraccess+" {
		return
	}

	t.accessType = task.What

	t.inflightInst[task.ID] = task
}

func (t *rBankTracerPlus) StepTask(task tracing.Task) {
	// Do nothing
}

func (t *rBankTracerPlus) EndTask(task tracing.Task) {
	ftask, found := t.inflightInst[task.ID]
	if !found {
		return
	}

	detail := ftask.Detail.(map[string]interface{})
	//reg := detail["reg"].(*insts.Reg)
	offset := detail["offset"].(int)

	switch t.accessType {
	case "readV":
		t.readVectorialCount[offset]++
		break

	case "writeV":
		t.writeVectorialCount[offset]++
		break
	}

	delete(t.inflightInst, task.ID)
}
