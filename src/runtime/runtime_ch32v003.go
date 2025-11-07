//go:build ch32v003

package runtime

import (
	"device/riscv"
	"device/wch"
	"machine"
)

const sysTickInterruptCode = 12

//export main
func main() {

	preinit()

	machine.PD0.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.PD4.Configure(machine.PinConfig{Mode: machine.PinOutput})

	machine.PD0.Set(true)

	initSysTick()

	riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)
	for {
		riscv.Asm("wfi")
	}

	// //machine.InitSerial()
	// run()
	// exit(0)
}

// System tick timer initialization.
// Sets up SysTick to tick every 100 millisecond.
func initSysTick() {

	// CH32V003 runs at 24 MHz with AHB prescaler of 3.
	// The datasheet says the AHB prescaler is 1, but that is incorrect.
	ticksPerMillisecond := 8e6 / 1000
	if ticksPerMillisecond == 0 {
		ticksPerMillisecond = 1
	}
	wch.PFIC.SetSTK_CMPLR(uint32(ticksPerMillisecond - 1))

	systickMillis = 0

	wch.PFIC.STK_CTLR.Set(wch.PFIC_STK_CTLR_STE | wch.PFIC_STK_CTLR_STIE | wch.PFIC_STK_CTLR_STCLK | wch.PFIC_STK_CTLR_STRE)

	wch.PFIC.SetIENR1_INTEN12(1)
}

//go:extern handleInterruptASM
var handleInterruptASM [0]uintptr

//export handleInterrupt
func handleInterrupt() {

	machine.PD4.Set(false)
	machine.PD4.Set(true)

	cause := riscv.MCAUSE.Get()
	if cause&(1<<31) != 0 {
		switch cause & 0xff {
		case riscv.MachineTimerInterrupt, sysTickInterruptCode:
			sysTickInterrupt()
		default:
			// Unhandled interrupt.
		}
	} else {
		handleException(uint32(cause))
	}
	// Clear MCAUSE so interrupt.In() reports the correct state.
	riscv.MCAUSE.Set(0)
}

func handleException(code uint32) {

	//machine.PD4.Set(true)

	print("fatal error: exception with mcause=")
	print(code)
	print(" mepc=")
	print(riscv.MEPC.Get())
	println()
	abort()
}

func sysTickInterrupt() {
	wch.PFIC.SetSTK_SR_CNTIF(0)
	systickMillis++
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

const nanosPerMillisecond = int64(1_000_000)

var systickMillis uint64

func ticks() timeUnit {
	// mask := riscv.DisableInterrupts()
	current := systickMillis
	// riscv.EnableInterrupts(mask)
	return timeUnit(current)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * nanosPerMillisecond
}

func nanosecondsToTicks(ns int64) timeUnit {
	if ns <= 0 {
		return 0
	}
	return timeUnit((ns + nanosPerMillisecond - 1) / nanosPerMillisecond)
}

func sleepTicks(d timeUnit) {
	// target := ticks() + d
	// for ticks() < target {
	// 	//riscv.Asm("wfi")
	// }
}

func exit(code int) {
	abort()
}
