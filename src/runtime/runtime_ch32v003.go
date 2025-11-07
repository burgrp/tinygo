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
	// Route traps to the TinyGo interrupt handler.
	//riscv.MTVEC.Set(uintptr(unsafe.Pointer(&handleInterruptASM)))
	//riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)

	machine.PD0.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.PD4.Configure(machine.PinConfig{Mode: machine.PinOutput})

	machine.PD0.Set(true)

	for {
		riscv.Asm("ecall")
	}

	// preinit()
	// initSysTick()
	// //machine.InitSerial()
	// run()
	// exit(0)
}

func initSysTick() {

	wch.PFIC.STK_CTLR.Set(0)
	wch.PFIC.STK_SR.Set(0)
	wch.PFIC.SetSTK_CNTL(0)

	// ticksPerMillisecond := machine.CPUFrequency / 1000
	// if ticksPerMillisecond == 0 {
	// 	ticksPerMillisecond = 1
	// }
	ticksPerMillisecond := 100
	wch.PFIC.SetSTK_CMPLR(uint32(ticksPerMillisecond - 1))

	wch.PFIC.SetSTK_CTLR_STRE(1)
	wch.PFIC.SetSTK_CTLR_STCLK(1)
	wch.PFIC.SetSTK_CTLR_STIE(1)
	wch.PFIC.SetSTK_CTLR_INIT(1)

	wch.PFIC.SetIENR1_INTEN12(1)
	wch.PFIC.IENR1.Set(0xFFFFFFFF)
	wch.PFIC.IENR2.Set(0xFFFFFFFF)

	wch.PFIC.SetSTK_CTLR_STE(1)

	systickMillis = 0

	//riscv.Asm("ecall")
	wch.PFIC.SetSTK_CTLR_SWIE(1)

}

//go:extern handleInterruptASM
var handleInterruptASM [0]uintptr

//export handleInterrupt
func handleInterrupt() {

	machine.PD4.Set(true)
	machine.PD4.Set(false)

	// for {
	// }

	// cause := riscv.MCAUSE.Get()
	// if cause&(1<<31) != 0 {
	// 	switch cause & 0xff {
	// 	case riscv.MachineTimerInterrupt, sysTickInterruptCode:
	// 		sysTickInterrupt()
	// 	default:
	// 		// Unhandled interrupt.
	// 	}
	// } else {
	// 	handleException(uint32(cause))
	// }
	// // Clear MCAUSE so interrupt.In() reports the correct state.
	// riscv.MCAUSE.Set(0)
}

func handleException(code uint32) {

	machine.PD4.Set(true)

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
