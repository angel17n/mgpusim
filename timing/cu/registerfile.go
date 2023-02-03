package cu

import (
	"log"

	"gitlab.com/akita/akita/v3/sim"
	"gitlab.com/akita/mgpusim/v3/insts"
	"gitlab.com/akita/mgpusim/v3/timing/wavefront"
)

const DEFAULT_VREG_SIZE = 16384

// A RegisterAccess is an incidence of reading or writing the register
type RegisterAccess struct {
	Time       sim.VTimeInSec
	SIMDID     int
	Inst       *wavefront.Inst
	Cu         *ComputeUnit
	RegType    string
	Reg        *insts.Reg
	RegCount   int
	LaneID     int
	WaveOffset int
	Data       []byte
	OK         bool
}

// A RegisterFile provides the communication interface for a set of registers.
type RegisterFile interface {
	Read(access RegisterAccess)
	Write(access RegisterAccess)
}

// A SimpleRegisterFile is a Register file that can always read and write
// registers immediately
type SimpleRegisterFile struct {
	storage []byte

	// In vector register, each lane can have up-to 256 VGPRs. Then the offset
	// difference from v0 lane 0 to v0 lane 1 is 256*4 = 1024B. Field
	// ByteSizePerLane should be set to 1024 in vector registers.
	ByteSizePerLane int
}

// NewSimpleRegisterFile creates and returns a new SimpleRegisterFile
func NewSimpleRegisterFile(
	byteSize uint64,
	byteSizePerLane int,
) *SimpleRegisterFile {
	r := new(SimpleRegisterFile)
	r.storage = make([]byte, byteSize)
	r.ByteSizePerLane = byteSizePerLane
	return r
}

func (r *SimpleRegisterFile) Write(access RegisterAccess) {
	accessType := "write"
	if access.Reg.IsSReg() {
		accessType += "S"
	} else {
		accessType += "V"
	}

	offset := r.getRegOffset(access.Reg, access.WaveOffset, access.LaneID)

	if access.Inst != nil {
		access.Cu.logRBankTask(access.Inst, access.Reg, accessType, false)

		if access.Reg.IsVReg() {
			access.Cu.logRBankTaskPlus(access.Inst, access.Reg, access.SIMDID*DEFAULT_VREG_SIZE+(offset/4), accessType, false)
		}
	}

	if access.RegCount == 0 {
		access.RegCount = 1
	}

	size := access.RegCount * 4
	copy(r.storage[offset:offset+size], access.Data[0:access.RegCount*4])
	access.OK = true

	if access.Inst != nil {
		access.Cu.logRBankTask(access.Inst, access.Reg, accessType, true)

		if access.Reg.IsVReg() {
			access.Cu.logRBankTaskPlus(access.Inst, access.Reg, access.SIMDID*DEFAULT_VREG_SIZE+(offset/4), accessType, true)
		}
	}
}

func (r *SimpleRegisterFile) Read(access RegisterAccess) {
	accessType := "read"
	if access.Reg.IsSReg() {
		accessType += "S"
	} else {
		accessType += "V"
	}

	offset := r.getRegOffset(access.Reg, access.WaveOffset, access.LaneID)

	if access.Inst != nil {
		access.Cu.logRBankTask(access.Inst, access.Reg, accessType, false)

		if access.Reg.IsVReg() {
			access.Cu.logRBankTaskPlus(access.Inst, access.Reg, access.SIMDID*DEFAULT_VREG_SIZE+(offset/4), accessType, false)
		}
	}

	if access.RegCount == 0 {
		access.RegCount = 1
	}

	size := access.RegCount * 4
	copy(access.Data, r.storage[offset:offset+size])
	access.OK = true
	if access.Inst != nil {
		access.Cu.logRBankTask(access.Inst, access.Reg, accessType, true)

		if access.Reg.IsVReg() {
			access.Cu.logRBankTaskPlus(access.Inst, access.Reg, access.SIMDID*DEFAULT_VREG_SIZE+(offset/4), accessType, true)
		}
	}
}

func (r *SimpleRegisterFile) getRegOffset(reg *insts.Reg, offset int, laneID int) int {
	if reg.IsSReg() {
		return reg.RegIndex()*4 + offset
	}

	if reg.IsVReg() {
		regOffset := reg.RegIndex()*4 + laneID*r.ByteSizePerLane + offset
		return regOffset
	}

	log.Panic("Register type not supported by register files")

	return 0
}
