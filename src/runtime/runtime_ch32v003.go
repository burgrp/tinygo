//go:build ch32v003

package runtime

import (
	"device/riscv"
	"machine"
	"unsafe"
)

//export main
func main() {
	// Set the interrupt vector base. Required even if we do not enable interrupts.
	riscv.MTVEC.Set(uintptr(unsafe.Pointer(&handleInterruptASM)))

	preinit()
	initPeripherals()
	run()
}

//go:extern handleInterruptASM
var handleInterruptASM [0]uintptr

//export handleInterrupt
func handleInterrupt() {
	cause := riscv.MCAUSE.Get()
	code := uint(cause &^ (1 << 31))
	if cause&(1<<31) != 0 {
		// Interrupt: none are expected, so just clear the cause.
		riscv.MCAUSE.Set(0)
		return
	}

	handleException(code)
	riscv.MCAUSE.Set(0)
}

func initPeripherals() {
	machine.InitSerial()
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

func getchar() byte {
	for machine.Serial.Buffered() == 0 {
	}
	v, _ := machine.Serial.ReadByte()
	return v
}

func buffered() int {
	return machine.Serial.Buffered()
}

func ticks() timeUnit {
	return timeUnit(readCycle())
}

func sleepTicks(d timeUnit) {
	target := ticks() + d
	for ticks() < target {
		riscv.Asm("nop")
	}
}

func readCycle() uint64 {
	for {
		hi1 := riscv.AsmFull("csrr {}, mcycleh", nil)
		lo := riscv.AsmFull("csrr {}, mcycle", nil)
		hi2 := riscv.AsmFull("csrr {}, mcycleh", nil)
		if hi1 == hi2 {
			return (uint64(hi1) << 32) | uint64(lo)
		}
	}
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	t := int64(ticks)
	quot := t / 6
	rem := t % 6
	return quot*125 + rem*125/6
}

func nanosecondsToTicks(ns int64) timeUnit {
	quot := ns / 125
	rem := ns % 125
	return timeUnit(quot*6 + rem*6/125)
}

func exit(code int) {
	abort()
}

func abort() {
	for {
		riscv.Asm("wfi")
	}
}

// handleException converts the machine exception code into a panic.
func handleException(code uint) {
	switch code {
	case 0:
		runtimePanic("instruction address misaligned")
	case 1:
		runtimePanic("instruction access fault")
	case 2:
		runtimePanic("illegal instruction")
	case 3:
		runtimePanic("breakpoint")
	case 4:
		runtimePanic("load address misaligned")
	case 5:
		runtimePanic("load access fault")
	case 6:
		runtimePanic("store address misaligned")
	case 7:
		runtimePanic("store access fault")
	case 8:
		runtimePanic("environment call from U-mode")
	case 9:
		runtimePanic("environment call from S-mode")
	case 11:
		runtimePanic("environment call from M-mode")
	default:
		runtimePanic("unknown exception")
	}
}
