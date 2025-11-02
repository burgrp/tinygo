//go:build ch32v003

package runtime

import (
	"device/riscv"
	"machine"
	"unsafe"
)

//export main
func main() {
	// Route traps to the TinyGo interrupt handler.
	riscv.MTVEC.Set(uintptr(unsafe.Pointer(&handleInterruptASM)))

	preinit()
	initSystem()
	run()
	exit(0)
}

func initSystem() {
	machine.InitSerial()
}

func putchar(c byte) {
	_ = machine.Serial.WriteByte(c)
}

func getchar() byte {
	for machine.Serial.Buffered() == 0 {
		riscv.Asm("wfi")
	}
	v, _ := machine.Serial.ReadByte()
	return v
}

func buffered() int {
	return machine.Serial.Buffered()
}

func abort() {
	for {
		riscv.Asm("wfi")
	}
}

var t timeUnit

func ticks() timeUnit {
	t = t + 10

	return t
	// for {
	// 	hi := uint32(riscv.CYCLEH.Get())
	// 	lo := uint32(riscv.CYCLE.Get())
	// 	hi2 := uint32(riscv.CYCLEH.Get())
	// 	if hi == hi2 {
	// 		return timeUnit(uint64(hi)<<32 | uint64(lo))
	// 	}
	// }
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 125 / 3
	//return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	//return timeUnit(ns)
	return timeUnit(ns * 3 / 125)
}

func sleepTicks(d timeUnit) {
	target := ticks() + d
	for ticks() < target {
		//riscv.Asm("wfi")
	}
}

func exit(code int) {
	abort()
}

//go:extern handleInterruptASM
var handleInterruptASM [0]uintptr

//export handleInterrupt
func handleInterrupt() {
	cause := riscv.MCAUSE.Get()
	if cause&(1<<31) != 0 {
		// Interrupt handling is not yet implemented for this device.
	} else {
		handleException(uint32(cause))
	}
	// Clear MCAUSE so interrupt.In() reports the correct state.
	riscv.MCAUSE.Set(0)
}

func handleException(code uint32) {
	print("fatal error: exception with mcause=")
	print(code)
	print(" mepc=")
	print(riscv.MEPC.Get())
	println()
	abort()
}
